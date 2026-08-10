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

func newTemplateTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&templatePO{}); err != nil {
		t.Fatal(err)
	}
	return db
}

// templateQueryDefs 与 transport handler 保持一致，adapter 测试本地复刻以避免循环依赖。
var templateQueryDefs = query.Defs{
	"name":          {Name: "name", Type: query.TypeString, Searchable: true, Filterable: true},
	"code":          {Name: "code", Type: query.TypeString, Searchable: true, Filterable: true},
	"template_type": {Name: "template_type", Type: query.TypeString, Filterable: true},
	"status":        {Name: "status", Type: query.TypeInt, Filterable: true},
	"remark":        {Name: "remark", Type: query.TypeString, Searchable: true},
	"created_at":    {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

func parseTemplateQuery(t *testing.T, raw string) *query.Query {
	t.Helper()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequestWithContext(context.Background(), "GET", "/?"+raw, nil)
	q, err := query.Parse(c, templateQueryDefs)
	if err != nil {
		t.Fatalf("parse query %q: %v", raw, err)
	}
	return q
}

func TestTemplateRepoCRUD(t *testing.T) {
	repo := NewTemplateRepo(newTemplateTestDB(t))
	ctx := context.Background()

	tpl, err := repo.Create(ctx, "欢迎消息", "welcome", "message", "欢迎 {{nickname}}", "你好 {{nickname}}", 1, "", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if tpl.ID == 0 || tpl.Code != "welcome" || tpl.Title != "欢迎 {{nickname}}" {
		t.Fatalf("unexpected created: %+v", tpl)
	}
	// repo 层不回填时间戳（由 DB trigger / 上层），此处仅验 ID。
	_ = tpl

	// GetByID / GetByCode
	got, err := repo.GetByID(ctx, tpl.ID)
	if err != nil || got.Name != "欢迎消息" {
		t.Fatalf("get by id: %v %+v", err, got)
	}
	byCode, err := repo.GetByCode(ctx, "welcome")
	if err != nil || byCode.ID != tpl.ID {
		t.Fatalf("get by code: %v %+v", err, byCode)
	}
	if _, err := repo.GetByCode(ctx, "missing"); err == nil {
		t.Error("expected miss on unknown code")
	}

	// List：等值过滤 template_type（只命中 message 类型那 1 条）
	if _, err := repo.Create(ctx, "密码重置", "reset", "text", "重置", "内容", 1, "", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	items, total, err := repo.List(ctx, parseTemplateQuery(t, `filter={"template_type":"message"}`))
	if err != nil || total != 1 || len(items) != 1 || items[0].Code != "welcome" {
		t.Fatalf("filter list: err=%v total=%d %+v", err, total, items)
	}

	// 关键词搜索 name（Searchable）
	items, total, err = repo.List(ctx, parseTemplateQuery(t, "q=欢迎"))
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("keyword list: err=%v total=%d %+v", err, total, items)
	}

	// 状态过滤 status=1
	items, total, _ = repo.List(ctx, parseTemplateQuery(t, "filter="+`{"status":1}`))
	if total != 2 || len(items) != 2 {
		t.Fatalf("status filter: total=%d %+v", total, items)
	}

	// Update
	if err := repo.Update(ctx, tpl.ID, "欢迎消息V2", "welcome_v2", "message", "t2", "c2", 0, "r", "1.1.0"); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.GetByID(ctx, tpl.ID)
	if got.Code != "welcome_v2" || got.Status != 0 || got.Name != "欢迎消息V2" {
		t.Fatalf("update not applied: %+v", got)
	}
	// code 更新后旧 code 不可再查到
	if _, err := repo.GetByCode(ctx, "welcome"); err == nil {
		t.Error("old code should be gone after update")
	}

	// UpdateStatusBatch
	if err := repo.UpdateStatusBatch(ctx, []uint{tpl.ID}, 1); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.GetByID(ctx, tpl.ID)
	if got.Status != 1 {
		t.Fatalf("batch status not applied: %+v", got)
	}

	// DeleteBatch
	if err := repo.DeleteBatch(ctx, []uint{tpl.ID, 0}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByID(ctx, tpl.ID); err == nil {
		t.Error("want deleted")
	}

	// 单条 Delete
	if _, err := repo.Create(ctx, "Tmp", "tmp", "text", "", "", 1, "", ""); err != nil {
		t.Fatal(err)
	}
	tmp, _ := repo.GetByCode(ctx, "tmp")
	if err := repo.Delete(ctx, tmp.ID); err != nil {
		t.Fatal(err)
	}
	if _, total, _ := repo.List(ctx, parseTemplateQuery(t, "page=1&page_size=10")); total != 1 {
		t.Fatalf("want 1 remaining, got %d", total)
	}
}
