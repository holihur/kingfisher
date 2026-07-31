package app

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"kingfisher/core/cache"
	"kingfisher/core/jwt"
	"kingfisher/extends/user/domain"
	"kingfisher/extends/user/port"
)

type AuthService struct {
	repo   port.UserRepository
	cache  cache.Cache
	jwtMgr *jwt.JWTManager
}

func NewAuthService(repo port.UserRepository, c cache.Cache, j *jwt.JWTManager) *AuthService {
	return &AuthService{repo: repo, cache: c, jwtMgr: j}
}

func (s *AuthService) Register(ctx context.Context, username, password, email string) (*domain.User, error) {
	// Check for existing user
	_, err := s.repo.FindByUsername(ctx, username)
	if err == nil {
		return nil, fmt.Errorf("user exists")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{Username: username, Password: string(hashed), Email: email, Status: 1, RoleID: 4} // default viewer
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	user.Password = "" // never return password
	return user, nil
}

var dummyHash = "$2a$12$LJ3m4ys3Lk0TSwHCpNqrIeN5U5Akn5dQUhBvPXFxFG7GqQvHCzB5q"

func (s *AuthService) Login(ctx context.Context, username, password string) (string, string, *domain.User, error) {
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
				return "", "", nil, fmt.Errorf("too many attempts")
			}
		}
		return "", "", nil, fmt.Errorf("wrong password")
	}
	if err != nil {
		return "", "", nil, fmt.Errorf("wrong password")
	}
	if user.Status != 1 {
		return "", "", nil, fmt.Errorf("user disabled")
	}

	// Clear fail count
	if s.cache != nil {
		_ = s.cache.Delete(ctx, "login_fail:"+username)
	}

	access, refresh, err := s.jwtMgr.GenerateToken(ctx, user.ID, "viewer", user.SessionVersion)
	if err != nil {
		return "", "", nil, err
	}
	user.Password = ""
	return access, refresh, user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return s.jwtMgr.RefreshToken(ctx, refreshToken)
}

// User CRUD
type UserService struct {
	repo port.UserRepository
}

func NewUserService(repo port.UserRepository) *UserService { return &UserService{repo: repo} }

func (s *UserService) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id uint, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) List(ctx context.Context, page, pageSize int, keyword string) ([]domain.User, int64, error) {
	return s.repo.FindAll(ctx, page, pageSize, keyword)
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
