package app

import (
	"context"
	"fmt"
	"testing"

	"kingfisher/extends/menu/domain"
)

type mockMenuRepo struct {
	menus    map[uint]*domain.Menu
	children map[uint]bool
}

func (m *mockMenuRepo) FindByParentID(ctx context.Context, parentID uint) ([]domain.Menu, error) {
	var out []domain.Menu
	for _, menu := range m.menus {
		if menu.ParentID == parentID {
			out = append(out, *menu)
		}
	}
	return out, nil
}

func (m *mockMenuRepo) FindAll(ctx context.Context) ([]domain.Menu, error) {
	var out []domain.Menu
	for _, menu := range m.menus {
		out = append(out, *menu)
	}
	return out, nil
}

func (m *mockMenuRepo) FindByID(ctx context.Context, id uint) (*domain.Menu, error) {
	menu, ok := m.menus[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return menu, nil
}

func (m *mockMenuRepo) FindByRoleIDs(ctx context.Context, roleIDs []uint) ([]domain.Menu, error) {
	return m.FindAll(ctx)
}

func (m *mockMenuRepo) Create(ctx context.Context, menu *domain.Menu) error {
	if m.menus == nil {
		m.menus = map[uint]*domain.Menu{}
	}
	menu.ID = uint(len(m.menus) + 1)
	if menu.Path != "" {
		for _, existing := range m.menus {
			if existing.Path == menu.Path {
				return fmt.Errorf("menu path already exists")
			}
		}
	}
	m.menus[menu.ID] = menu
	return nil
}

func (m *mockMenuRepo) Update(ctx context.Context, id uint, updates map[string]any) error {
	return nil
}

func (m *mockMenuRepo) Delete(ctx context.Context, id uint) error {
	if m.children != nil && m.children[id] {
		return fmt.Errorf("menu has children")
	}
	delete(m.menus, id)
	return nil
}

func (m *mockMenuRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		delete(m.menus, id)
	}
	return nil
}

func (m *mockMenuRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	return nil
}

func (m *mockMenuRepo) HasChildren(ctx context.Context, id uint) (bool, error) {
	return m.children != nil && m.children[id], nil
}

func TestMenuGetTree(t *testing.T) {
	repo := &mockMenuRepo{
		menus: map[uint]*domain.Menu{
			1: {ID: 1, Name: "Dashboard", Path: "/dashboard"},
			2: {ID: 2, Name: "系统管理", Path: "/system"},
			3: {ID: 3, ParentID: 2, Name: "用户管理", Path: "/system/users"},
		},
	}
	svc := NewMenuService(repo, nil)
	tree, err := svc.GetTree(context.Background())
	if err != nil {
		t.Fatal("get tree:", err)
	}
	if len(tree) < 2 {
		t.Errorf("want >=2 root items, got %d", len(tree))
	}
}

func TestMenuCreate(t *testing.T) {
	repo := &mockMenuRepo{menus: map[uint]*domain.Menu{}}
	svc := NewMenuService(repo, nil)
	m := &domain.Menu{Name: "测试", Path: "/test"}
	if err := svc.Create(context.Background(), m); err != nil {
		t.Fatal("create:", err)
	}
	if m.ID == 0 {
		t.Error("id should be set")
	}
}

func TestMenuCreateDuplicatePath(t *testing.T) {
	repo := &mockMenuRepo{
		menus: map[uint]*domain.Menu{
			1: {ID: 1, Name: "existing", Path: "/test"},
		},
	}
	svc := NewMenuService(repo, nil)
	m := &domain.Menu{Name: "dup", Path: "/test"}
	if err := svc.Create(context.Background(), m); err == nil {
		t.Error("duplicate path should fail")
	}
}

func TestMenuDeleteWithChildren(t *testing.T) {
	repo := &mockMenuRepo{
		menus:    map[uint]*domain.Menu{1: {ID: 1, Name: "parent"}},
		children: map[uint]bool{1: true},
	}
	svc := NewMenuService(repo, nil)
	if err := svc.Delete(context.Background(), 1); err == nil {
		t.Error("delete with children should fail")
	}
}

func TestMenuDeleteWithoutChildren(t *testing.T) {
	repo := &mockMenuRepo{
		menus:    map[uint]*domain.Menu{1: {ID: 1, Name: "leaf"}},
		children: map[uint]bool{1: false},
	}
	svc := NewMenuService(repo, nil)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatal("delete:", err)
	}
	if len(repo.menus) != 0 {
		t.Error("menu should be gone")
	}
}

func TestMenuGetByID(t *testing.T) {
	repo := &mockMenuRepo{
		menus: map[uint]*domain.Menu{1: {ID: 1, Name: "Dashboard", Path: "/dashboard"}},
	}
	svc := NewMenuService(repo, nil)
	m, err := svc.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatal("get:", err)
	}
	if m.Name != "Dashboard" {
		t.Error("name mismatch")
	}
}

func TestBuildTree(t *testing.T) {
	menus := []domain.Menu{
		{ID: 1, Name: "Root"},
		{ID: 2, ParentID: 1, Name: "Child"},
		{ID: 3, ParentID: 1, Name: "Child2"},
		{ID: 4, Name: "Root2"},
	}
	tree := buildTree(menus, 0)
	if len(tree) != 2 {
		t.Errorf("want 2 roots, got %d", len(tree))
	}
	if len(tree[0].Children) != 2 {
		t.Errorf("want 2 children, got %d", len(tree[0].Children))
	}
}
