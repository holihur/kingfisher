package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/extends/rbac/transport"
	adapter "kingfisher/extends/worktask/adapter/mysql"
	"kingfisher/extends/worktask/app"
)

type Module struct{ handler *Handler }

func NewModule(db *gorm.DB, resolver ScopeResolver) *Module {
	return &Module{handler: NewHandler(app.NewService(adapter.NewRepository(db)), resolver)}
}

func (m *Module) Name() string                    { return "worktask" }
func (m *Module) Init(context.Context) error      { return nil }
func (m *Module) Shutdown(context.Context) error  { return nil }
func (m *Module) RegisterPublic(*gin.RouterGroup) {}

func (m *Module) RegisterProtected(r *gin.RouterGroup) {
	tasks := r.Group("/tasks")
	tasks.GET("", transport.RequirePerm("worktask:list"), m.handler.List)
	tasks.GET("/:id", transport.RequirePerm("worktask:list"), m.handler.GetByID)
	tasks.POST("", transport.RequirePerm("worktask:create"), m.handler.Create)
	tasks.PUT("/:id", transport.RequirePerm("worktask:update"), m.handler.Update)
	tasks.DELETE("/:id", transport.RequirePerm("worktask:delete"), m.handler.Delete)
}
