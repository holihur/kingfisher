package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	adapter "kingfisher/extends/config/adapter/mysql"
	"kingfisher/extends/config/app"
	rbacTransport "kingfisher/extends/rbac/transport"
)

type ConfigModule struct{ handler *ConfigHandler }

func NewConfigModule(db *gorm.DB, c cache.Cache) *ConfigModule {
	repo := adapter.NewConfigRepo(db)
	svc := app.NewConfigService(repo, c)
	return &ConfigModule{handler: NewConfigHandler(svc)}
}
func (m *ConfigModule) Name() string                       { return "config" }
func (m *ConfigModule) Init(ctx context.Context) error     { return nil }
func (m *ConfigModule) Shutdown(ctx context.Context) error { return nil }
func (m *ConfigModule) RegisterPublic(r *gin.RouterGroup)  {}
func (m *ConfigModule) RegisterProtected(r *gin.RouterGroup) {
	configs := r.Group("/configs")
	configs.GET("", rbacTransport.RequirePerm("config:list"), m.handler.GetAll)
	configs.GET("/:key", rbacTransport.RequirePerm("config:list"), m.handler.Get)
	configs.PUT("/:key", rbacTransport.RequirePerm("config:update"), m.handler.Set)
	configs.DELETE("/:key", rbacTransport.RequirePerm("config:update"), m.handler.Delete)
}
