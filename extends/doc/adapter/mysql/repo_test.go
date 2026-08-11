package adapter

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"kingfisher/core/query"
	"kingfisher/extends/doc/domain"
)

// newTestRepo 建一个内存 SQLite + 自动建表的测试仓储。
func newTestRepo(t *testing.T) (*DocRepo, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.TempDir()+"/test.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1) // 单连接，保证同库
	// 文档 owner_name 关联需要 users 表
	if err := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT)").Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	if err := db.AutoMigrate(&docDirectoryPO{}, &docDirRolePO{}, &documentPO{}, &docVersionPO{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return NewDocRepo(db), db
}

func seedVisibleData(t *testing.T, db *gorm.DB) {
	t.Helper()
	// 目录：1 全角色可见(1,3,4)；2 仅 admin(1)
	db.Create(&docDirectoryPO{ID: 1, ParentID: 0, Name: "公开目录", Sort: 1})
	db.Create(&docDirectoryPO{ID: 2, ParentID: 0, Name: "受限目录", Sort: 2})
	db.Create(&docDirRolePO{DirID: 1, RoleID: 1})
	db.Create(&docDirRolePO{DirID: 1, RoleID: 3})
	db.Create(&docDirRolePO{DirID: 1, RoleID: 4})
	db.Create(&docDirRolePO{DirID: 2, RoleID: 1})
	// 文档：doc1 已发布共享；doc2 草稿（作者 2）；doc3 private 已发布（作者 2）
	db.Create(&documentPO{ID: 1, DirID: 1, Title: "已发布共享", Content: "c1", OwnerID: 1, Visibility: "shared", Status: "published", CurrentVersion: 1})
	db.Create(&documentPO{ID: 2, DirID: 1, Title: "草稿", Content: "c2", OwnerID: 2, Visibility: "shared", Status: "draft", CurrentVersion: 1})
	db.Create(&documentPO{ID: 3, DirID: 1, Title: "私有", Content: "c3", OwnerID: 2, Visibility: "private", Status: "published", CurrentVersion: 1})
	db.Create(&documentPO{ID: 4, DirID: 2, Title: "受限目录文档", Content: "c4", OwnerID: 1, Visibility: "shared", Status: "published", CurrentVersion: 1})
}

func TestVisibleScope(t *testing.T) {
	repo, db := newTestRepo(t)
	seedVisibleData(t, db)
	ctx := context.Background()

	// 非作者 viewer(role 4) 看 dir1：只能见已发布共享的 doc1
	viewerDocs, total, err := repo.ListDocs(ctx, 1, &query.Query{Page: 1, PageSize: 20}, 99, []uint{4}, false)
	if err != nil {
		t.Fatalf("list viewer: %v", err)
	}
	if total != 1 || len(viewerDocs) != 1 || viewerDocs[0].Title != "已发布共享" {
		t.Fatalf("viewer 应只见 已发布共享，got total=%d titles=%+v", total, viewerDocs)
	}

	// 作者本人(2) 看 dir1：可见自己的草稿+私有+共享
	_, total, err = repo.ListDocs(ctx, 1, &query.Query{Page: 1, PageSize: 20}, 2, []uint{3}, false)
	if err != nil {
		t.Fatalf("list owner: %v", err)
	}
	if total != 3 {
		t.Fatalf("作者应见 3 篇（草稿+私有+共享），got %d", total)
	}

	// 作者(2) 看 dir2（仅 role1 可见，role 不含 1）：不可见
	dir2Docs, _, err := repo.ListDocs(ctx, 2, &query.Query{Page: 1, PageSize: 20}, 2, []uint{3}, false)
	if err != nil {
		t.Fatalf("list dir2: %v", err)
	}
	if len(dir2Docs) != 0 {
		t.Fatalf("作者无 role1 不该看到受限目录文档，got %d", len(dir2Docs))
	}

	// admin：全量可见（含草稿/私有/受限目录）
	adminDocs, total, err := repo.ListDocs(ctx, 1, &query.Query{Page: 1, PageSize: 20}, 1, []uint{1}, true)
	if err != nil {
		t.Fatalf("list admin: %v", err)
	}
	if total != 3 {
		t.Fatalf("admin 应见 dir1 全部 3 篇，got %d", total)
	}
	_ = adminDocs
}

func TestGetDocByIDVisibility(t *testing.T) {
	repo, db := newTestRepo(t)
	seedVisibleData(t, db)
	ctx := context.Background()

	// 私有文档对他人不可读（404 语义）
	if _, err := repo.GetDocByID(ctx, 3, 99, []uint{4}, false); err == nil {
		t.Fatal("非作者读私有文档应失败")
	}
	// 作者可读
	if _, err := repo.GetDocByID(ctx, 3, 2, []uint{3}, false); err != nil {
		t.Fatalf("作者读私有文档应成功: %v", err)
	}
	// admin 可读
	if _, err := repo.GetDocByID(ctx, 3, 1, []uint{1}, true); err != nil {
		t.Fatalf("admin 读私有文档应成功: %v", err)
	}
}

func TestCreateUpdateVersionAndConflict(t *testing.T) {
	repo, db := newTestRepo(t)
	seedVisibleData(t, db)
	ctx := context.Background()

	// 创建文档 + 版本 1
	doc, err := repo.CreateWithVersion(ctx, &domain.Document{DirID: 1, Title: "新文档", Content: "<p>v1</p>", OwnerID: 1, Visibility: "shared", Status: "draft"}, &domain.DocVersion{OwnerID: 1, Note: "init"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if doc.CurrentVersion != 1 {
		t.Fatalf("初始版本应为 1，got %d", doc.CurrentVersion)
	}

	// 更新 → 版本 2
	if err := repo.UpdateWithVersion(ctx, doc.ID, "新文档v2", "<p>v2</p>", "shared", 1, "update"); err != nil {
		t.Fatalf("update: %v", err)
	}
	vers, err := repo.ListVersions(ctx, doc.ID)
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(vers) != 2 {
		t.Fatalf("应有 2 个版本，got %d", len(vers))
	}
	got, _ := repo.GetDocByID(ctx, doc.ID, 1, []uint{1}, true)
	if got.Content != "<p>v2</p>" || got.CurrentVersion != 2 {
		t.Fatalf("更新后 content/current_version 不符: %+v", got)
	}

	// 还原到 v1 → 追加版本 3，内容回退
	if err := repo.RestoreToVersion(ctx, doc.ID, 1, 1, "restore"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	got2, _ := repo.GetDocByID(ctx, doc.ID, 1, []uint{1}, true)
	if got2.Content != "<p>v1</p>" || got2.CurrentVersion != 3 {
		t.Fatalf("还原后应回到 v1 内容且版本为 3: %+v", got2)
	}

	// 删除文档级联删版本
	if err := repo.DeleteDoc(ctx, doc.ID); err != nil {
		t.Fatalf("delete doc: %v", err)
	}
	vers2, _ := repo.ListVersions(ctx, doc.ID)
	if len(vers2) != 0 {
		t.Fatalf("删除后版本应级联清空，got %d", len(vers2))
	}
}

func TestSetDirRolesReplaces(t *testing.T) {
	repo, db := newTestRepo(t)
	seedVisibleData(t, db)
	ctx := context.Background()

	if err := repo.SetDirRoles(ctx, 1, []uint{1, 3}); err != nil {
		t.Fatalf("set roles: %v", err)
	}
	ids, err := repo.GetDirRoleIDs(ctx, 1)
	if err != nil {
		t.Fatalf("get roles: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("应只剩 2 个角色，got %v", ids)
	}
	// 清空 = 全量替换为空
	if err := repo.SetDirRoles(ctx, 1, []uint{}); err != nil {
		t.Fatalf("clear roles: %v", err)
	}
	ids2, _ := repo.GetDirRoleIDs(ctx, 1)
	if len(ids2) != 0 {
		t.Fatalf("清空后应无角色，got %v", ids2)
	}
}
