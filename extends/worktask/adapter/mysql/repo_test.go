package adapter

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"kingfisher/core/dataaccess"
	"kingfisher/core/query"
)

func TestRepositoryScopesTasks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.TempDir()+"/tasks.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	if err := db.AutoMigrate(&taskPO{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create([]taskPO{{ID: 1, Title: "自己的任务", OwnerID: 10, DepartmentID: 20, Status: "open"}, {ID: 2, Title: "他人任务", OwnerID: 11, DepartmentID: 30, Status: "open"}}).Error; err != nil {
		t.Fatal(err)
	}

	items, total, err := repo.List(context.Background(), &query.Query{Page: 1, PageSize: 20}, dataaccess.Self("owner_id", 10))
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("本人范围应只返回任务 1，total=%d items=%+v", total, items)
	}
	if _, err := repo.GetByID(context.Background(), 2, dataaccess.Self("owner_id", 10)); err == nil {
		t.Fatal("本人范围不应读取他人任务")
	}
	if err := repo.Update(context.Background(), 2, map[string]any{"title": "越权"}, dataaccess.Self("owner_id", 10)); err == nil {
		t.Fatal("本人范围不应更新他人任务")
	}
	_, total, err = repo.List(context.Background(), &query.Query{Page: 1, PageSize: 20}, dataaccess.All())
	if err != nil || total != 2 {
		t.Fatalf("全部范围应返回两条任务，total=%d err=%v", total, err)
	}
}
