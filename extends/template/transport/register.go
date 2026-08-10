package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	rbacTransport "kingfisher/extends/rbac/transport"
	adapter "kingfisher/extends/template/adapter/mysql"
	"kingfisher/extends/template/app"
)

// TemplateModule 模版模块，实现 router.Module 接口
type TemplateModule struct{ handler *TemplateHandler }

func NewTemplateModule(db *gorm.DB, _ cache.Cache) *TemplateModule {
	repo := adapter.NewTemplateRepo(db)
	svc := app.NewTemplateService(repo)
	return &TemplateModule{handler: NewTemplateHandler(svc)}
}

func (m *TemplateModule) Name() string                       { return "template" }
func (m *TemplateModule) Init(ctx context.Context) error     { return nil }
func (m *TemplateModule) Shutdown(ctx context.Context) error { return nil }
func (m *TemplateModule) RegisterPublic(r *gin.RouterGroup)  {}

func (m *TemplateModule) RegisterProtected(r *gin.RouterGroup) {
	tpls := r.Group("/templates")
	tpls.GET("", rbacTransport.RequirePerm("template:list"), m.handler.List)
	tpls.GET("/:id", rbacTransport.RequirePerm("template:list"), m.handler.GetByID)
	tpls.POST("", rbacTransport.RequirePerm("template:create"), m.handler.Create)
	tpls.PUT("/:id", rbacTransport.RequirePerm("template:update"), m.handler.Update)
	tpls.POST("/batch-delete", rbacTransport.RequirePerm("template:delete"), m.handler.BatchDelete)
	tpls.POST("/batch-status", rbacTransport.RequirePerm("template:update"), m.handler.BatchUpdateStatus)
	tpls.DELETE("/:id", rbacTransport.RequirePerm("template:delete"), m.handler.Delete)
}
