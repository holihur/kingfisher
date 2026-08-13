package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"

	"kingfisher/core/query"
	"kingfisher/extends/message/app"
	"kingfisher/extends/message/domain"
	messageTask "kingfisher/extends/message/task"
)

// fakeMessageRepo 内存实现 port.MessageRepository，测 worker handler 用。
type fakeMessageRepo struct {
	msgs []*domain.Message
}

func (f *fakeMessageRepo) Create(_ context.Context, m *domain.Message) error {
	f.msgs = append(f.msgs, m)
	return nil
}
func (f *fakeMessageRepo) ListByRecipient(ctx context.Context, recipientID uint, q *query.Query) ([]domain.Message, int64, error) {
	return nil, 0, errors.New("not implemented")
}
func (f *fakeMessageRepo) GetByID(ctx context.Context, id, recipientID uint) (*domain.Message, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeMessageRepo) MarkRead(ctx context.Context, id, recipientID uint) error {
	return errors.New("not implemented")
}
func (f *fakeMessageRepo) ListBySender(ctx context.Context, senderID uint, q *query.Query) ([]domain.Message, int64, error) {
	return nil, 0, errors.New("not implemented")
}
func (f *fakeMessageRepo) ListSentBatches(ctx context.Context, senderID uint, q *query.Query) ([]domain.MessageBatch, int64, error) {
	return nil, 0, errors.New("not implemented")
}
func (f *fakeMessageRepo) Revoke(ctx context.Context, id, senderID uint) error {
	return errors.New("not implemented")
}
func (f *fakeMessageRepo) RevokeBatch(ctx context.Context, batchID, senderID uint) error {
	return errors.New("not implemented")
}
func (f *fakeMessageRepo) ListBatchMessages(ctx context.Context, batchID, senderID uint) ([]domain.Message, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeMessageRepo) DeleteBatch(ctx context.Context, ids []uint, recipientID uint) error {
	return errors.New("not implemented")
}
func (f *fakeMessageRepo) CountUnread(ctx context.Context, recipientID uint) (int64, error) {
	return 0, errors.New("not implemented")
}

func newTestWorker() (*MessageWorker, *fakeMessageRepo) {
	r := &fakeMessageRepo{}
	return NewMessageWorker(app.NewMessageService(r)), r
}

func TestHandleSendMessageBatch(t *testing.T) {
	w, repo := newTestWorker()
	task, err := messageTask.NewSendMessageTask(messageTask.SendMessagePayload{
		SenderID: 1, SenderType: "admin", RecipientIDs: []uint{2, 3},
		Title: "通知", Content: "你好",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.HandleSendMessage(context.Background(), task); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(repo.msgs) != 2 {
		t.Fatalf("want 2 messages, got %d", len(repo.msgs))
	}
	for i, want := range []uint{2, 3} {
		m := repo.msgs[i]
		if m.RecipientID != want || m.SenderID != 1 || m.Title != "通知" || m.Status != "sent" {
			t.Errorf("message %d mismatch: %+v", i, m)
		}
	}
}

func TestHandleSendMessageCorruptPayload(t *testing.T) {
	w, _ := newTestWorker()
	task := asynq.NewTask(messageTask.TypeSendMessage, []byte("{not-json"))
	err := w.HandleSendMessage(context.Background(), task)
	if err == nil {
		t.Fatal("want error for corrupt payload")
	}
	if !errors.Is(err, asynq.SkipRetry) {
		t.Errorf("want SkipRetry, got %v", err)
	}
}

func TestHandleSendMessageEmptyRecipients(t *testing.T) {
	w, _ := newTestWorker()
	task, err := messageTask.NewSendMessageTask(messageTask.SendMessagePayload{
		SenderID: 1, SenderType: "admin", RecipientIDs: nil, Title: "x",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = w.HandleSendMessage(context.Background(), task)
	if err == nil {
		t.Fatal("want error for empty recipients")
	}
	if !errors.Is(err, asynq.SkipRetry) {
		t.Errorf("want SkipRetry, got %v", err)
	}
}
