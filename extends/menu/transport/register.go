package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	adapter "kingfisher/extends/menu/adapter/mysql"
	app "kingfisher/extends/menu/app"
	rbacTransport "kingfisher/extends/rbac/transport"
)

type MenuModule struct{ handler *MenuHandler }

func NewMenuModule(db *gorm.DB, c cache.Cache) *MenuModule {
	repo := adapter.NewMenuRepo(db)
	svc := app.NewMenuService(repo, c)
	return &MenuModule{handler: NewMenuHandler(svc)}
}
func (m *MenuModule) Name() string                       { return "menu" }
func (m *MenuModule) Init(ctx context.Context) error     { return nil }
func (m *MenuModule) Shutdown(ctx context.Context) error { return nil }
func (m *MenuModule) RegisterPublic(r *gin.RouterGroup)  {}
func (m *MenuModule) RegisterProtected(r *gin.RouterGroup) {
	menus := r.Group("/menus")
	menus.GET("/tree", rbacTransport.RequirePerm("menu:list"), m.handler.GetTree)
	menus.GET("/my", m.handler.GetMyTree) // role-filtered, all authenticated users
	menus.GET("/:id", rbacTransport.RequirePerm("menu:list"), m.handler.GetByID)
	menus.POST("", rbacTransport.RequirePerm("menu:create"), m.handler.Create)
	menus.PUT("/:id", rbacTransport.RequirePerm("menu:update"), m.handler.Update)
	menus.POST("/batch-delete", rbacTransport.RequirePerm("menu:delete"), m.handler.BatchDelete)
	menus.POST("/batch-status", rbacTransport.RequirePerm("menu:update"), m.handler.BatchUpdateStatus)
	menus.DELETE("/:id", rbacTransport.RequirePerm("menu:delete"), m.handler.Delete)
}
