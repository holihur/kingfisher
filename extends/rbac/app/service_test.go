package app

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"kingfisher/core/query"
	"kingfisher/extends/rbac/domain"
)

// mockRoleRepo implements port.RoleRepository
type mockRoleRepo struct {
	roles map[uint]*domain.Role
	perms map[uint][]domain.Permission
	menus map[uint][]domain.Menu
}

func (m *mockRoleRepo) FindAll(ctx context.Context, q *query.Query) ([]domain.Role, int64, error) {
	var out []domain.Role
	for _, r := range m.roles {
		out = append(out, *r)
	}
	return out, int64(len(out)), nil
}

func (m *mockRoleRepo) FindByID(ctx context.Context, id uint) (*domain.Role, error) {
	r, ok := m.roles[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (m *mockRoleRepo) FindByCode(ctx context.Context, code string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Code == code {
			return r, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	if _, ok := m.roles[role.ID]; ok {
		return fmt.Errorf("exists")
	}
	if m.roles == nil {
		m.roles = map[uint]*domain.Role{}
	}
	role.ID = uint(len(m.roles) + 1)
	m.roles[role.ID] = role
	return nil
}

func (m *mockRoleRepo) Update(ctx context.Context, id uint, updates map[string]any) error {
	if _, ok := m.roles[id]; !ok {
		return fmt.Errorf("not found")
	}
	if name, ok := updates["name"].(string); ok {
		m.roles[id].Name = name
	}
	return nil
}

func (m *mockRoleRepo) Delete(ctx context.Context, id uint) error {
	if _, ok := m.roles[id]; !ok {
		return fmt.Errorf("not found")
	}
	delete(m.roles, id)
	return nil
}

func (m *mockRoleRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		delete(m.roles, id)
	}
	return nil
}

func (m *mockRoleRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	return nil
}

func (m *mockRoleRepo) AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error {
	return nil
}

func (m *mockRoleRepo) AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	return nil
}

func (m *mockRoleRepo) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	return []string{"user:list", "user:create"}, nil
}

func (m *mockRoleRepo) GetRolePermissions(ctx context.Context, roleID uint) ([]domain.Permission, error) {
	return m.perms[roleID], nil
}

func (m *mockRoleRepo) GetRoleMenus(ctx context.Context, roleID uint) ([]domain.Menu, error) {
	return m.menus[roleID], nil
}

// mockPermRepo implements port.PermissionRepository
type mockPermRepo struct{}

func (m *mockPermRepo) FindAll(ctx context.Context) ([]domain.Permission, error) {
	return []domain.Permission{
		{ID: 1, Name: "查看用户", Code: "user:list", Resource: "user", Action: "read"},
		{ID: 2, Name: "创建用户", Code: "user:create", Resource: "user", Action: "create"},
	}, nil
}

func TestRoleList(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{
			1: {ID: 1, Name: "超级管理员", Code: "admin", Status: 1, Level: 0},
			3: {ID: 3, Name: "编辑", Code: "editor", Status: 1, Level: 1},
		},
	}
	svc := NewRoleService(repo, nil)
	roles, total, err := svc.List(context.Background(), &query.Query{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal("list:", err)
	}
	if len(roles) != 2 || total != 2 {
		t.Errorf("want 2 roles, got %d total=%d", len(roles), total)
	}
}

func TestRoleCreate(t *testing.T) {
	repo := &mockRoleRepo{roles: map[uint]*domain.Role{}}
	svc := NewRoleService(repo, nil)
	r := &domain.Role{Name: "测试", Code: "test"}
	if err := svc.Create(context.Background(), r); err != nil {
		t.Fatal("create:", err)
	}
	if r.ID == 0 {
		t.Error("id should be set")
	}
}

func TestRoleGetByID(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{
			1: {ID: 1, Name: "admin", Code: "admin", Status: 1},
		},
	}
	svc := NewRoleService(repo, nil)
	role, err := svc.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatal("get:", err)
	}
	if role.Name != "admin" {
		t.Error("name mismatch")
	}
}

func TestRoleGetByIDNotFound(t *testing.T) {
	repo := &mockRoleRepo{roles: map[uint]*domain.Role{}}
	svc := NewRoleService(repo, nil)
	_, err := svc.GetByID(context.Background(), 999)
	if err == nil {
		t.Error("should return error")
	}
}

func TestRoleDelete(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{1: {ID: 1, Name: "test", Code: "test"}},
	}
	svc := NewRoleService(repo, nil)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatal("delete:", err)
	}
	if len(repo.roles) != 0 {
		t.Error("role should be deleted")
	}
}

func TestPermList(t *testing.T) {
	svc := NewPermService(&mockPermRepo{})
	perms, err := svc.List(context.Background())
	if err != nil {
		t.Fatal("list:", err)
	}
	if len(perms) != 2 {
		t.Errorf("want 2 perms, got %d", len(perms))
	}
}

func TestGetUserPermissions(t *testing.T) {
	repo := &mockRoleRepo{}
	svc := NewRoleService(repo, nil)
	perms, err := svc.GetUserPermissions(context.Background(), 1)
	if err != nil {
		t.Fatal("get:", err)
	}
	if len(perms) != 2 {
		t.Error("should return 2 permissions")
	}
}

func TestRoleUpdate(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{1: {ID: 1, Name: "old", Code: "old", Status: 1}},
	}
	svc := NewRoleService(repo, nil)
	if err := svc.Update(context.Background(), 1, map[string]any{"name": "new"}); err != nil {
		t.Fatal("update:", err)
	}
}

func TestAssignPermissions(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{1: {ID: 1, Name: "admin", Code: "admin", Status: 1}},
	}
	svc := NewRoleService(repo, nil)
	if err := svc.AssignPermissions(context.Background(), 1, []uint{1, 3, 5}); err != nil {
		t.Fatal("assign:", err)
	}
}

func TestAssignMenus(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{1: {ID: 1, Name: "admin", Code: "admin", Status: 1}},
	}
	svc := NewRoleService(repo, nil)
	if err := svc.AssignMenus(context.Background(), 1, []uint{1, 2}); err != nil {
		t.Fatal("assign:", err)
	}
}

func TestGetRolePermissions(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{1: {ID: 1, Name: "admin", Code: "admin"}},
		perms: map[uint][]domain.Permission{
			1: {{ID: 1, Name: "查看用户", Code: "user:list"}},
		},
	}
	svc := NewRoleService(repo, nil)
	perms, err := svc.GetRolePermissions(context.Background(), 1)
	if err != nil {
		t.Fatal("get:", err)
	}
	if len(perms) != 1 {
		t.Error("should have 1 perm")
	}
}

func TestGetRoleMenus(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{1: {ID: 1, Name: "admin", Code: "admin"}},
		menus: map[uint][]domain.Menu{
			1: {{ID: 1, Name: "Dashboard"}},
		},
	}
	svc := NewRoleService(repo, nil)
	menus, err := svc.GetRoleMenus(context.Background(), 1)
	if err != nil {
		t.Fatal("get:", err)
	}
	if len(menus) != 1 {
		t.Error("should have 1 menu")
	}
}

func TestStrSliceRoundTrip(t *testing.T) {
	input := []string{"user:list", "menu:create", "role:delete"}
	s := strings.Join(input, ",")
	out := strSlice(s)
	if len(out) != len(input) {
		t.Errorf("want %d, got %d", len(input), len(out))
	}
	for i, v := range input {
		if i < len(out) && out[i] != v {
			t.Errorf("pos %d: want %s, got %s", i, v, out[i])
		}
	}
}

// ---- 未覆盖方法补充：GetRoleLandingPage / 受保护角色的写保护 ----

func TestRoleGetLandingPage(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{1: {ID: 1, Code: "admin", LandingPage: "/dashboard"}},
	}
	svc := NewRoleService(repo, nil)
	page, err := svc.GetRoleLandingPage(context.Background(), 1)
	if err != nil || page != "/dashboard" {
		t.Fatalf("landing page: err=%v page=%q", err, page)
	}
	if _, err := svc.GetRoleLandingPage(context.Background(), 999); err == nil {
		t.Error("missing role should error")
	}
}

func TestRoleUpdateAdminProtected(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{
			1: {ID: 1, Code: "admin", Name: "管理员"},
			2: {ID: 2, Code: "editor", Name: "编辑"},
		},
	}
	svc := NewRoleService(repo, nil)
	if err := svc.Update(context.Background(), 1, map[string]any{"name": "x"}); err == nil {
		t.Error("admin role must not be modifiable")
	}
	if err := svc.Update(context.Background(), 2, map[string]any{"name": "编辑2"}); err != nil {
		t.Fatal("update non-admin:", err)
	}
	if repo.roles[2].Name != "编辑2" {
		t.Error("non-admin update not applied")
	}
}

func TestRoleDeleteAdminProtected(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{
			1: {ID: 1, Code: "admin"},
			2: {ID: 2, Code: "editor"},
		},
	}
	svc := NewRoleService(repo, nil)
	if err := svc.Delete(context.Background(), 1); err == nil {
		t.Error("admin role must not be deletable")
	}
	if err := svc.Delete(context.Background(), 2); err != nil {
		t.Fatal("delete non-admin:", err)
	}
	if err := svc.Delete(context.Background(), 999); err == nil {
		t.Error("missing role delete should error")
	}
}

func TestRoleBatchDeleteAdminProtected(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{
			1: {ID: 1, Code: "admin"},
			2: {ID: 2, Code: "editor"},
		},
	}
	svc := NewRoleService(repo, nil)
	// 任一含 admin 则整批拒绝
	if err := svc.BatchDelete(context.Background(), []uint{1, 2}); err == nil {
		t.Error("batch delete containing admin should fail")
	}
	if len(repo.roles) != 2 {
		t.Error("nothing should be deleted on rejection")
	}
	// 全是非 admin 则成功
	if err := svc.BatchDelete(context.Background(), []uint{2}); err != nil {
		t.Fatal("batch delete non-admin:", err)
	}
	if len(repo.roles) != 1 {
		t.Error("editor should be gone")
	}
}

func TestRoleBatchUpdateStatusAdminProtected(t *testing.T) {
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{
			1: {ID: 1, Code: "admin"},
			2: {ID: 2, Code: "editor"},
		},
	}
	svc := NewRoleService(repo, nil)
	if err := svc.BatchUpdateStatus(context.Background(), []uint{1}, 0); err == nil {
		t.Error("admin batch status must be blocked")
	}
	if err := svc.BatchUpdateStatus(context.Background(), []uint{2}, 0); err != nil {
		t.Fatal("batch status:", err)
	}
}
