package transport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"

	"kingfisher/core/query"
	"kingfisher/extends/message/app"
	"kingfisher/extends/message/domain"
	messageTask "kingfisher/extends/message/task"
)

// stubRepo 最小 stub，Send 路径不触达 repo（只入队），其余方法无需实现。
type stubRepo struct{}

func (stubRepo) Create(_ context.Context, m *domain.Message) error { return nil }
func (stubRepo) ListByRecipient(ctx context.Context, recipientID uint, q *query.Query) ([]domain.Message, int64, error) {
	return nil, 0, errors.New("not implemented")
}
func (stubRepo) GetByID(ctx context.Context, id, recipientID uint) (*domain.Message, error) {
	return nil, errors.New("not implemented")
}
func (stubRepo) MarkRead(ctx context.Context, id, recipientID uint) error {
	return errors.New("not implemented")
}
func (stubRepo) ListBySender(ctx context.Context, senderID uint, q *query.Query) ([]domain.Message, int64, error) {
	return nil, 0, errors.New("not implemented")
}
func (stubRepo) ListSentBatches(ctx context.Context, senderID uint, q *query.Query) ([]domain.MessageBatch, int64, error) {
	return nil, 0, errors.New("not implemented")
}
func (stubRepo) RevokeBatch(ctx context.Context, batchID, senderID uint) error {
	return errors.New("not implemented")
}
func (stubRepo) Revoke(ctx context.Context, id, senderID uint) error {
	return errors.New("not implemented")
}
func (stubRepo) ListBatchMessages(ctx context.Context, batchID, senderID uint) ([]domain.Message, error) {
	return nil, errors.New("not implemented")
}
func (stubRepo) DeleteBatch(ctx context.Context, ids []uint, recipientID uint) error {
	return errors.New("not implemented")
}
func (stubRepo) CountUnread(ctx context.Context, recipientID uint) (int64, error) {
	return 0, errors.New("not implemented")
}

// fakeProducer 记录入队的任务，便于断言异步行为。
type fakeProducer struct {
	tasks []*asynq.Task
	err   error
}

func (f *fakeProducer) Enqueue(_ context.Context, task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.tasks = append(f.tasks, task)
	return &asynq.TaskInfo{ID: "test-task"}, nil
}

func newTestHandler(p *fakeProducer) *MessageHandler {
	return NewMessageHandler(app.NewMessageService(stubRepo{}), p)
}

func doSend(t *testing.T, h *MessageHandler, userID uint, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/messages", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", userID)
	if body != "" {
		c.Request.Body = io.NopCloser(strings.NewReader(body))
	}
	h.Send(c)
	return w
}

func TestSendEnqueuesAsync(t *testing.T) {
	p := &fakeProducer{}
	h := newTestHandler(p)
	w := doSend(t, h, 1, `{"recipient_ids":[3,4],"title":"通知","content":"你好"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Enqueued   bool `json:"enqueued"`
			Recipients int  `json:"recipients"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Data.Enqueued || resp.Data.Recipients != 2 {
		t.Errorf("resp=%+v", resp.Data)
	}
	// 入队且未同步写库（producer 收到 1 个 task）
	if len(p.tasks) != 1 {
		t.Fatalf("tasks=%d want 1", len(p.tasks))
	}
	if p.tasks[0].Type() != messageTask.TypeSendMessage {
		t.Errorf("type=%s", p.tasks[0].Type())
	}
	pl, err := messageTask.ParseSendMessagePayload(p.tasks[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(pl.RecipientIDs) != 2 || pl.SenderID != 1 || pl.Title != "通知" {
		t.Errorf("payload=%+v", pl)
	}
}

func TestSendLegacySingleRecipient(t *testing.T) {
	p := &fakeProducer{}
	h := newTestHandler(p)
	w := doSend(t, h, 1, `{"recipient_id":5,"title":"单发","content":""}`)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
	if len(p.tasks) != 1 {
		t.Fatalf("tasks=%d want 1", len(p.tasks))
	}
	pl, _ := messageTask.ParseSendMessagePayload(p.tasks[0])
	if len(pl.RecipientIDs) != 1 || pl.RecipientIDs[0] != 5 {
		t.Errorf("legacy single not wrapped: %+v", pl.RecipientIDs)
	}
}

func TestSendNoRecipients(t *testing.T) {
	p := &fakeProducer{}
	h := newTestHandler(p)
	w := doSend(t, h, 1, `{"title":"无人","content":""}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d want 400", w.Code)
	}
	if len(p.tasks) != 0 {
		t.Error("should not enqueue when no recipients")
	}
}

func TestSendProducerError(t *testing.T) {
	p := &fakeProducer{err: errors.New("redis down")}
	h := newTestHandler(p)
	w := doSend(t, h, 1, `{"recipient_ids":[1],"title":"x"}`)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d want 500", w.Code)
	}
}
