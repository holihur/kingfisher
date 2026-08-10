// Package worker 站内信模块的独立任务消费者。
// 只注册本模块的任务类型，不与其他模块的 handler 混用。
package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"

	"kingfisher/extends/message/app"
	messageTask "kingfisher/extends/message/task"
)

// MessageWorker 站内信异步发送的消费者。
type MessageWorker struct {
	svc *app.MessageService
}

// NewMessageWorker 创建站内信 worker。
func NewMessageWorker(svc *app.MessageService) *MessageWorker {
	return &MessageWorker{svc: svc}
}

// Name 模块名。
func (w *MessageWorker) Name() string { return "message" }

// RegisterWorkers 注册本模块的任务 handler。
func (w *MessageWorker) RegisterWorkers(mux *asynq.ServeMux) {
	mux.HandleFunc(messageTask.TypeSendMessage, w.HandleSendMessage)
}

// Shutdown worker 无独立资源，无需清理。
func (w *MessageWorker) Shutdown(ctx context.Context) error { return nil }

// HandleSendMessage 处理站内信异步发送：解析载荷后批量落库。
// 返回普通 error 时 asynq 按 MaxRetry 重试；载荷损坏/无收件人等永久性错误
// 返回 SkipRetry 包装错误，直接归档到死信，避免无效重试。
func (w *MessageWorker) HandleSendMessage(ctx context.Context, t *asynq.Task) error {
	p, err := messageTask.ParseSendMessagePayload(t)
	if err != nil {
		return fmt.Errorf("unmarshal %s payload: %w", t.Type(), asynq.SkipRetry)
	}
	if len(p.RecipientIDs) == 0 {
		return fmt.Errorf("invalid %s payload (empty recipients): %w", t.Type(), asynq.SkipRetry)
	}
	_, err = w.svc.SendBatch(ctx, p.SenderID, p.SenderType, p.RecipientIDs, p.Title, p.Content)
	return err
}
