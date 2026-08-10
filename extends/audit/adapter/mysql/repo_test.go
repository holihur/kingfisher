package adapter

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"kingfisher/core/query"
	"kingfisher/extends/audit/domain"
)

func newAuditTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&auditPO{}); err != nil {
		t.Fatal(err)
	}
	return db
}

// auditQueryDefs 与 transport handler 保持一致，adapter 测试本地复刻以避免循环依赖。
var auditQueryDefs = query.Defs{
	"username":    {Name: "username", Type: query.TypeString, Searchable: true, Filterable: true},
	"ip":          {Name: "ip", Type: query.TypeString, Searchable: true},
	"user_agent":  {Name: "user_agent", Type: query.TypeString, Searchable: true},
	"user_id":     {Name: "user_id", Type: query.TypeUint, Filterable: true},
	"resource":    {Name: "resource", Type: query.TypeString, Filterable: true},
	"action":      {Name: "action", Type: query.TypeString, Filterable: true},
	"resource_id": {Name: "resource_id", Type: query.TypeUint, Filterable: true},
	"created_at":  {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

func parseAuditQuery(t *testing.T, raw string) *query.Query {
	t.Helper()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequestWithContext(context.Background(), "GET", "/?"+raw, nil)
	q, err := query.Parse(c, auditQueryDefs)
	if err != nil {
		t.Fatalf("parse query %q: %v", raw, err)
	}
	return q
}

func TestAuditRepoInsertBatchAndFind(t *testing.T) {
	repo := NewAuditRepo(newAuditTestDB(t))
	ctx := context.Background()

	logs := []domain.AuditLog{
		{UserID: 1, Username: "admin", Action: "login", Resource: "auth", ResourceID: 0, IP: "127.0.0.1", UserAgent: "ua"},
		{UserID: 1, Username: "admin", Action: "create", Resource: "template", ResourceID: 3, IP: "127.0.0.1", UserAgent: "ua"},
		{UserID: 2, Username: "editor", Action: "login", Resource: "auth", ResourceID: 0, IP: "10.0.0.1", UserAgent: "ua"},
	}
	if err := repo.InsertBatch(ctx, logs); err != nil {
		t.Fatal(err)
	}

	// 全量分页
	all, total, err := repo.FindAll(ctx, parseAuditQuery(t, "page=1&page_size=10"))
	if err != nil || total != 3 || len(all) != 3 {
		t.Fatalf("findall: err=%v total=%d n=%d", err, total, len(all))
	}

	// 等值过滤 action=login
	res, total, err := repo.FindAll(ctx, parseAuditQuery(t, `filter={"action":"login"}`))
	if err != nil || total != 2 || len(res) != 2 {
		t.Fatalf("filter action=login: err=%v total=%d %+v", err, total, res)
	}
	for _, r := range res {
		if r.Action != "login" {
			t.Errorf("unexpected action %q", r.Action)
		}
	}

	// 组合过滤 action + user_id
	res, total, _ = repo.FindAll(ctx, parseAuditQuery(t, `filter={"action":"login","user_id":2}`))
	if total != 1 || len(res) != 1 || res[0].Username != "editor" {
		t.Fatalf("combined filter: total=%d %+v", total, res)
	}

	// 关键词搜索 username（Searchable）
	res, total, _ = repo.FindAll(ctx, parseAuditQuery(t, "q=admin"))
	if total != 2 || len(res) != 2 {
		t.Fatalf("keyword admin: total=%d %+v", total, res)
	}

	// 分页：page_size=2 取第一页两条，total 仍为 3
	res, total, _ = repo.FindAll(ctx, parseAuditQuery(t, "page=1&page_size=2"))
	if total != 3 || len(res) != 2 {
		t.Fatalf("pagination: total=%d n=%d", total, len(res))
	}

	// 字段完整回填（含时间戳）
	if all[0].ID == 0 || all[0].CreatedAt.IsZero() {
		t.Errorf("id/created_at not backfilled: %+v", all[0])
	}
}
