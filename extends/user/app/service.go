package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"kingfisher/core/cache"
	"kingfisher/core/jwt"
	"kingfisher/core/query"
	"kingfisher/extends/user/domain"
	"kingfisher/extends/user/port"
)

var ErrMFARequired = errors.New("mfa_required")
var ErrMFASetupRequired = errors.New("mfa_setup_required")

type AuthService struct {
	repo           port.UserRepository
	cache          cache.Cache
	jwtMgr         *jwt.JWTManager
	getLandingPage func(ctx context.Context, roleID uint) (string, error)
	getConfig      func(ctx context.Context, key string) (string, error)
	sendEmail      func(ctx context.Context, to, subject, body string) error
	renderTemplate func(ctx context.Context, code string, vars map[string]string) (subject, body string, err error)
	resetTokenTTL  time.Duration
	mfaSvc         *MFAService
}

type MFARequiredError struct {
	Token         string
	Methods       []string
	SetupRequired bool
}

func (e *MFARequiredError) Error() string {
	if e.SetupRequired {
		return "mfa_setup_required"
	}
	return "mfa_required"
}

func NewAuthService(repo port.UserRepository, c cache.Cache, j *jwt.JWTManager) *AuthService {
	return &AuthService{repo: repo, cache: c, jwtMgr: j, resetTokenTTL: 30 * time.Minute}
}

// SetEmailSender 注入邮件发送函数（异步入队）。
func (s *AuthService) SetEmailSender(fn func(ctx context.Context, to, subject, body string) error) {
	s.sendEmail = fn
}

// SetTemplateRenderer 注入模板渲染函数。
func (s *AuthService) SetTemplateRenderer(fn func(ctx context.Context, code string, vars map[string]string) (subject, body string, err error)) {
	s.renderTemplate = fn
}

// ForgotPassword 找回密码：按邮箱查找用户，生成一次性 token 存 Redis，异步发重置邮件。
// 出于防枚举考虑：无论邮箱是否存在都返回成功（避免暴露用户是否注册）。
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	// 找回密码开关：password_reset_enabled=false 时禁用（防枚举：返回相同错误）
	if s.getConfig != nil {
		if v, e := s.getConfig(ctx, "password_reset_enabled"); e == nil && v != "" && v != "true" {
			return fmt.Errorf("password reset disabled")
		}
	}
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		// 数据库查询失败不能吞掉：它不等于"用户不存在"，应返回真实错误
		return fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		// 用户不存在也返回 nil，防枚举；但仍尝试记录（无收件人可发则不发送）
		return nil
	}
	// 生成一次性 token
	token, err := randomToken(32)
	if err != nil {
		return fmt.Errorf("gen token: %w", err)
	}
	// 存 Redis：reset:token:<token> = userID，30 分钟过期
	if s.cache != nil {
		key := "reset:token:" + token
		if err := s.cache.Set(ctx, key, fmt.Sprint(user.ID), s.resetTokenTTL); err != nil {
			return fmt.Errorf("store reset token: %w", err)
		}
	}
	// 渲染模板并异步发邮件（未注入 email sender 则跳过）
	if s.sendEmail != nil && s.renderTemplate != nil {
		resetURL := resetLinkBase(ctx, s.getConfig) + "/reset?token=" + token
		vars := map[string]string{"nickname": user.Nickname, "reset_url": resetURL, "token": token}
		subject, body, err := s.renderTemplate(ctx, "password_reset", vars)
		if err == nil {
			_ = s.sendEmail(ctx, user.Email, subject, body)
		}
	}
	return nil
}

// ResetPassword 重置密码：验证一次性 token，更新密码并使旧 session 失效。
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 || len(newPassword) > 64 {
		return fmt.Errorf("password length invalid")
	}
	if s.cache == nil {
		return fmt.Errorf("reset token unavailable")
	}
	key := "reset:token:" + token
	uidStr, err := s.cache.Get(ctx, key)
	if err != nil || uidStr == "" {
		return fmt.Errorf("invalid or expired reset token")
	}
	var userID uint64
	if _, err := fmt.Sscanf(uidStr, "%d", &userID); err != nil || userID == 0 {
		return fmt.Errorf("invalid reset token")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.repo.Update(ctx, uint(userID), map[string]any{"password": string(hashed)}); err != nil {
		return err
	}
	// 使旧 session 失效
	_ = s.repo.IncrementSessionVersion(ctx, uint(userID))
	// 删除已用 token
	_ = s.cache.Delete(ctx, key)
	return nil
}

// randomToken 生成 n 字节随机 token（十六进制）。
func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// resetLinkBase 返回前端地址（前端跳转重置页用），从系统配置 site_url 读取，否则默认 localhost:8080。
func resetLinkBase(ctx context.Context, getConfig func(ctx context.Context, key string) (string, error)) string {
	if getConfig != nil {
		if v, e := getConfig(ctx, "site_url"); e == nil && v != "" {
			return strings.TrimRight(v, "/")
		}
	}
	return "http://localhost:8080"
}

// SetLandingPageProvider 注入角色落地页查询函数。
func (s *AuthService) SetLandingPageProvider(fn func(ctx context.Context, roleID uint) (string, error)) {
	s.getLandingPage = fn
}

func (s *AuthService) SetConfigProvider(fn func(ctx context.Context, key string) (string, error)) {
	s.getConfig = fn
}

func (s *AuthService) SetMFAService(m *MFAService) {
	s.mfaSvc = m
	if m != nil {
		m.SetConfigProvider(s.getConfig)
		m.SetEmailSender(s.sendEmail)
	}
}

func (s *AuthService) Register(ctx context.Context, username, password, email string) (*domain.User, error) {
	// 注册开关：registration_enabled=false 时拒绝注册
	if s.getConfig != nil {
		if v, e := s.getConfig(ctx, "registration_enabled"); e == nil && v != "" && v != "true" {
			return nil, fmt.Errorf("registration disabled")
		}
	}
	// Check for existing user
	_, err := s.repo.FindByUsername(ctx, username)
	if err == nil {
		return nil, fmt.Errorf("user exists")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// 默认注册角色：从配置读取（default_register_role_id），否则用访客(4)
	roleID := uint(4)
	if s.getConfig != nil {
		if v, e := s.getConfig(ctx, "default_register_role_id"); e == nil && v != "" {
			if n, err := strconv.ParseUint(v, 10, 64); err == nil && n > 0 {
				roleID = uint(n)
			}
		}
	}
	user := &domain.User{Username: username, Password: string(hashed), Email: email, Status: 1, RoleIDs: []uint{roleID}}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	user.Password = "" // never return password
	return user, nil
}

var dummyHash = "$2a$12$LJ3m4ys3Lk0TSwHCpNqrIeN5U5Akn5dQUhBvPXFxFG7GqQvHCzB5q"

func (s *AuthService) Login(ctx context.Context, username, password string) (string, string, *domain.User, string, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	hashToCheck := dummyHash
	if err == nil {
		hashToCheck = user.Password
	}
	if bcrypt.CompareHashAndPassword([]byte(hashToCheck), []byte(password)) != nil {
		// Check rate limit
		if s.cache != nil {
			count, _ := s.cache.Incr(ctx, "login_fail:"+username)
			if count == 1 {
				_ = s.cache.Expire(ctx, "login_fail:"+username, 15*time.Minute)
			}
			if count > 5 {
				return "", "", nil, "", fmt.Errorf("too many attempts")
			}
		}
		return "", "", nil, "", fmt.Errorf("wrong password")
	}
	if err != nil {
		return "", "", nil, "", fmt.Errorf("wrong password")
	}
	if user.Status != 1 {
		return "", "", nil, "", fmt.Errorf("user disabled")
	}

	if s.cache != nil {
		_ = s.cache.Delete(ctx, "login_fail:"+username)
	}
	if s.mfaSvc != nil && s.mfaSvc.IsMFARequired(ctx, user) {
		status, _ := s.repo.GetMFAStatus(ctx, user.ID)
		var methods []string
		if status != nil {
			methods = status.Methods
		}
		if len(methods) == 0 {
			return "", "", nil, "", &MFARequiredError{Token: "", Methods: nil, SetupRequired: true}
		}
		token, err := s.mfaSvc.GenerateMFAToken(ctx, user.ID)
		if err != nil {
			return "", "", nil, "", err
		}
		return "", "", nil, "", &MFARequiredError{Token: token, Methods: methods}
	}
	var roleIDs []uint
	var roleCodes []string
	for _, r := range user.Roles {
		roleIDs = append(roleIDs, r.ID)
		roleCodes = append(roleCodes, r.Code)
	}
	if len(roleIDs) == 0 {
		return "", "", nil, "", fmt.Errorf("user has no roles")
	}
	access, refresh, err := s.jwtMgr.GenerateToken(ctx, user.ID, roleIDs, roleCodes, user.Username, user.SessionVersion)
	if err != nil {
		return "", "", nil, "", err
	}
	landing := ""
	if s.getLandingPage != nil {
		if lp, e := s.getLandingPage(ctx, roleIDs[0]); e == nil {
			landing = lp
		}
	}
	user.Password = ""
	return access, refresh, user, landing, nil
}

func (s *AuthService) VerifyMFA(ctx context.Context, mfaToken, method, code string) (string, string, *domain.User, string, error) {
	if s.mfaSvc == nil {
		return "", "", nil, "", fmt.Errorf("mfa not enabled")
	}
	uid, err := s.mfaSvc.ResolveMFAToken(ctx, mfaToken)
	if err != nil {
		return "", "", nil, "", fmt.Errorf("invalid mfa token")
	}
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return "", "", nil, "", err
	}
	if user.Status != 1 {
		return "", "", nil, "", fmt.Errorf("user disabled")
	}
	if !s.mfaSvc.VerifyLogin(ctx, uid, method, code) {
		return "", "", nil, "", fmt.Errorf("invalid mfa code")
	}
	s.mfaSvc.ConsumeMFAToken(ctx, mfaToken)
	var roleIDs []uint
	var roleCodes []string
	for _, r := range user.Roles {
		roleIDs = append(roleIDs, r.ID)
		roleCodes = append(roleCodes, r.Code)
	}
	if len(roleIDs) == 0 {
		return "", "", nil, "", fmt.Errorf("user has no roles")
	}
	access, refresh, err := s.jwtMgr.GenerateToken(ctx, user.ID, roleIDs, roleCodes, user.Username, user.SessionVersion)
	if err != nil {
		return "", "", nil, "", err
	}
	landing := ""
	if s.getLandingPage != nil {
		if lp, e := s.getLandingPage(ctx, roleIDs[0]); e == nil {
			landing = lp
		}
	}
	user.Password = ""
	if s.cache != nil {
		_ = s.cache.Delete(ctx, "login_fail:"+user.Username)
	}
	return access, refresh, user, landing, nil
}

func (s *AuthService) SendMFACode(ctx context.Context, mfaToken, method string) error {
	if s.mfaSvc == nil {
		return fmt.Errorf("mfa not enabled")
	}
	uid, err := s.mfaSvc.ResolveMFAToken(ctx, mfaToken)
	if err != nil {
		return err
	}
	switch method {
	case "sms":
		return s.mfaSvc.SendSMSCode(ctx, uid)
	case "email":
		return s.mfaSvc.SendEmailCode(ctx, uid)
	default:
		return fmt.Errorf("unsupported method")
	}
}

func (s *AuthService) RevokeToken(ctx context.Context, tokenStr string) error {
	return s.jwtMgr.RevokeToken(ctx, tokenStr)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return s.jwtMgr.RefreshToken(ctx, refreshToken)
}

type UserService struct {
	repo         port.UserRepository
	cache        cache.Cache
	getUserPerms func(ctx context.Context, userID uint) ([]string, error)
}

func NewUserService(repo port.UserRepository, c cache.Cache) *UserService {
	return &UserService{repo: repo, cache: c}
}

func (s *UserService) SetPermProvider(fn func(ctx context.Context, userID uint) ([]string, error)) {
	s.getUserPerms = fn
}

func (s *UserService) isPermSubset(ctx context.Context, parentID uint, roleIDs []uint) (bool, error) {
	if s.getUserPerms == nil {
		parent, err := s.repo.FindByID(ctx, parentID)
		if err != nil {
			return false, err
		}
		allowed := make(map[uint]bool, len(parent.RoleIDs))
		for _, rid := range parent.RoleIDs {
			allowed[rid] = true
		}
		for _, rid := range roleIDs {
			if !allowed[rid] {
				return false, nil
			}
		}
		return true, nil
	}
	parentPerms, err := s.getUserPerms(ctx, parentID)
	if err != nil {
		return false, err
	}
	allowed := make(map[string]bool, len(parentPerms))
	for _, p := range parentPerms {
		allowed[p] = true
	}
	reqPerms, err := s.repo.GetPermCodesByRoleIDs(ctx, roleIDs)
	if err != nil {
		return false, err
	}
	for _, p := range reqPerms {
		if !allowed[p] {
			return false, nil
		}
	}
	return true, nil
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id uint, updates map[string]any) error {
	_, hasRole := updates["role_ids"]
	_, hasDept := updates["dept_ids"]
	if hasRole {
		if target, err := s.repo.FindByID(ctx, id); err == nil && target.ParentID != nil {
			if reqRoles, ok := updates["role_ids"].([]uint); ok {
				if ok2, err := s.isPermSubset(ctx, *target.ParentID, reqRoles); err != nil {
					return err
				} else if !ok2 {
					return fmt.Errorf("role %d not in parent permissions", reqRoles[0])
				}
			}
		}
	}
	if err := s.repo.Update(ctx, id, updates); err != nil {
		return err
	}
	if hasRole && s.cache != nil {
		_ = s.cache.Delete(ctx, "user:perms:"+strconv.FormatUint(uint64(id), 10))
	}
	if hasDept && s.cache != nil {
		_ = s.cache.Delete(ctx, "user:perms:"+strconv.FormatUint(uint64(id), 10))
	}
	if hasRole || hasDept {
		_ = s.pruneSubAccounts(ctx, id)
	}
	return nil
}

func (s *UserService) pruneSubAccounts(ctx context.Context, parentID uint) error {
	parent, err := s.repo.FindByID(ctx, parentID)
	if err != nil {
		return err
	}
	if parent.ParentID != nil {
		return nil
	}
	subs, err := s.repo.FindSubAccounts(ctx, parentID)
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		return nil
	}
	var parentPerms map[string]bool
	if s.getUserPerms != nil {
		perms, _ := s.getUserPerms(ctx, parentID)
		parentPerms = make(map[string]bool, len(perms))
		for _, p := range perms {
			parentPerms[p] = true
		}
	} else {
		allowed := make(map[uint]bool, len(parent.RoleIDs))
		for _, rid := range parent.RoleIDs {
			allowed[rid] = true
		}
		parentPerms = nil
		for _, sub := range subs {
			directIDs, err := s.repo.FindDirectRoleIDs(ctx, sub.ID)
			if err != nil {
				continue
			}
			kept := make([]uint, 0, len(directIDs))
			changed := false
			for _, rid := range directIDs {
				if allowed[rid] {
					kept = append(kept, rid)
				} else {
					changed = true
				}
			}
			if !changed {
				continue
			}
			_ = s.repo.Update(ctx, sub.ID, map[string]any{"role_ids": kept})
			if s.cache != nil {
				_ = s.cache.Delete(ctx, "user:perms:"+strconv.FormatUint(uint64(sub.ID), 10))
			}
		}
		return nil
	}
	for _, sub := range subs {
		directIDs, err := s.repo.FindDirectRoleIDs(ctx, sub.ID)
		if err != nil {
			continue
		}
		kept := make([]uint, 0, len(directIDs))
		changed := false
		for _, rid := range directIDs {
			perms, _ := s.repo.GetPermCodesByRoleIDs(ctx, []uint{rid})
			subset := true
			for _, p := range perms {
				if !parentPerms[p] {
					subset = false
					break
				}
			}
			if subset {
				kept = append(kept, rid)
			} else {
				changed = true
			}
		}
		if !changed {
			continue
		}
		_ = s.repo.Update(ctx, sub.ID, map[string]any{"role_ids": kept})
		if s.cache != nil {
			_ = s.cache.Delete(ctx, "user:perms:"+strconv.FormatUint(uint64(sub.ID), 10))
		}
	}
	return nil
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	subs, _ := s.repo.FindSubAccounts(ctx, id)
	for _, sub := range subs {
		_ = s.repo.Delete(ctx, sub.ID)
		if s.cache != nil {
			_ = s.cache.Delete(ctx, "user:perms:"+strconv.FormatUint(uint64(sub.ID), 10))
		}
	}
	return nil
}

func (s *UserService) BatchDelete(ctx context.Context, ids []uint) error {
	if err := s.repo.DeleteBatch(ctx, ids); err != nil {
		return err
	}
	for _, pid := range ids {
		subs, _ := s.repo.FindSubAccounts(ctx, pid)
		for _, sub := range subs {
			_ = s.repo.Delete(ctx, sub.ID)
			if s.cache != nil {
				_ = s.cache.Delete(ctx, "user:perms:"+strconv.FormatUint(uint64(sub.ID), 10))
			}
		}
	}
	return nil
}

func (s *UserService) BatchUpdateStatus(ctx context.Context, ids []uint, status int) error {
	if err := s.repo.UpdateStatusBatch(ctx, ids, status); err != nil {
		return err
	}
	if status != 1 {
		for _, pid := range ids {
			subs, _ := s.repo.FindSubAccounts(ctx, pid)
			for _, sub := range subs {
				_ = s.repo.Update(ctx, sub.ID, map[string]any{"status": status})
				if s.cache != nil {
					_ = s.cache.Delete(ctx, "user:perms:"+strconv.FormatUint(uint64(sub.ID), 10))
				}
			}
		}
	}
	return nil
}

func (s *UserService) CreateUser(ctx context.Context, username, password, email string, roleIDs, deptIDs []uint) (*domain.User, error) {
	_, err := s.repo.FindByUsername(ctx, username)
	if err == nil {
		return nil, fmt.Errorf("user exists")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash: %w", err)
	}
	if len(roleIDs) == 0 && len(deptIDs) == 0 {
		roleIDs = []uint{4}
	}
	user := &domain.User{Username: username, Password: string(hashed), Email: email, Status: 1, RoleIDs: roleIDs, DeptIDs: deptIDs}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (s *UserService) CreateSubAccount(ctx context.Context, parentID uint, username, password, email string, roleIDs []uint) (*domain.User, error) {
	parent, err := s.repo.FindByID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("parent not found")
	}
	if parent.ParentID != nil {
		return nil, fmt.Errorf("sub account cannot create sub account")
	}
	if _, err := s.repo.FindByUsername(ctx, username); err == nil {
		return nil, fmt.Errorf("user exists")
	}
	if len(roleIDs) == 0 {
		return nil, fmt.Errorf("sub account must have at least one role")
	}
	if ok, err := s.isPermSubset(ctx, parentID, roleIDs); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("role %d not in parent permissions", roleIDs[0])
	}
	if cnt, _ := s.repo.CountSubAccounts(ctx, parentID); cnt >= 20 {
		return nil, fmt.Errorf("sub account limit reached")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash: %w", err)
	}
	pid := parentID
	user := &domain.User{Username: username, Password: string(hashed), Email: email, Status: 1, RoleIDs: roleIDs, ParentID: &pid}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (s *UserService) ListSubAccounts(ctx context.Context, parentID uint) ([]domain.User, error) {
	parent, err := s.repo.FindByID(ctx, parentID)
	if err != nil {
		return nil, err
	}
	if parent.ParentID != nil {
		return nil, fmt.Errorf("sub account has no sub accounts")
	}
	return s.repo.FindSubAccounts(ctx, parentID)
}

func (s *UserService) UpdateSubAccount(ctx context.Context, parentID, subID uint, roleIDs []uint) error {
	sub, err := s.repo.FindByID(ctx, subID)
	if err != nil {
		return err
	}
	if sub.ParentID == nil || *sub.ParentID != parentID {
		return fmt.Errorf("not your sub account")
	}
	if _, err := s.repo.FindByID(ctx, parentID); err != nil {
		return fmt.Errorf("parent not found")
	}
	if len(roleIDs) == 0 {
		return fmt.Errorf("sub account must have at least one role")
	}
	if ok, err := s.isPermSubset(ctx, parentID, roleIDs); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("role %d not in parent permissions", roleIDs[0])
	}
	if err := s.repo.Update(ctx, subID, map[string]any{"role_ids": roleIDs}); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, "user:perms:"+strconv.FormatUint(uint64(subID), 10))
	}
	return nil
}

func (s *UserService) DeleteSubAccount(ctx context.Context, parentID, subID uint) error {
	sub, err := s.repo.FindByID(ctx, subID)
	if err != nil {
		return err
	}
	if sub.ParentID == nil || *sub.ParentID != parentID {
		return fmt.Errorf("not your sub account")
	}
	return s.repo.Delete(ctx, subID)
}

func (s *UserService) AdminListSubAccounts(ctx context.Context, q *query.Query) ([]domain.User, int64, error) {
	q.Filters = append(q.Filters, query.Condition{Field: "is_sub_account", Op: "eq", Value: true})
	return s.repo.FindAll(ctx, q)
}

func (s *UserService) List(ctx context.Context, q *query.Query) ([]domain.User, int64, error) {
	return s.repo.FindAll(ctx, q)
}
func (s *UserService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	// Placeholder: returns empty list for now; RBAC module handles real permission lookup
	return nil, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint, oldPwd, newPwd string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPwd)) != nil {
		return fmt.Errorf("wrong password")
	}
	if len(newPwd) < 8 || len(newPwd) > 64 {
		return fmt.Errorf("password length invalid")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPwd), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.repo.Update(ctx, userID, map[string]any{"password": string(hashed)}); err != nil {
		return err
	}
	return s.repo.IncrementSessionVersion(ctx, userID)
}

func (s *UserService) RevokeSessions(ctx context.Context, userID uint) error {
	return s.repo.IncrementSessionVersion(ctx, userID)
}

// UpdateProfile updates the current user's own profile fields (email, nickname, avatar).
func (s *UserService) UpdateProfile(ctx context.Context, userID uint, email, nickname, avatar string) error {
	updates := map[string]any{}
	if email != "" {
		updates["email"] = email
	}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if len(updates) == 0 {
		return nil
	}
	return s.Update(ctx, userID, updates)
}
