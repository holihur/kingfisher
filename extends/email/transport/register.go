// Package transport 邮件模块的注册与任务接线。
package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"kingfisher/core/mailer"
	"kingfisher/core/taskqueue"
	"kingfisher/extends/email/task"
	"kingfisher/extends/email/worker"
)

// EmailModule 邮件模块：仅注册异步 worker，无 HTTP 路由。
type EmailModule struct {
	mailer *mailer.Mailer
	worker *worker.EmailWorker
}

// NewEmailModule 创建邮件模块。
func NewEmailModule(m *mailer.Mailer, log *zap.Logger) *EmailModule {
	return &EmailModule{
		mailer: m,
		worker: worker.NewEmailWorker(m, log),
	}
}

// Worker 实现 taskqueue.WorkerProvider，供 main 收集注册。
func (m *EmailModule) Worker() taskqueue.WorkerModule { return m.worker }

// ---- router.Module 实现（无 HTTP 路由，空实现）----

func (m *EmailModule) Name() string                         { return "email" }
func (m *EmailModule) Init(ctx context.Context) error       { return nil }
func (m *EmailModule) Shutdown(ctx context.Context) error   { return nil }
func (m *EmailModule) RegisterPublic(r *gin.RouterGroup)    {}
func (m *EmailModule) RegisterProtected(r *gin.RouterGroup) {}

// Producer 供其他模块入队发送邮件。
type Producer interface {
	EnqueueEmail(ctx context.Context, to, subject, body string) error
}

// EmailProducer 将邮件入队到 asynq。
type EmailProducer struct {
	producer taskqueue.Producer
}

// NewEmailProducer 创建邮件入队器。
func NewEmailProducer(p taskqueue.Producer) *EmailProducer {
	return &EmailProducer{producer: p}
}

// EnqueueEmail 入队一封邮件（异步发送）。
func (e *EmailProducer) EnqueueEmail(ctx context.Context, to, subject, body string) error {
	t, err := task.NewSendEmailTask(task.SendEmailPayload{To: to, Subject: subject, Body: body})
	if err != nil {
		return err
	}
	_, err = e.producer.Enqueue(ctx, t)
	return err
}

var _ = asynq.TaskInfo{} // 保持 asynq 导入（Enqueue 返回类型）
