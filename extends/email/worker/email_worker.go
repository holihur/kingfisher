// Package worker 邮件模块的异步任务消费者。
package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"kingfisher/core/mailer"
	"kingfisher/core/taskqueue"
	emailTask "kingfisher/extends/email/task"
)

// EmailWorker 邮件异步发送的消费者。
type EmailWorker struct {
	mailer *mailer.Mailer
	log    *zap.Logger
}

// NewEmailWorker 创建邮件 worker。
func NewEmailWorker(m *mailer.Mailer, log *zap.Logger) *EmailWorker {
	return &EmailWorker{mailer: m, log: log}
}

// Name 模块名。
func (w *EmailWorker) Name() string { return "email" }

// RegisterWorkers 注册本模块的任务 handler。
func (w *EmailWorker) RegisterWorkers(mux *asynq.ServeMux) {
	mux.HandleFunc(emailTask.TypeSendEmail, w.HandleSendEmail)
}

// TaskTypes 声明本模块可被周期任务调用的任务类型。
func (w *EmailWorker) TaskTypes() []taskqueue.TaskTypeInfo {
	return []taskqueue.TaskTypeInfo{
		{
			Type:           emailTask.TypeSendEmail,
			Label:          "邮件发送",
			PayloadExample: `{"to":"user@example.com","subject":"标题","body":"正文"}`,
		},
	}
}

// Shutdown worker 无独立资源，无需清理。
func (w *EmailWorker) Shutdown(ctx context.Context) error { return nil }

// HandleSendEmail 处理邮件发送：解析载荷后通过 SMTP 发送。
func (w *EmailWorker) HandleSendEmail(ctx context.Context, t *asynq.Task) error {
	p, err := emailTask.ParseSendEmailPayload(t)
	if err != nil {
		return fmt.Errorf("unmarshal %s payload: %w", t.Type(), asynq.SkipRetry)
	}
	if p.To == "" {
		return fmt.Errorf("invalid %s payload (empty to): %w", t.Type(), asynq.SkipRetry)
	}
	if err := w.mailer.Send(p.To, p.Subject, p.Body); err != nil {
		w.log.Error("邮件发送失败", zap.String("to", p.To), zap.Error(err))
		return err
	}
	w.log.Info("邮件已发送", zap.String("to", p.To), zap.String("subject", p.Subject))
	return nil
}
