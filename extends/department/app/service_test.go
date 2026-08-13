package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"kingfisher/core/errcode"
	adapter "kingfisher/extends/department/adapter/mysql"
	"kingfisher/extends/department/domain"
)

// mockCache 记录 DeleteByPattern 调用（部门角色变更应清权限缓存）
type mockCache struct {
	patterns []string
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) { return "", nil }
func (m *mockCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	return nil
}
func (m *mockCache) Delete(ctx context.Context, keys ...string) error { return nil }
func (m *mockCache) DeleteByPattern(ctx context.Context, pattern string) error {
	m.patterns = append(m.patterns, pattern)
	return nil
}
func (m *mockCache) Exists(ctx context.Context, key string) (bool, error) { return false, nil }
func (m *mockCache) Incr(ctx context.Context, key string) (int64, error)  { return 0, nil }
func (m *mockCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

// newTestSvc 内存 SQLite（建 department 相关表）+ mockCache 的测试服务。
func newTestSvc(t *testing.T) (*DepartmentService, *adapter.DepartmentRepo, *mockCache) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.TempDir()+"/test.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	stmts := []string{
		`CREATE TABLE departments (id INTEGER PRIMARY KEY, parent_id INTEGER DEFAULT 0, name TEXT NOT NULL, sort INTEGER DEFAULT 0, status INTEGER DEFAULT 1, remark TEXT DEFAULT '', version TEXT DEFAULT '1.0.0', created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE user_departments (user_id INTEGER NOT NULL, department_id INTEGER NOT NULL, PRIMARY KEY (user_id, department_id))`,
		`CREATE TABLE department_roles (department_id INTEGER NOT NULL, role_id INTEGER NOT NULL, PRIMARY KEY (department_id, role_id))`,
		`CREATE TABLE roles (id INTEGER PRIMARY KEY, name TEXT, code TEXT)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("exec %s: %v", s, err)
		}
	}
	repo := adapter.NewDepartmentRepo(db)
	cache := &mockCache{}
	return NewDepartmentService(repo, cache), repo, cache
}

func TestTreeBuildsHierarchy(t *testing.T) {
	svc, repo, _ := newTestSvc(t)
	ctx := context.Background()
	// 平铺插入：技术部(1)、产品部(2)、后端组(3,parent=1)
	if err := repo.Create(ctx, &domain.Department{ID: 1, ParentID: 0, Name: "技术部", Sort: 1}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, &domain.Department{ID: 2, ParentID: 0, Name: "产品部", Sort: 2}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, &domain.Department{ID: 3, ParentID: 1, Name: "后端组", Sort: 1}); err != nil {
		t.Fatal(err)
	}
	tree, err := svc.Tree(ctx)
	if err != nil {
		t.Fatalf("tree: %v", err)
	}
	if len(tree) != 2 {
		t.Fatalf("根部门应 2 个，got %d", len(tree))
	}
	if len(tree[0].Children) != 1 || tree[0].Children[0].Name != "后端组" {
		t.Fatalf("技术部应有 1 个子部门，got %+v", tree[0])
	}
}

func TestDeleteRejectsWithChildren(t *testing.T) {
	svc, repo, _ := newTestSvc(t)
	ctx := context.Background()
	if err := repo.Create(ctx, &domain.Department{ID: 1, ParentID: 0, Name: "技术部"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, &domain.Department{ID: 2, ParentID: 1, Name: "后端组"}); err != nil {
		t.Fatal(err)
	}
	err := svc.Delete(ctx, 1)
	if err == nil {
		t.Fatal("有子部门应拒绝删除")
	}
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrDeptHasChildren {
		t.Fatalf("应返回 ErrDeptHasChildren，got %v", err)
	}
}

func TestDeleteLeafFlushesCache(t *testing.T) {
	svc, repo, cache := newTestSvc(t)
	ctx := context.Background()
	if err := repo.Create(ctx, &domain.Department{ID: 1, ParentID: 0, Name: "产品部"}); err != nil {
		t.Fatal(err)
	}
	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("删除叶子部门: %v", err)
	}
	if len(cache.patterns) != 1 || cache.patterns[0] != "user:perms:*" {
		t.Fatalf("删除部门应清权限缓存，got %v", cache.patterns)
	}
}

func TestAssignRolesFlushesCache(t *testing.T) {
	svc, repo, cache := newTestSvc(t)
	ctx := context.Background()
	if err := repo.Create(ctx, &domain.Department{ID: 1, ParentID: 0, Name: "技术部"}); err != nil {
		t.Fatal(err)
	}
	if err := svc.AssignRoles(ctx, 1, []uint{3}); err != nil {
		t.Fatalf("assign roles: %v", err)
	}
	if len(cache.patterns) != 1 || cache.patterns[0] != "user:perms:*" {
		t.Fatalf("分配部门角色应清权限缓存，got %v", cache.patterns)
	}
}
