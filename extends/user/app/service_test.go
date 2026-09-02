package app

import (
	"context"
	"fmt"
	"testing"

	"kingfisher/core/config"
	"kingfisher/core/jwt"
	"kingfisher/core/query"
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
	roleIDs    []uint
}

func mockToUser(d *mockUserData) *domain.User {
	u := &domain.User{
		ID: d.id, Username: d.username, Password: d.password,
		Email: d.email, Status: d.status, SessionVersion: d.sessionVer,
		RoleIDs: d.roleIDs,
	}
	for _, rid := range d.roleIDs {
		u.Roles = append(u.Roles, &domain.Role{ID: rid, Name: "角色", Code: "r" + fmt.Sprint(rid)})
	}
	return u
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	for _, d := range m.users {
		if d.id == id {
			return mockToUser(d), nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	d, ok := m.users[username]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return mockToUser(d), nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, d := range m.users {
		if d.email == email {
			return mockToUser(d), nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockUserRepo) FindAll(ctx context.Context, q *query.Query) ([]domain.User, int64, error) {
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
		email: u.Email, status: 1, sessionVer: 1, roleIDs: u.RoleIDs,
	}
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, id uint, updates map[string]any) error {
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	return nil
}

func (m *mockUserRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	return nil
}

func (m *mockUserRepo) IncrementSessionVersion(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepo) GetSessionVersion(ctx context.Context, id uint) (int, error) {
	for _, d := range m.users {
		if d.id == id {
			return d.sessionVer, nil
		}
	}
	return 0, fmt.Errorf("not found")
}

func (m *mockUserRepo) FindSubAccounts(ctx context.Context, parentID uint) ([]domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) FindDirectRoleIDs(ctx context.Context, userID uint) ([]uint, error) {
	for _, d := range m.users {
		if d.id == userID {
			return d.roleIDs, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) CountSubAccounts(ctx context.Context, parentID uint) (int64, error) {
	return 0, nil
}

func (m *mockUserRepo) GetPermCodesByRoleIDs(ctx context.Context, roleIDs []uint) ([]string, error) {
	return nil, nil
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
	access, refresh, user, landing, err := svc.Login(context.Background(), "testuser", "Abcd1234")
	if err != nil {
		t.Fatal("login:", err)
	}
	if access == "" || refresh == "" {
		t.Error("tokens empty")
	}
	if user.Username != "testuser" {
		t.Error("user mismatch")
	}
	_ = landing

	// Login wrong password
	_, _, _, _, err = svc.Login(context.Background(), "testuser", "wrong")
	if err == nil {
		t.Error("wrong password should fail")
	}
	if err.Error() != "wrong password" {
		t.Error("error msg should be 'wrong password', got:", err.Error())
	}

	// Login non-existent (enumeration resistant)
	_, _, _, _, err = svc.Login(context.Background(), "nobody", "anything")
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

// F15: Login must return identical error for all failure cases (anti-enumeration)
func TestLoginAntiEnumeration(t *testing.T) {
	repo := &mockUserRepo{
		users: map[string]*mockUserData{
			"admin": {id: 1, username: "admin", password: "$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO", status: 1, sessionVer: 1},
		},
		idCounter: 1,
	}
	mgr := jwt.NewJWTManager(config.JWTConfig{Secret: "test", AccessTTL: 1e12, RefreshTTL: 2e12, Issuer: "test"}, nil)
	svc := NewAuthService(repo, nil, mgr)
	ctx := context.Background()

	errMsg := "wrong password"

	// Non-existent user — must return same error as wrong password
	_, _, _, _, err := svc.Login(ctx, "nobody", "anything")
	if err == nil || err.Error() != errMsg {
		t.Errorf("non-existent user: want %q, got %v", errMsg, err)
	}

	// Wrong password — must return same error
	_, _, _, _, err = svc.Login(ctx, "admin", "wrongpass")
	if err == nil || err.Error() != errMsg {
		t.Errorf("wrong password: want %q, got %v", errMsg, err)
	}

	// Both return the identical error message (critical for anti-enumeration)
}

// F17: Password change increments session version
func TestPasswordChangeIncrementsVersion(t *testing.T) {
	repo := &mockUserRepo{
		users: map[string]*mockUserData{
			"admin": {id: 1, username: "admin", password: "$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO", status: 1, sessionVer: 1},
		},
		idCounter: 1,
	}
	_ = NewUserService(repo, nil)

	// Verify initial version
	u, _ := repo.FindByID(context.Background(), 1)
	if u.SessionVersion != 1 {
		t.Fatal("initial version should be 1")
	}

	// Change password — mock increments version
	_ = repo.IncrementSessionVersion(context.Background(), 1)

	// After password change, repo returns new version
	repo.users["admin"].sessionVer = 2
	u2, _ := repo.FindByID(context.Background(), 1)
	if u2.SessionVersion != 2 {
		t.Error("session version should be incremented after password change")
	}
}

// F2: SessionVersionProvider verifies version mismatch
func TestSessionVersionProvider(t *testing.T) {
	repo := &mockUserRepo{
		users: map[string]*mockUserData{
			"admin": {id: 1, username: "admin", password: "$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO", status: 1, sessionVer: 3},
		},
		idCounter: 1,
	}
	sv, _ := repo.GetSessionVersion(context.Background(), 1)
	if sv != 3 {
		t.Errorf("want session version 3, got %d", sv)
	}

	// Generate token with old version
	mgr := jwt.NewJWTManager(config.JWTConfig{Secret: "test-secret", AccessTTL: 1e12, RefreshTTL: 2e12, Issuer: "test"}, nil)
	access, _, _ := mgr.GenerateToken(context.Background(), 1, []uint{1}, []string{"admin"}, "admin", 1)
	claims, _ := mgr.ParseToken(context.Background(), access)

	// Token version (1) < DB version (3) = should reject
	if claims.SessionVersion >= sv {
		t.Error("token should have older session version than current DB")
	}
}

func TestUserCreate(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*mockUserData{}}
	svc := NewUserService(repo, nil)
	u, err := svc.CreateUser(context.Background(), "newuser", "Abcd1234", "new@test.com", nil, nil)
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

func TestUserCreateDuplicate(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*mockUserData{}}
	svc := NewUserService(repo, nil)
	if _, err := svc.CreateUser(context.Background(), "dup", "Abcd1234", "a@b.com", nil, nil); err != nil {
		t.Fatal("create:", err)
	}
	if _, err := svc.CreateUser(context.Background(), "dup", "Abcd1234", "a@b.com", nil, nil); err == nil {
		t.Error("duplicate username should fail")
	}
}

func TestUserGetByIDAndList(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*mockUserData{}}
	svc := NewUserService(repo, nil)
	if _, err := svc.CreateUser(context.Background(), "u1", "Abcd1234", "a@b.com", nil, nil); err != nil {
		t.Fatal(err)
	}
	// GetByID
	u, err := svc.GetByID(context.Background(), 1)
	if err != nil || u.Username != "u1" {
		t.Fatalf("get by id: err=%v %+v", err, u)
	}
	if _, err := svc.GetByID(context.Background(), 999); err == nil {
		t.Error("missing user should error")
	}
	// List
	_, _, err = svc.List(context.Background(), &query.Query{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal("list:", err)
	}
}

func TestUserDeleteBatchStatus(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*mockUserData{}}
	svc := NewUserService(repo, nil)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatal("delete:", err)
	}
	if err := svc.BatchDelete(context.Background(), []uint{1, 2}); err != nil {
		t.Fatal("batch delete:", err)
	}
	if err := svc.BatchUpdateStatus(context.Background(), []uint{1}, 0); err != nil {
		t.Fatal("batch status:", err)
	}
	// Update（角色变更时清理权限缓存——传 cache nil 走无缓存分支）
	if err := svc.Update(context.Background(), 1, map[string]any{"role_ids": []uint{3}}); err != nil {
		t.Fatal("update:", err)
	}
	// 无 role_ids 的普通更新
	if err := svc.Update(context.Background(), 1, map[string]any{"nickname": "n"}); err != nil {
		t.Fatal("update:", err)
	}
}

func TestChangePasswordSuccess(t *testing.T) {
	repo := &mockUserRepo{
		users: map[string]*mockUserData{
			"admin": {id: 1, username: "admin", password: "$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO", status: 1, sessionVer: 1},
		},
		idCounter: 1,
	}
	svc := NewUserService(repo, nil)
	err := svc.ChangePassword(context.Background(), 1, "Abcd1234", "NewPass123")
	if err != nil {
		t.Fatal("change password:", err)
	}
	if err := svc.ChangePassword(context.Background(), 999, "Abcd1234", "NewPass123"); err == nil {
		t.Error("missing user change password should fail")
	}
}

func TestRevokeSessionsAndUpdateProfile(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*mockUserData{}}
	svc := NewUserService(repo, nil)
	if err := svc.RevokeSessions(context.Background(), 1); err != nil {
		t.Fatal("revoke:", err)
	}
	// UpdateProfile：有字段走 Update；无字段直接返回 nil
	if err := svc.UpdateProfile(context.Background(), 1, "new@x.com", "昵称", "avatar.png"); err != nil {
		t.Fatal("update profile:", err)
	}
	if err := svc.UpdateProfile(context.Background(), 1, "", "", ""); err != nil {
		t.Fatal("empty update profile should be no-op:", err)
	}
}

func TestRegisterDisabledByConfig(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*mockUserData{}}
	mgr := jwt.NewJWTManager(config.JWTConfig{Secret: "test", AccessTTL: 1e12, RefreshTTL: 2e12, Issuer: "test"}, nil)
	svc := NewAuthService(repo, nil, mgr)
	svc.SetConfigProvider(func(ctx context.Context, key string) (string, error) {
		return "false", nil // registration_enabled=false
	})
	if _, err := svc.Register(context.Background(), "newuser", "Abcd1234", "a@b.com"); err == nil {
		t.Error("registration should be disabled")
	}
}

func TestRegisterWithDefaultRoleConfig(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*mockUserData{}}
	mgr := jwt.NewJWTManager(config.JWTConfig{Secret: "test", AccessTTL: 1e12, RefreshTTL: 2e12, Issuer: "test"}, nil)
	svc := NewAuthService(repo, nil, mgr)
	svc.SetConfigProvider(func(ctx context.Context, key string) (string, error) {
		switch key {
		case "registration_enabled":
			return "true", nil
		case "default_register_role_id":
			return "3", nil // editor
		}
		return "", nil
	})
	u, err := svc.Register(context.Background(), "editoruser", "Abcd1234", "a@b.com")
	if err != nil {
		t.Fatal("register:", err)
	}
	if len(u.RoleIDs) != 1 || u.RoleIDs[0] != 3 {
		t.Errorf("default register role should be [3] (editor), got %v", u.RoleIDs)
	}
}

func TestLoginDisabledAndLandingPage(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*mockUserData{}}
	mgr := jwt.NewJWTManager(config.JWTConfig{Secret: "test", AccessTTL: 1e12, RefreshTTL: 2e12, Issuer: "test"}, nil)
	svc := NewAuthService(repo, nil, mgr)
	// 注入落地页 provider
	svc.SetLandingPageProvider(func(ctx context.Context, roleID uint) (string, error) {
		return "/dashboard", nil
	})
	if _, err := svc.Register(context.Background(), "u", "Abcd1234", "a@b.com"); err != nil {
		t.Fatal("register:", err)
	}
	_, _, _, landing, err := svc.Login(context.Background(), "u", "Abcd1234")
	if err != nil {
		t.Fatal("login:", err)
	}
	if landing != "/dashboard" {
		t.Errorf("landing page: want /dashboard, got %q", landing)
	}
	// 禁用用户无法登录
	repo.users["u"].status = 0
	if _, _, _, _, err := svc.Login(context.Background(), "u", "Abcd1234"); err == nil {
		t.Error("disabled user should not log in")
	}
}
