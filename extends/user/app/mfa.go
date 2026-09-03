package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"kingfisher/core/cache"
	"kingfisher/extends/user/domain"
	"kingfisher/extends/user/port"
)

type SMSSender func(ctx context.Context, phone, code string) error

type MFAService struct {
	repo      port.UserRepository
	cache     cache.Cache
	getConfig func(ctx context.Context, key string) (string, error)
	sendEmail func(ctx context.Context, to, subject, body string) error
	sendSMS   SMSSender
	mu        sync.Mutex
	mem       map[string]memEntry
}

type memEntry struct {
	val string
	exp time.Time
}

func NewMFAService(repo port.UserRepository, c cache.Cache) *MFAService {
	return &MFAService{repo: repo, cache: c, sendSMS: defaultSMSSender, mem: make(map[string]memEntry)}
}

func (s *MFAService) memGet(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.mem[key]
	if !ok {
		return ""
	}
	if time.Now().After(e.exp) {
		delete(s.mem, key)
		return ""
	}
	return e.val
}

func (s *MFAService) memSet(key, val string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mem[key] = memEntry{val: val, exp: time.Now().Add(ttl)}
}

func (s *MFAService) memDel(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.mem, key)
}

func (s *MFAService) cacheGet(ctx context.Context, key string) string {
	if s.cache != nil {
		if v, _ := s.cache.Get(ctx, key); v != "" {
			return v
		}
	}
	return s.memGet(key)
}

func (s *MFAService) cacheSet(ctx context.Context, key, val string, ttl time.Duration) {
	if s.cache != nil {
		_ = s.cache.Set(ctx, key, val, ttl)
		return
	}
	s.memSet(key, val, ttl)
}

func (s *MFAService) cacheDel(ctx context.Context, key string) {
	if s.cache != nil {
		_ = s.cache.Delete(ctx, key)
	}
	s.memDel(key)
}

func defaultSMSSender(ctx context.Context, phone, code string) error {
	return nil
}

func (s *MFAService) SetConfigProvider(fn func(ctx context.Context, key string) (string, error)) {
	s.getConfig = fn
}

func (s *MFAService) SetEmailSender(fn func(ctx context.Context, to, subject, body string) error) {
	s.sendEmail = fn
}

func (s *MFAService) SetSMSSender(fn SMSSender) {
	s.sendSMS = fn
}

func (s *MFAService) GetStatus(ctx context.Context, userID uint) (*domain.MFAStatus, error) {
	return s.repo.GetMFAStatus(ctx, userID)
}

func (s *MFAService) isEnforceAll(ctx context.Context) bool {
	if s.getConfig == nil {
		return false
	}
	v, _ := s.getConfig(ctx, "mfa_enforce")
	return v == "all"
}

func (s *MFAService) isEnforceAdmin(ctx context.Context, roleCodes []string) bool {
	if s.getConfig == nil {
		return false
	}
	v, _ := s.getConfig(ctx, "mfa_enforce")
	if v != "admin" {
		return false
	}
	for _, c := range roleCodes {
		if c == "admin" {
			return true
		}
	}
	return false
}

func (s *MFAService) IsMFARequired(ctx context.Context, user *domain.User) bool {
	status, err := s.repo.GetMFAStatus(ctx, user.ID)
	if err == nil && status.Enabled {
		return true
	}
	if s.isEnforceAll(ctx) {
		return true
	}
	var codes []string
	for _, r := range user.Roles {
		codes = append(codes, r.Code)
	}
	return s.isEnforceAdmin(ctx, codes)
}

func (s *MFAService) SetupTOTP(ctx context.Context, userID uint) (*domain.MFASecretInfo, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "Kingfisher", AccountName: user.Username})
	if err != nil {
		return nil, err
	}
	secret := key.Secret()
	if err := s.repo.SetMFASecret(ctx, userID, secret); err != nil {
		return nil, err
	}
	return &domain.MFASecretInfo{Secret: secret, URL: key.URL()}, nil
}

func (s *MFAService) VerifyTOTPSetup(ctx context.Context, userID uint, code string) ([]string, error) {
	secret, err := s.repo.GetMFASecret(ctx, userID)
	if err != nil || secret == "" {
		return nil, fmt.Errorf("no totp secret")
	}
	if !totp.Validate(code, secret) {
		if s.tryBackupCode(ctx, userID, code) {
			return nil, fmt.Errorf("use backup code not allowed for setup")
		}
		return nil, fmt.Errorf("invalid code")
	}
	if err := s.repo.Update(ctx, userID, map[string]any{"mfa_totp_enabled": true}); err != nil {
		return nil, err
	}
	codes, err := s.generateBackupCodes(ctx, userID)
	if err != nil {
		return nil, err
	}
	return codes, nil
}

func (s *MFAService) DisableTOTP(ctx context.Context, userID uint, code string) error {
	status, err := s.repo.GetMFAStatus(ctx, userID)
	if err != nil {
		return err
	}
	if !status.TOTPEnabled {
		return fmt.Errorf("not enabled")
	}
	secret, _ := s.repo.GetMFASecret(ctx, userID)
	if secret != "" && totp.Validate(code, secret) {
		return s.repo.Update(ctx, userID, map[string]any{"mfa_totp_enabled": false})
	}
	if s.tryBackupCode(ctx, userID, code) {
		return s.repo.Update(ctx, userID, map[string]any{"mfa_totp_enabled": false})
	}
	return fmt.Errorf("invalid code")
}

func (s *MFAService) VerifyTOTP(ctx context.Context, userID uint, code string) bool {
	secret, err := s.repo.GetMFASecret(ctx, userID)
	if err != nil || secret == "" {
		return false
	}
	return totp.Validate(code, secret)
}

func (s *MFAService) SendSMSCode(ctx context.Context, userID uint) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	phone := user.Phone
	if phone == "" {
		st, _ := s.repo.GetMFAStatus(ctx, userID)
		if st != nil {
			phone = st.Phone
		}
	}
	if phone == "" {
		return fmt.Errorf("no phone")
	}
	key := fmt.Sprintf("mfa:sms:%d", userID)
	if s.cacheGet(ctx, key) != "" {
		return fmt.Errorf("too frequent")
	}
	code := randomDigits(6)
	s.cacheSet(ctx, key, code, 5*time.Minute)
	if s.sendSMS != nil {
		_ = s.sendSMS(ctx, phone, code)
	}
	return nil
}

func (s *MFAService) SendEmailCode(ctx context.Context, userID uint) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	email := user.Email
	if email == "" {
		return fmt.Errorf("no email")
	}
	key := fmt.Sprintf("mfa:email:%d", userID)
	if s.cacheGet(ctx, key) != "" {
		return fmt.Errorf("too frequent")
	}
	code := randomDigits(6)
	s.cacheSet(ctx, key, code, 5*time.Minute)
	if s.sendEmail != nil {
		subject := "Kingfisher 登录验证码"
		body := fmt.Sprintf("您的登录验证码是 %s，5 分钟内有效。", code)
		_ = s.sendEmail(ctx, email, subject, body)
	}
	return nil
}

func (s *MFAService) VerifySMSCode(ctx context.Context, userID uint, code string) bool {
	if s.tryBackupCode(ctx, userID, code) {
		return true
	}
	key := fmt.Sprintf("mfa:sms:%d", userID)
	v := s.cacheGet(ctx, key)
	if v == "" || v != code {
		return false
	}
	s.cacheDel(ctx, key)
	return true
}

func (s *MFAService) VerifyEmailCode(ctx context.Context, userID uint, code string) bool {
	if s.tryBackupCode(ctx, userID, code) {
		return true
	}
	key := fmt.Sprintf("mfa:email:%d", userID)
	v := s.cacheGet(ctx, key)
	if v == "" || v != code {
		return false
	}
	s.cacheDel(ctx, key)
	return true
}

func (s *MFAService) EnableSMS(ctx context.Context, userID uint, phone, code string) error {
	key := fmt.Sprintf("mfa:sms:%d", userID)
	expected := s.cacheGet(ctx, key)
	if expected == "" || expected != code {
		return fmt.Errorf("invalid code")
	}
	s.cacheDel(ctx, key)
	return s.repo.Update(ctx, userID, map[string]any{"phone": phone, "mfa_sms_enabled": true})
}

func (s *MFAService) EnableEmail(ctx context.Context, userID uint, code string) error {
	key := fmt.Sprintf("mfa:email:%d", userID)
	expected := s.cacheGet(ctx, key)
	if expected == "" || expected != code {
		return fmt.Errorf("invalid code")
	}
	s.cacheDel(ctx, key)
	return s.repo.Update(ctx, userID, map[string]any{"mfa_email_enabled": true})
}

func (s *MFAService) DisableSMS(ctx context.Context, userID uint) error {
	return s.repo.Update(ctx, userID, map[string]any{"mfa_sms_enabled": false})
}

func (s *MFAService) DisableEmail(ctx context.Context, userID uint) error {
	return s.repo.Update(ctx, userID, map[string]any{"mfa_email_enabled": false})
}

func (s *MFAService) VerifyLogin(ctx context.Context, userID uint, method, code string) bool {
	switch method {
	case "totp":
		if s.VerifyTOTP(ctx, userID, code) {
			return true
		}
	case "sms":
		if s.VerifySMSCode(ctx, userID, code) {
			return true
		}
	case "email":
		if s.VerifyEmailCode(ctx, userID, code) {
			return true
		}
	case "backup":
		return s.tryBackupCode(ctx, userID, code)
	default:
		if s.VerifyTOTP(ctx, userID, code) {
			return true
		}
		if s.VerifySMSCode(ctx, userID, code) {
			return true
		}
		if s.VerifyEmailCode(ctx, userID, code) {
			return true
		}
	}
	return s.tryBackupCode(ctx, userID, code)
}

func (s *MFAService) GenerateMFAToken(ctx context.Context, userID uint) (string, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}
	key := "mfa:token:" + token
	s.cacheSet(ctx, key, fmt.Sprint(userID), 5*time.Minute)
	return token, nil
}

func (s *MFAService) ResolveMFAToken(ctx context.Context, token string) (uint, error) {
	key := "mfa:token:" + token
	v := s.cacheGet(ctx, key)
	if v == "" {
		return 0, fmt.Errorf("invalid or expired mfa token")
	}
	var id uint64
	_, err := fmt.Sscanf(v, "%d", &id)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (s *MFAService) ConsumeMFAToken(ctx context.Context, token string) {
	s.cacheDel(ctx, "mfa:token:"+token)
}

func (s *MFAService) ResetMFA(ctx context.Context, userID uint) error {
	return s.repo.Update(ctx, userID, map[string]any{
		"mfa_totp_secret": "", "mfa_totp_enabled": false, "mfa_sms_enabled": false, "mfa_email_enabled": false, "mfa_backup_codes": "",
	})
}

func (s *MFAService) generateBackupCodes(ctx context.Context, userID uint) ([]string, error) {
	codes := make([]string, 8)
	hashed := make([]string, 8)
	for i := 0; i < 8; i++ {
		c := randomDigits(8)
		codes[i] = c
		h, err := bcrypt.GenerateFromPassword([]byte(c), 10)
		if err != nil {
			return nil, err
		}
		hashed[i] = string(h)
	}
	joined := strings.Join(hashed, ",")
	if err := s.repo.SetBackupCodes(ctx, userID, joined); err != nil {
		b, _ := json.Marshal(hashed)
		_ = s.repo.Update(ctx, userID, map[string]any{"mfa_backup_codes": string(b)})
		_ = s.repo.SetBackupCodes(ctx, userID, joined)
	}
	return codes, nil
}

func (s *MFAService) tryBackupCode(ctx context.Context, userID uint, code string) bool {
	raw, err := s.repo.GetBackupCodes(ctx, userID)
	if err != nil || raw == "" {
		return false
	}
	var list []string
	if strings.HasPrefix(strings.TrimSpace(raw), "[") {
		_ = json.Unmarshal([]byte(raw), &list)
	} else {
		list = strings.Split(raw, ",")
	}
	remaining := make([]string, 0, len(list))
	found := false
	for _, h := range list {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if !found && bcrypt.CompareHashAndPassword([]byte(h), []byte(code)) == nil {
			found = true
			continue
		}
		remaining = append(remaining, h)
	}
	if found {
		_ = s.repo.SetBackupCodes(ctx, userID, strings.Join(remaining, ","))
	}
	return found
}

func randomDigits(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		b[i] = byte('0' + num.Int64())
	}
	return string(b)
}
