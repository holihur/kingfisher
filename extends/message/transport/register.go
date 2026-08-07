package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	adapter "kingfisher/extends/message/adapter/mysql"
	"kingfisher/extends/message/app"
	rbacTransport "kingfisher/extends/rbac/transport"
)

type MessageModule struct{ handler *MessageHandler }

func NewMessageModule(db *gorm.DB, _ cache.Cache) *MessageModule {
	repo := adapter.NewMessageRepo(db)
	svc := app.NewMessageService(repo)
	return &MessageModule{handler: NewMessageHandler(svc)}
}

func (m *MessageModule) Name() string                       { return "message" }
func (m *MessageModule) Init(ctx context.Context) error     { return nil }
func (m *MessageModule) Shutdown(ctx context.Context) error { return nil }
func (m *MessageModule) RegisterPublic(r *gin.RouterGroup)  {}

func (m *MessageModule) RegisterProtected(r *gin.RouterGroup) {
	// 管理员发送
	msgs := r.Group("/messages")
	msgs.POST("", rbacTransport.RequirePerm("message:create"), m.handler.Send)

	// 个人收件箱
	me := r.Group("/me/messages")
	me.GET("", m.handler.List)
	me.GET("/unread-count", m.handler.UnreadCount)
	me.GET("/:id", m.handler.GetByID)
	me.PUT("/:id/read", m.handler.MarkRead)
	me.POST("/batch-delete", m.handler.BatchDelete)
}
