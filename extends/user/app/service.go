package app

import (
	"context"
	"fmt"
	"strconv"
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
}

func NewAuthService(repo port.UserRepository, c cache.Cache, j *jwt.JWTManager) *AuthService {
	return &AuthService{repo: repo, cache: c, jwtMgr: j}
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
	user := &domain.User{Username: username, Password: string(hashed), Email: email, Status: 1, RoleID: roleID}
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

	roleCode := "viewer"
	if user.RoleID == 1 {
		roleCode = "admin"
	}
	if user.RoleID == 3 {
		roleCode = "editor"
	}
	access, refresh, err := s.jwtMgr.GenerateToken(ctx, user.ID, user.RoleID, roleCode, user.Username, user.SessionVersion)
	if err != nil {
		return "", "", nil, "", err
	}
	// 查询角色落地页（登录后跳转页面）
	landing := ""
	if s.getLandingPage != nil {
		if lp, e := s.getLandingPage(ctx, user.RoleID); e == nil {
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
	// When role changes, invalidate cached permissions so RBAC middleware
	// picks up the new role's permissions on next request.
	if _, ok := updates["role_id"]; ok && s.cache != nil {
		_ = s.cache.Delete(ctx, "user:perms:"+strconv.Itoa(int(id)))
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

func (s *UserService) CreateUser(ctx context.Context, username, password, email string) (*domain.User, error) {
	_, err := s.repo.FindByUsername(ctx, username)
	if err == nil {
		return nil, fmt.Errorf("user exists")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash: %w", err)
	}
	user := &domain.User{Username: username, Password: string(hashed), Email: email, Status: 1, RoleID: 4}
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
