package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"kingfisher/core/query"
	"kingfisher/extends/message/domain"
)

var errMsgNotFound = errors.New("message not found")

// mockMessageRepo 内存实现 port.MessageRepository，测 MessageService 用。
type mockMessageRepo struct {
	msgs map[uint]*domain.Message
	seq  uint
}

func (m *mockMessageRepo) ListByRecipient(ctx context.Context, recipientID uint, q *query.Query) ([]domain.Message, int64, error) {
	var out []domain.Message
	for _, v := range m.msgs {
		if v.RecipientID == recipientID {
			out = append(out, *v)
		}
	}
	return out, int64(len(out)), nil
}

func (m *mockMessageRepo) GetByID(ctx context.Context, id, recipientID uint) (*domain.Message, error) {
	v, ok := m.msgs[id]
	if !ok || v.RecipientID != recipientID {
		return nil, errMsgNotFound
	}
	return v, nil
}

func (m *mockMessageRepo) Create(ctx context.Context, msg *domain.Message) error {
	m.seq++
	msg.ID = m.seq
	if m.msgs == nil {
		m.msgs = map[uint]*domain.Message{}
	}
	m.msgs[msg.ID] = msg
	return nil
}

func (m *mockMessageRepo) MarkRead(ctx context.Context, id, recipientID uint) error {
	v, ok := m.msgs[id]
	if !ok || v.RecipientID != recipientID {
		return errMsgNotFound
	}
	v.IsRead = true
	return nil
}

func (m *mockMessageRepo) ListBySender(ctx context.Context, senderID uint, q *query.Query) ([]domain.Message, int64, error) {
	var out []domain.Message
	for _, v := range m.msgs {
		if v.SenderID == senderID {
			out = append(out, *v)
		}
	}
	return out, int64(len(out)), nil
}

func (m *mockMessageRepo) ListSentBatches(ctx context.Context, senderID uint, q *query.Query) ([]domain.MessageBatch, int64, error) {
	// 简单按 batch_id 聚合（测试用）
	byBatch := map[int64]*domain.MessageBatch{}
	for _, v := range m.msgs {
		if v.SenderID != senderID {
			continue
		}
		b := byBatch[v.BatchID]
		if b == nil {
			b = &domain.MessageBatch{BatchID: v.BatchID, Title: v.Title, Content: v.Content, CreatedAt: v.CreatedAt.Format(time.RFC3339)}
			byBatch[v.BatchID] = b
		}
		b.RecipientCount++
		if b.Status != "revoked" && v.Status == "revoked" {
			b.Status = "revoked"
		} else if b.Status == "" {
			b.Status = v.Status
		}
	}
	out := make([]domain.MessageBatch, 0, len(byBatch))
	for _, b := range byBatch {
		out = append(out, *b)
	}
	return out, int64(len(out)), nil
}

func (m *mockMessageRepo) RevokeBatch(ctx context.Context, batchID, senderID uint) error {
	found := false
	for _, v := range m.msgs {
		if uint(v.BatchID) == batchID && v.SenderID == senderID {
			v.Status = "revoked"
			found = true
		}
	}
	if !found {
		return errMsgNotFound
	}
	return nil
}

func (m *mockMessageRepo) Revoke(ctx context.Context, id, senderID uint) error {
	v, ok := m.msgs[id]
	if !ok || v.SenderID != senderID {
		return errMsgNotFound
	}
	v.Status = "revoked"
	return nil
}

func (m *mockMessageRepo) ListBatchMessages(ctx context.Context, batchID, senderID uint) ([]domain.Message, error) {
	var out []domain.Message
	for _, v := range m.msgs {
		if uint(v.BatchID) == batchID && v.SenderID == senderID {
			out = append(out, *v)
		}
	}
	return out, nil
}

func (m *mockMessageRepo) Delete(ctx context.Context, id, recipientID uint) error {
	v, ok := m.msgs[id]
	if !ok || v.RecipientID != recipientID {
		return errMsgNotFound
	}
	delete(m.msgs, id)
	return nil
}

func (m *mockMessageRepo) DeleteBatch(ctx context.Context, ids []uint, recipientID uint) error {
	for _, id := range ids {
		if v, ok := m.msgs[id]; ok && v.RecipientID == recipientID {
			delete(m.msgs, id)
		}
	}
	return nil
}

func (m *mockMessageRepo) CountUnread(ctx context.Context, recipientID uint) (int64, error) {
	var n int64
	for _, v := range m.msgs {
		if v.RecipientID == recipientID && !v.IsRead {
			n++
		}
	}
	return n, nil
}

func TestMessageServiceCreateDefaultsAndScope(t *testing.T) {
	svc := NewMessageService(&mockMessageRepo{})
	ctx := context.Background()

	// Create 默认值
	msg, err := svc.Create(ctx, 1, "admin", 2, "欢迎", "你好")
	if err != nil || msg.ID == 0 || msg.Status != "sent" || msg.IsRead {
		t.Fatalf("create defaults: err=%v %+v", err, msg)
	}

	// 越权读取
	if _, err := svc.GetByID(ctx, msg.ID, 3); err == nil {
		t.Error("should deny read by non-recipient")
	}
	// 正常读取
	got, err := svc.GetByID(ctx, msg.ID, 2)
	if err != nil || got.Title != "欢迎" {
		t.Fatalf("read own: err=%v %+v", err, got)
	}

	// List + UnreadCount
	if _, total, _ := svc.List(ctx, 2, &query.Query{Page: 1, PageSize: 20}); total != 1 {
		t.Errorf("list total=%d", total)
	}
	if n, _ := svc.UnreadCount(ctx, 2); n != 1 {
		t.Errorf("unread=%d", n)
	}

	// MarkRead + 未读数归零
	if err := svc.MarkRead(ctx, msg.ID, 2); err != nil {
		t.Fatal(err)
	}
	if n, _ := svc.UnreadCount(ctx, 2); n != 0 {
		t.Errorf("unread after read=%d", n)
	}

	// 越权 MarkRead 失败
	if err := svc.MarkRead(ctx, msg.ID, 3); err == nil {
		t.Error("sent")
	}

	// DeleteBatch + 删除后列表空
	if err := svc.DeleteBatch(ctx, []uint{msg.ID}, 2); err != nil {
		t.Fatal(err)
	}
	if _, total, _ := svc.List(ctx, 2, &query.Query{Page: 1, PageSize: 20}); total != 0 {
		t.Errorf("total after delete=%d", total)
	}
	// List 隔离：给 user3 的未读数与 user2 无关
	for range 3 {
		if _, err := svc.Create(ctx, 1, "admin", uint(3), "批量", ""); err != nil {
			t.Fatal(err)
		}
	}
	if n, _ := svc.UnreadCount(ctx, 2); n != 0 {
		t.Errorf("recipient 2 unread should stay 0, got %d", n)
	}
}
