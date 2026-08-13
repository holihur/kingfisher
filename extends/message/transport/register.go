package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/taskqueue"
	adapter "kingfisher/extends/message/adapter/mysql"
	"kingfisher/extends/message/app"
	messageWorker "kingfisher/extends/message/worker"
	rbacTransport "kingfisher/extends/rbac/transport"
)

type MessageModule struct {
	handler *MessageHandler
	worker  *messageWorker.MessageWorker
}

func NewMessageModule(db *gorm.DB, producer taskqueue.Producer) *MessageModule {
	repo := adapter.NewMessageRepo(db)
	svc := app.NewMessageService(repo)
	return &MessageModule{
		handler: NewMessageHandler(svc, producer),
		worker:  messageWorker.NewMessageWorker(svc),
	}
}

// Worker 注册模式：主程序通过该可选接口收集本站内信模块的独立 worker。
func (m *MessageModule) Worker() taskqueue.WorkerModule { return m.worker }

func (m *MessageModule) Name() string                       { return "message" }
func (m *MessageModule) Init(ctx context.Context) error     { return nil }
func (m *MessageModule) Shutdown(ctx context.Context) error { return nil }
func (m *MessageModule) RegisterPublic(r *gin.RouterGroup)  {}

func (m *MessageModule) RegisterProtected(r *gin.RouterGroup) {
	// 管理员：发送 + 已发送列表 + 撤回
	msgs := r.Group("/messages")
	msgs.POST("", rbacTransport.RequirePerm("message:create"), m.handler.Send)
	msgs.GET("", rbacTransport.RequirePerm("message:list"), m.handler.ListSent)
	msgs.GET("/batch/:batchId", rbacTransport.RequirePerm("message:list"), m.handler.ListBatchMessages)
	msgs.PUT("/batch/:batchId/revoke", rbacTransport.RequirePerm("message:update"), m.handler.RevokeBatch)
	msgs.PUT("/:id/revoke", rbacTransport.RequirePerm("message:update"), m.handler.Revoke)

	// 个人收件箱
	me := r.Group("/me/messages")
	me.GET("", m.handler.List)
	me.GET("/unread-count", m.handler.UnreadCount)
	me.GET("/:id", m.handler.GetByID)
	me.PUT("/:id/read", m.handler.MarkRead)
	me.POST("/batch-delete", m.handler.BatchDelete)
}
