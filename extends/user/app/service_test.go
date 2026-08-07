package app

import (
	"context"
	"fmt"
	"testing"

	"kingfisher/core/config"
	"kingfisher/core/jwt"
	"kingfisher/extends/user/domain"
)

// mockUserRepo implements port.UserRepository for testing
type mockUserRepo struct {
	users     map[string]*mockUserData
	idCounter uint
}

type mockUserData struct {
	id         uint
	username   string
	password   string
	email      string
	status     int
	sessionVer int
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	for _, d := range m.users {
		if d.id == id {
			return &domain.User{
				ID: d.id, Username: d.username, Password: d.password,
				Email: d.email, Status: d.status, SessionVersion: d.sessionVer,
			}, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	d, ok := m.users[username]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return &domain.User{
		ID: d.id, Username: d.username, Password: d.password,
		Email: d.email, Status: d.status, SessionVersion: d.sessionVer,
	}, nil
}

func (m *mockUserRepo) FindAll(ctx context.Context, page, pageSize int, keyword string) ([]domain.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	if m.users == nil {
		m.users = map[string]*mockUserData{}
	}
	if _, ok := m.users[u.Username]; ok {
		return fmt.Errorf("exists")
	}
	m.idCounter++
	u.ID = m.idCounter
	m.users[u.Username] = &mockUserData{
		id: u.ID, username: u.Username, password: u.Password,
		email: u.Email, status: 1, sessionVer: 1,
	}
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, id uint, updates map[string]any) error {
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepo) IncrementSessionVersion(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepo) GetSessionVersion(ctx context.Context, id uint) (int, error) {
	return 1, nil
}

func TestRegisterAndLogin(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*mockUserData{}}
	mgr := jwt.NewJWTManager(config.JWTConfig{Secret: "test", AccessTTL: 1e12, RefreshTTL: 2e12, Issuer: "test"}, nil) // nanoseconds ≈ 16min/33min
	svc := NewAuthService(repo, nil, mgr)

	// Register
	user, err := svc.Register(context.Background(), "testuser", "Abcd1234", "test@test.com")
	if err != nil {
		t.Fatal("register:", err)
	}
	if user.Username != "testuser" {
		t.Error("username mismatch")
	}
	if user.Password != "" {
		t.Error("password should be empty in response")
	}

	// Register duplicate
	_, err = svc.Register(context.Background(), "testuser", "Abcd1234", "")
	if err == nil {
		t.Error("duplicate should fail")
	}

	// Login
	access, refresh, user, err := svc.Login(context.Background(), "testuser", "Abcd1234")
	if err != nil {
		t.Fatal("login:", err)
	}
	if access == "" || refresh == "" {
		t.Error("tokens empty")
	}
	if user.Username != "testuser" {
		t.Error("user mismatch")
	}

	// Login wrong password
	_, _, _, err = svc.Login(context.Background(), "testuser", "wrong")
	if err == nil {
		t.Error("wrong password should fail")
	}
	if err.Error() != "wrong password" {
		t.Error("error msg should be 'wrong password', got:", err.Error())
	}

	// Login non-existent (enumeration resistant)
	_, _, _, err = svc.Login(context.Background(), "nobody", "anything")
	if err == nil {
		t.Error("non-existent should fail")
	}
	if err.Error() != "wrong password" {
		t.Error("should say 'wrong password', got:", err.Error())
	}

	// Refresh
	newAccess, err := svc.RefreshToken(context.Background(), refresh)
	if err != nil {
		t.Fatal("refresh:", err)
	}
	if newAccess == "" {
		t.Error("empty new access")
	}
}

func TestChangePasswordWrongOld(t *testing.T) {
	repo := &mockUserRepo{
		users: map[string]*mockUserData{
			"admin": {id: 1, username: "admin", password: "$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO", status: 1, sessionVer: 1},
		},
		idCounter: 1,
	}
	svc := NewUserService(repo, nil)

	err := svc.ChangePassword(context.Background(), 1, "WrongPass1", "NewPass123")
	if err == nil {
		t.Error("wrong old password should fail")
	}
}

func TestChangePasswordShortNew(t *testing.T) {
	repo := &mockUserRepo{
		users: map[string]*mockUserData{
			"admin": {id: 1, username: "admin", password: "$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO", status: 1, sessionVer: 1},
		},
		idCounter: 1,
	}
	svc := NewUserService(repo, nil)

	err := svc.ChangePassword(context.Background(), 1, "Abcd1234", "short")
	if err == nil {
		t.Error("short new password should fail")
	}
}

func TestUserCreate(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*mockUserData{}}
	svc := NewUserService(repo, nil)
	u, err := svc.CreateUser(context.Background(), "newuser", "Abcd1234", "new@test.com")
	if err != nil {
		t.Fatal("create:", err)
	}
	if u.ID == 0 {
		t.Error("id should be set")
	}
	if u.Username != "newuser" {
		t.Error("username mismatch")
	}
	if u.Password != "" {
		t.Error("password should be empty in response")
	}
}
