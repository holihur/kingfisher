package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

type AuthService struct {
	repo   port.UserRepository
	cache  cache.Cache
	jwtMgr *jwt.JWTManager
	// getLandingPage 返回角色落地页（登录后跳转的页面），由 cmd 注入
	getLandingPage func(ctx context.Context, roleID uint) (string, error)
	// getConfig 读取系统配置值，由 cmd 注入（注册开关、默认注册角色）
	getConfig func(ctx context.Context, key string) (string, error)
	// sendEmail 发送邮件（异步入队），由 cmd 注入 email producer
	sendEmail func(ctx context.Context, to, subject, body string) error
	// renderTemplate 按模板 code 渲染主题与正文（含 {{var}} 替换），由 cmd 注入
	renderTemplate func(ctx context.Context, code string, vars map[string]string) (subject, body string, err error)
	// resetTokenTTL 密码重置 token 有效期（30 分钟）
	resetTokenTTL time.Duration
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
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil || user == nil {
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

// SetConfigProvider 注入系统配置查询函数。
func (s *AuthService) SetConfigProvider(fn func(ctx context.Context, key string) (string, error)) {
	s.getConfig = fn
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

	// Clear fail count
	if s.cache != nil {
		_ = s.cache.Delete(ctx, "login_fail:"+username)
	}

	// 收集所有角色的 ID 与 code（多角色权限/菜单取并集）
	var roleIDs []uint
	var roleCodes []string
	for _, r := range user.Roles {
		roleIDs = append(roleIDs, r.ID)
		roleCodes = append(roleCodes, r.Code)
	}
	if len(roleIDs) == 0 {
		// 兜底：不应发生（用户至少一个角色），避免生成空角色 token
		return "", "", nil, "", fmt.Errorf("user has no roles")
	}
	access, refresh, err := s.jwtMgr.GenerateToken(ctx, user.ID, roleIDs, roleCodes, user.Username, user.SessionVersion)
	if err != nil {
		return "", "", nil, "", err
	}
	// 查询角色落地页（登录后跳转页面），取第一个角色的
	landing := ""
	if s.getLandingPage != nil {
		if lp, e := s.getLandingPage(ctx, roleIDs[0]); e == nil {
			landing = lp
		}
	}
	user.Password = ""
	return access, refresh, user, landing, nil
}

func (s *AuthService) RevokeToken(ctx context.Context, tokenStr string) error {
	return s.jwtMgr.RevokeToken(ctx, tokenStr)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return s.jwtMgr.RefreshToken(ctx, refreshToken)
}

// User CRUD
type UserService struct {
	repo  port.UserRepository
	cache cache.Cache
}

func NewUserService(repo port.UserRepository, c cache.Cache) *UserService {
	return &UserService{repo: repo, cache: c}
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id uint, updates map[string]any) error {
	if err := s.repo.Update(ctx, id, updates); err != nil {
		return err
	}
	// When roles change, invalidate cached permissions so RBAC middleware
	// picks up the new roles' permissions on next request.
	if _, ok := updates["role_ids"]; ok && s.cache != nil {
		_ = s.cache.Delete(ctx, "user:perms:"+strconv.FormatUint(uint64(id), 10))
	}
	return nil
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) BatchDelete(ctx context.Context, ids []uint) error {
	return s.repo.DeleteBatch(ctx, ids)
}

func (s *UserService) BatchUpdateStatus(ctx context.Context, ids []uint, status int) error {
	return s.repo.UpdateStatusBatch(ctx, ids, status)
}

func (s *UserService) CreateUser(ctx context.Context, username, password, email string, roleIDs []uint) (*domain.User, error) {
	_, err := s.repo.FindByUsername(ctx, username)
	if err == nil {
		return nil, fmt.Errorf("user exists")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash: %w", err)
	}
	// 未指定角色时默认访客(4)
	if len(roleIDs) == 0 {
		roleIDs = []uint{4}
	}
	user := &domain.User{Username: username, Password: string(hashed), Email: email, Status: 1, RoleIDs: roleIDs}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
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
