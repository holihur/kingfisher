package adapter

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"kingfisher/core/query"
)

func newDictTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&dictTypePO{}, &dictEntryPO{}); err != nil {
		t.Fatal(err)
	}
	return db
}

// dictTypeQueryDefs / dictEntryQueryDefs 与 transport handler 保持一致。
var dictTypeQueryDefs = query.Defs{
	"code":       {Name: "code", Type: query.TypeString, Searchable: true, Filterable: true},
	"name":       {Name: "name", Type: query.TypeString, Searchable: true, Filterable: true},
	"remark":     {Name: "remark", Type: query.TypeString, Searchable: true},
	"is_public":  {Name: "is_public", Type: query.TypeBool, Filterable: true},
	"status":     {Name: "status", Type: query.TypeInt, Filterable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

var dictEntryQueryDefs = query.Defs{
	"label":      {Name: "label", Type: query.TypeString, Searchable: true, Filterable: true},
	"value":      {Name: "value", Type: query.TypeString, Searchable: true, Filterable: true},
	"remark":     {Name: "remark", Type: query.TypeString, Searchable: true},
	"status":     {Name: "status", Type: query.TypeInt, Filterable: true},
	"sort":       {Name: "sort", Type: query.TypeInt, Filterable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

func parseDefsQuery(t *testing.T, defs query.Defs, raw string) *query.Query {
	t.Helper()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?"+raw, nil)
	q, err := query.Parse(c, defs)
	if err != nil {
		t.Fatalf("parse query %q: %v", raw, err)
	}
	return q
}

func TestDictTypeRepoCRUD(t *testing.T) {
	db := newDictTestDB(t)
	repo := NewDictTypeRepo(db)
	ctx := context.Background()

	dt, err := repo.Create(ctx, "gender", "性别", true, 1, "", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if dt.ID == 0 || dt.Code != "gender" || !dt.IsPublic {
		t.Fatalf("unexpected created: %+v", dt)
	}
	// 注：code 唯一性由 service 层 + DB 唯一索引保证；sqlite 内存库不建唯一索引，
	// 因此 repo 层不在此处断言冲突（GetByCode 可查到重复的第一条）。

	// GetByCode / GetByID
	byCode, err := repo.GetByCode(ctx, "gender")
	if err != nil || byCode.ID != dt.ID {
		t.Fatalf("get by code: %v %+v", err, byCode)
	}
	got, err := repo.GetByID(ctx, dt.ID)
	if err != nil || got.Name != "性别" {
		t.Fatalf("get by id: %v %+v", err, got)
	}

	// List：等值过滤 is_public + keyword
	if _, err := repo.Create(ctx, "status", "启用", false, 1, "", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	items, total, err := repo.List(ctx, parseDefsQuery(t, dictTypeQueryDefs, `filter={"is_public":true}`))
	if err != nil || total != 1 || len(items) != 1 || items[0].Code != "gender" {
		t.Fatalf("filter is_public: err=%v total=%d %+v", err, total, items)
	}
	items, total, err = repo.List(ctx, parseDefsQuery(t, dictTypeQueryDefs, "q=性别"))
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("keyword: err=%v total=%d %+v", err, total, items)
	}

	// Update
	if err := repo.Update(ctx, dt.ID, "gender_v2", "性别V2", true, 0, "r", "1.1.0"); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.GetByID(ctx, dt.ID)
	if got.Code != "gender_v2" || got.Status != 0 {
		t.Fatalf("update not applied: %+v", got)
	}

	// UpdateStatusBatch
	if err := repo.UpdateStatusBatch(ctx, []uint{dt.ID}, 1); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.GetByID(ctx, dt.ID)
	if got.Status != 1 {
		t.Fatalf("batch status: %+v", got)
	}

	// DeleteBatch
	if err := repo.DeleteBatch(ctx, []uint{dt.ID, 0}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByID(ctx, dt.ID); err == nil {
		t.Error("want deleted")
	}
	if err := repo.Delete(ctx, dt.ID); err != nil {
		t.Fatalf("delete missing id should be no-op: %v", err)
	}
}

func TestDictEntryRepoByTypeID(t *testing.T) {
	db := newDictTestDB(t)
	dtr := NewDictTypeRepo(db)
	entr := NewDictEntryRepo(db)
	ctx := context.Background()

	t1, _ := dtr.Create(ctx, "gender", "性别", true, 1, "", "1.0.0")
	t2, _ := dtr.Create(ctx, "status", "状态", false, 1, "", "1.0.0")
	e1, err := entr.Create(ctx, t1.ID, "男", "male", 1, 1, "", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entr.Create(ctx, t1.ID, "女", "female", 2, 1, "", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if _, err := entr.Create(ctx, t2.ID, "启用", "enabled", 1, 1, "", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	_ = e1

	// ListByTypeID：只返回该 type 下条目，越界隔离
	items, total, err := entr.ListByTypeID(ctx, t1.ID, parseDefsQuery(t, dictEntryQueryDefs, "page=1&page_size=10"))
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("list by type: err=%v total=%d %+v", err, total, items)
	}
	// 关键词/过滤
	filtered, total, err := entr.ListByTypeID(ctx, t1.ID, parseDefsQuery(t, dictEntryQueryDefs, `filter={"value":"male"}`))
	if err != nil || total != 1 || len(filtered) != 1 || filtered[0].Label != "男" {
		t.Fatalf("entry filter: err=%v total=%d %+v", err, total, filtered)
	}

	// ListByTypeCode（公开类型可读，含排序 sort）
	byCode, err := entr.ListByTypeCode(ctx, "gender")
	if err != nil || len(byCode) != 2 {
		t.Fatalf("list by code: err=%v %+v", err, byCode)
	}
	if byCode[0].Sort > byCode[1].Sort {
		t.Error("expected entries sorted by sort asc")
	}

	// GetByID / Update
	got, err := entr.GetByID(ctx, 1)
	if err != nil || got.Label != "男" {
		t.Fatalf("get entry: err=%v %+v", err, got)
	}
	if err := entr.Update(ctx, got.ID, t1.ID, "男V2", "male", 3, 0, "", "1.0.1"); err != nil {
		t.Fatal(err)
	}
	got, _ = entr.GetByID(ctx, got.ID)
	if got.Label != "男V2" || got.Sort != 3 || got.Status != 0 {
		t.Fatalf("entry update not applied: %+v", got)
	}

	// UpdateStatusBatch / DeleteByTypeID
	if err := entr.UpdateStatusBatch(ctx, []uint{got.ID}, 1); err != nil {
		t.Fatal(err)
	}
	if err := entr.DeleteByTypeID(ctx, t1.ID); err != nil {
		t.Fatal(err)
	}
	if _, total, _ := entr.ListByTypeID(ctx, t1.ID, &query.Query{Page: 1, PageSize: 10}); total != 0 {
		t.Fatalf("want 0 after delete by type, got %d", total)
	}
	// t2 的条目不受影响
	if _, total, _ := entr.ListByTypeID(ctx, t2.ID, &query.Query{Page: 1, PageSize: 10}); total != 1 {
		t.Fatalf("t2 entries must survive, got %d", total)
	}
	// DeleteBatch
	if err := entr.DeleteBatch(ctx, []uint{1, 0}); err != nil {
		t.Fatal(err)
	}
	if err := entr.Delete(ctx, 1); err != nil {
		t.Fatalf("delete ok: %v", err)
	}
}