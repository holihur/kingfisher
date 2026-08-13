package adapter

import (
	"context"
	"sort"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTestRoleRepo 建一个内存 SQLite + 关联表自动迁移的测试仓储。
func newTestRoleRepo(t *testing.T) (*RoleRepo, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.TempDir()+"/test.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&rolePO{}, &permissionPO{}, &rolePermissionPO{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	// user_roles 与部门关联表（GetUserPermissions UNION 需要）
	stmts := []string{
		`CREATE TABLE user_roles (user_id INTEGER NOT NULL, role_id INTEGER NOT NULL, PRIMARY KEY (user_id, role_id))`,
		`CREATE TABLE user_departments (user_id INTEGER NOT NULL, department_id INTEGER NOT NULL, PRIMARY KEY (user_id, department_id))`,
		`CREATE TABLE department_roles (department_id INTEGER NOT NULL, role_id INTEGER NOT NULL, PRIMARY KEY (department_id, role_id))`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("exec %s: %v", s, err)
		}
	}
	return NewRoleRepo(db), db
}

func TestGetUserPermissionsUnionDirectAndDept(t *testing.T) {
	repo, db := newTestRoleRepo(t)
	ctx := context.Background()

	// 角色：role1(admin) role3(editor)
	db.Create(&rolePO{ID: 1, Name: "管理员", Code: "admin"})
	db.Create(&rolePO{ID: 3, Name: "编辑", Code: "editor"})
	// 权限：user:list(user 的 read)、doc:list、role:list
	db.Create(&permissionPO{ID: 1, Name: "查看用户", Code: "user:list", Resource: "user", Action: "read"})
	db.Create(&permissionPO{ID: 33, Name: "查看文档", Code: "doc:list", Resource: "doc", Action: "read"})
	db.Create(&permissionPO{ID: 9, Name: "查看角色", Code: "role:list", Resource: "role", Action: "read"})
	// role_permissions：role1→user:list+role:list；role3→doc:list
	db.Create(&rolePermissionPO{RoleID: 1, PermissionID: 1})
	db.Create(&rolePermissionPO{RoleID: 1, PermissionID: 9})
	db.Create(&rolePermissionPO{RoleID: 3, PermissionID: 33})
	// 用户 10：直接角色 role1（admin）；部门 5 挂 role3（editor）
	db.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES (10, 1)`)
	db.Exec(`INSERT INTO user_departments (user_id, department_id) VALUES (10, 5)`)
	db.Exec(`INSERT INTO department_roles (department_id, role_id) VALUES (5, 3)`)

	codes, err := repo.GetUserPermissions(ctx, 10)
	if err != nil {
		t.Fatalf("get perms: %v", err)
	}
	sort.Strings(codes)
	want := []string{"doc:list", "role:list", "user:list"}
	if len(codes) != len(want) {
		t.Fatalf("应合并直接+部门角色权限，got %v want %v", codes, want)
	}
	for i := range want {
		if codes[i] != want[i] {
			t.Fatalf("权限码不匹配，got %v want %v", codes, want)
		}
	}
}

func TestGetUserPermissionsDeptOnly(t *testing.T) {
	repo, db := newTestRoleRepo(t)
	ctx := context.Background()

	db.Create(&rolePO{ID: 3, Name: "编辑", Code: "editor"})
	db.Create(&permissionPO{ID: 33, Name: "查看文档", Code: "doc:list", Resource: "doc", Action: "read"})
	db.Create(&rolePermissionPO{RoleID: 3, PermissionID: 33})
	// 用户 20 无直接角色，仅通过部门 6 继承 role3
	db.Exec(`INSERT INTO user_departments (user_id, department_id) VALUES (20, 6)`)
	db.Exec(`INSERT INTO department_roles (department_id, role_id) VALUES (6, 3)`)

	codes, err := repo.GetUserPermissions(ctx, 20)
	if err != nil {
		t.Fatalf("get perms: %v", err)
	}
	if len(codes) != 1 || codes[0] != "doc:list" {
		t.Fatalf("应仅通过部门继承 doc:list，got %v", codes)
	}
}

func TestGetUserPermissionsDirectOnlyNoDept(t *testing.T) {
	repo, db := newTestRoleRepo(t)
	ctx := context.Background()

	db.Create(&rolePO{ID: 1, Name: "管理员", Code: "admin"})
	db.Create(&permissionPO{ID: 1, Name: "查看用户", Code: "user:list", Resource: "user", Action: "read"})
	db.Create(&rolePermissionPO{RoleID: 1, PermissionID: 1})
	// 用户 30 只有直接角色，无部门
	db.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES (30, 1)`)

	codes, err := repo.GetUserPermissions(ctx, 30)
	if err != nil {
		t.Fatalf("get perms: %v", err)
	}
	if len(codes) != 1 || codes[0] != "user:list" {
		t.Fatalf("应仅返回直接角色权限 user:list，got %v", codes)
	}
}
