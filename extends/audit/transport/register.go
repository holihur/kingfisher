// Package audit implements audit logic.

package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	adapter "kingfisher/extends/audit/adapter/mysql"
	"kingfisher/extends/audit/app"
)

type AuditModule struct {
	handler *AuditHandler
	svc     *app.AuditService
}

func NewAuditModule(db *gorm.DB) *AuditModule {
	repo := adapter.NewAuditRepo(db)
	svc := app.NewAuditService(repo)
	return &AuditModule{handler: NewAuditHandler(svc), svc: svc}
}
func (m *AuditModule) Name() string                       { return "audit" }
func (m *AuditModule) Init(ctx context.Context) error     { return nil }
func (m *AuditModule) Shutdown(ctx context.Context) error { return nil }
func (m *AuditModule) RegisterPublic(r *gin.RouterGroup)  {}
func (m *AuditModule) RegisterProtected(r *gin.RouterGroup) {
	r.GET("/audit-logs", m.handler.List)
}
