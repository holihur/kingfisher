package adapter

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"kingfisher/core/query"
	"kingfisher/extends/department/domain"
)

// newTestDeptRepo 建一个内存 SQLite + 自动建表的测试仓储。
func newTestDeptRepo(t *testing.T) (*DepartmentRepo, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.TempDir()+"/test.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1) // 单连接，保证同库
	if err := db.AutoMigrate(&departmentPO{}, &departmentRolePO{}, &userDepartmentPO{}, &rolePO{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return NewDepartmentRepo(db), db
}

func TestDeptRepoCRUDAndTree(t *testing.T) {
	repo, db := newTestDeptRepo(t)
	ctx := context.Background()

	// 角色表建两个角色供关联读取
	db.Create(&rolePO{ID: 1, Name: "管理员", Code: "admin"})
	db.Create(&rolePO{ID: 3, Name: "编辑", Code: "editor"})

	// 创建：技术部(1)、后端组(2, parent=1)
	tech := &domain.Department{ParentID: 0, Name: "技术部", Sort: 1, Status: 1, Remark: "研发"}
	if err := repo.Create(ctx, tech); err != nil {
		t.Fatalf("create tech: %v", err)
	}
	if tech.ID == 0 {
		t.Fatal("create should set ID")
	}
	backend := &domain.Department{ParentID: tech.ID, Name: "后端组", Sort: 1, Status: 1}
	if err := repo.Create(ctx, backend); err != nil {
		t.Fatalf("create backend: %v", err)
	}

	// FindAll（扁平，按 sort/id）
	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("应 2 个部门，got %d", len(all))
	}

	// SetRoles + GetByID 含角色
	if err := repo.SetRoles(ctx, tech.ID, []uint{1, 3}); err != nil {
		t.Fatalf("set roles: %v", err)
	}
	got, err := repo.GetByID(ctx, tech.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if len(got.RoleIDs) != 2 || len(got.Roles) != 2 {
		t.Fatalf("应挂 2 个角色，got ids=%v roles=%d", got.RoleIDs, len(got.Roles))
	}
	if got.Roles[0].Name != "管理员" {
		t.Fatalf("角色名应解析，got %v", got.Roles)
	}

	// SetRoles 替换
	if err := repo.SetRoles(ctx, tech.ID, []uint{3}); err != nil {
		t.Fatalf("set roles replace: %v", err)
	}
	got2, _ := repo.GetByID(ctx, tech.ID)
	if len(got2.RoleIDs) != 1 || got2.RoleIDs[0] != 3 {
		t.Fatalf("替换后应仅剩 role 3，got %v", got2.RoleIDs)
	}

	// HasChildren
	has, err := repo.HasChildren(ctx, tech.ID)
	if err != nil || !has {
		t.Fatalf("技术部应有子部门，has=%v err=%v", has, err)
	}

	// Update
	if err := repo.Update(ctx, tech.ID, map[string]any{"name": "技术中心", "sort": 5}); err != nil {
		t.Fatalf("update: %v", err)
	}
	got3, _ := repo.GetByID(ctx, tech.ID)
	if got3.Name != "技术中心" || got3.Sort != 5 {
		t.Fatalf("update 未生效，got %+v", got3)
	}
}

func TestDeptRepoListPageSubtreeFilter(t *testing.T) {
	repo, db := newTestDeptRepo(t)
	ctx := context.Background()
	// 技术部(1) + 后端组(2,parent=1) + 产品部(3,parent=0)
	for _, d := range []*domain.Department{
		{ID: 1, ParentID: 0, Name: "技术部", Sort: 1},
		{ID: 2, ParentID: 1, Name: "后端组", Sort: 1},
		{ID: 3, ParentID: 0, Name: "产品部", Sort: 2},
	} {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatal(err)
		}
	}

	// 无筛选：全部
	_, total, err := repo.ListPage(ctx, &query.Query{Page: 1, PageSize: 20})
	if err != nil || total != 3 {
		t.Fatalf("全部应 3 条，got total=%d err=%v", total, err)
	}
	// subtree_id=1：技术部 + 后端组（不含产品部）
	sub, total, err := repo.ListPage(ctx, &query.Query{Page: 1, PageSize: 20, Filters: []query.Condition{{Field: "subtree_id", Op: query.OpEq, Value: uint64(1)}}})
	if err != nil || total != 2 {
		t.Fatalf("subtree 应 2 条，got total=%d err=%v", total, err)
	}
	names := map[string]bool{}
	for _, d := range sub {
		names[d.Name] = true
	}
	if !names["技术部"] || !names["后端组"] || names["产品部"] {
		t.Fatalf("subtree 应含 技术部+后端组，不含产品部，got %v", names)
	}
	_ = db
}

func TestDeptRepoDeleteCascade(t *testing.T) {
	repo, db := newTestDeptRepo(t)
	ctx := context.Background()

	db.Create(&rolePO{ID: 3, Name: "编辑", Code: "editor"})
	dept := &domain.Department{ParentID: 0, Name: "产品部", Sort: 1, Status: 1}
	if err := repo.Create(ctx, dept); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.SetRoles(ctx, dept.ID, []uint{3}); err != nil {
		t.Fatalf("set roles: %v", err)
	}
	// 造一条 user_departments 成员
	db.Exec("INSERT INTO user_departments (user_id, department_id) VALUES (10, ?)", dept.ID)

	// 删除应级联清理关联表
	if err := repo.Delete(ctx, dept.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	var drCount int64
	db.Model(&departmentRolePO{}).Where("department_id = ?", dept.ID).Count(&drCount)
	if drCount != 0 {
		t.Fatalf("department_roles 应级联删除，got %d", drCount)
	}
	var udCount int64
	db.Model(&userDepartmentPO{}).Where("department_id = ?", dept.ID).Count(&udCount)
	if udCount != 0 {
		t.Fatalf("user_departments 应级联删除，got %d", udCount)
	}
	var deptCount int64
	db.Model(&departmentPO{}).Where("id = ?", dept.ID).Count(&deptCount)
	if deptCount != 0 {
		t.Fatalf("部门本身应删除，got %d", deptCount)
	}
}
