package adapter

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"kingfisher/core/query"
	"kingfisher/extends/message/domain"
)

func newMessageTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&messagePO{}); err != nil {
		t.Fatal(err)
	}
	return db
}

// messageQueryDefs 与 transport handler 保持一致，adapter 测试本地复刻以避免循环依赖。
var messageQueryDefs = query.Defs{
	"title":      {Name: "title", Type: query.TypeString, Searchable: true, Filterable: true},
	"is_read":    {Name: "is_read", Type: query.TypeBool, Filterable: true},
	"status":     {Name: "status", Type: query.TypeString, Filterable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

// parseMessageQuery 用与 handler 一致的 defs 解析查询字符串。
func parseMessageQuery(t *testing.T, raw string) *query.Query {
	t.Helper()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequestWithContext(context.Background(), "GET", "/?"+raw, nil)
	q, err := query.Parse(c, messageQueryDefs)
	if err != nil {
		t.Fatalf("parse query %q: %v", raw, err)
	}
	return q
}

func seedMessage(t *testing.T, repo *MessageRepo, recipientID uint, title string) uint {
	t.Helper()
	m := &domain.Message{SenderID: 1, SenderType: "admin", RecipientID: recipientID, Title: title, Content: "c", Status: "sent"}
	if err := repo.Create(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	return m.ID
}

// TestMessageRepoScopedByRecipient 验证查询均以 recipient_id 为界，防越权。
// TestMessageRevokedExcludedFromUnread 撤回后：未读数排除、收件箱列表/详情不可见。
func TestMessageRevokedExcludedFromUnread(t *testing.T) {
	repo := NewMessageRepo(newMessageTestDB(t))
	ctx := context.Background()

	id1 := seedMessage(t, repo, 1, "待撤回")
	seedMessage(t, repo, 1, "正常")

	if n, _ := repo.CountUnread(ctx, 1); n != 2 {
		t.Fatalf("初始未读应为 2，got %d", n)
	}

	// 撤回 id1（发送者本人）
	if err := repo.Revoke(ctx, id1, 1); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	// 未读数排除撤回
	if n, _ := repo.CountUnread(ctx, 1); n != 1 {
		t.Fatalf("撤回后未读应为 1，got %d", n)
	}
	// 收件箱列表排除撤回
	items, _, _ := repo.ListByRecipient(ctx, 1, parseMessageQuery(t, "page=1&page_size=10"))
	for _, m := range items {
		if m.ID == id1 {
			t.Fatalf("收件箱不应含已撤回消息 id=%d", id1)
		}
	}
	// 详情对收件人不可见
	if _, err := repo.GetByID(ctx, id1, 1); err == nil {
		t.Fatal("已撤回消息详情应对收件人不可见")
	}
}

func TestMessageRepoScopedByRecipient(t *testing.T) {
	repo := NewMessageRepo(newMessageTestDB(t))
	ctx := context.Background()

	id1 := seedMessage(t, repo, 1, "欢迎")
	id2 := seedMessage(t, repo, 1, "密码重置")
	other := seedMessage(t, repo, 2, "给别人的")

	// CountUnread
	if n, err := repo.CountUnread(ctx, 1); err != nil || n != 2 {
		t.Fatalf("unread: err=%v n=%d", err, n)
	}

	// List 只见自己的
	items, total, err := repo.ListByRecipient(ctx, 1, parseMessageQuery(t, "page=1&page_size=10"))
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("list: err=%v total=%d n=%d", err, total, len(items))
	}

	// 关键词搜索（title 为 Searchable）
	items, _, err = repo.ListByRecipient(ctx, 1, parseMessageQuery(t, "q=重置"))
	if err != nil || len(items) != 1 || items[0].ID != id2 {
		t.Fatalf("keyword list: err=%v %+v", err, items)
	}

	// 等值过滤 is_read
	items, _, _ = repo.ListByRecipient(ctx, 1, parseMessageQuery(t, "filter="+`{"is_read":false}`))
	if len(items) != 2 {
		t.Fatalf("filter unread: %+v", items)
	}

	// GetByID 越权防护
	if _, err := repo.GetByID(ctx, other, 1); err == nil {
		t.Error("should deny accessing other's message")
	}
	if _, err := repo.GetByID(ctx, id1, 2); err == nil {
		t.Error("should deny cross-recipient access")
	}
	got, err := repo.GetByID(ctx, id1, 1)
	if err != nil || got.ID != id1 {
		t.Fatalf("get own: err=%v %+v", err, got)
	}

	// MarkRead（限本人）：未命中自己的消息 = 幂等空更新，返回 nil 且不改动他人消息
	if err := repo.MarkRead(ctx, id1, 2); err != nil {
		t.Fatalf("MarkRead non-owned should be a no-op nil error, got %v", err)
	}
	// 无主命中后，他人的消息仍未读
	otherMsg, _ := repo.GetByID(ctx, other, 2)
	if otherMsg.IsRead {
		t.Error("MarkRead by recipient-1 must not mutate recipient-2's message")
	}
	if err := repo.MarkRead(ctx, id1, 1); err != nil {
		t.Fatalf("markread: %v", err)
	}
	msg, _ := repo.GetByID(ctx, id1, 1)
	if !msg.IsRead || msg.ReadAt == nil {
		t.Errorf("markread not applied: %+v", msg)
	}
	if n, _ := repo.CountUnread(ctx, 1); n != 1 {
		t.Errorf("want 1 unread after markread, got %d", n)
	}

	// DeleteBatch（限本人）
	if err := repo.DeleteBatch(ctx, []uint{id2, other}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByID(ctx, id2, 1); err == nil {
		t.Error("want id2 deleted")
	}
	// other 不属于 recipient 1，应保留
	if _, err := repo.GetByID(ctx, other, 2); err != nil {
		t.Errorf("other's message should remain: %v", err)
	}
	// 越权删除不报错但也不越界删除他人消息
	if err := repo.DeleteBatch(ctx, []uint{other}, 2); err != nil {
		t.Fatalf("delete own other: %v", err)
	}
}
