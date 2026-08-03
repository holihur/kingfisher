package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	adapter "kingfisher/extends/menu/adapter/mysql"
	app "kingfisher/extends/menu/app"
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
	menus.GET("/tree", m.handler.GetTree)
	menus.GET("/:id", m.handler.GetByID)
	menus.POST("", m.handler.Create)
	menus.PUT("/:id", m.handler.Update)
	menus.DELETE("/:id", m.handler.Delete)
}
