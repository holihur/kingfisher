// Package audit implements audit logic.

package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	adapter "kingfisher/extends/audit/adapter/mysql"
	"kingfisher/extends/audit/app"
	"kingfisher/extends/audit/domain"
	rbacTransport "kingfisher/extends/rbac/transport"
)

// AuditRecorder is a callback used by other modules to record audit entries.
type AuditRecorder func(ctx context.Context, userID uint, username, action, resource, result string, ip, userAgent string)

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
	r.GET("/audit-logs", rbacTransport.RequirePerm("audit:list"), m.handler.List)
}
func (m *AuditModule) Middleware() gin.HandlerFunc { return AuditMiddleware(m.svc) }

// Service returns the audit service for query use by other modules.
func (m *AuditModule) Service() *app.AuditService { return m.svc }

// AuditLogCallback returns an AuditRecorder for other modules to use.
func (m *AuditModule) AuditLogCallback() AuditRecorder {
	return func(ctx context.Context, userID uint, username, action, resource, result string, ip, userAgent string) {
		m.svc.Log(ctx, &domain.AuditLog{
			UserID:    userID,
			Username:  username,
			Action:    action,
			Resource:  resource,
			Result:    result,
			IP:        ip,
			UserAgent: userAgent,
		})
	}
}
