// Package config implements config logic.

package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	adapter "kingfisher/extends/config/adapter/mysql"
	"kingfisher/extends/config/app"
)

type ConfigModule struct{ handler *ConfigHandler }

func NewConfigModule(db *gorm.DB) *ConfigModule {
	repo := adapter.NewConfigRepo(db)
	svc := app.NewConfigService(repo)
	return &ConfigModule{handler: NewConfigHandler(svc)}
}
func (m *ConfigModule) Name() string                       { return "config" }
func (m *ConfigModule) Init(ctx context.Context) error     { return nil }
func (m *ConfigModule) Shutdown(ctx context.Context) error { return nil }
func (m *ConfigModule) RegisterPublic(r *gin.RouterGroup)  {}
func (m *ConfigModule) RegisterProtected(r *gin.RouterGroup) {
	configs := r.Group("/configs")
	configs.GET("", m.handler.GetAll)
	configs.GET("/:key", m.handler.Get)
	configs.PUT("/:key", m.handler.Set)
	configs.DELETE("/:key", m.handler.Delete)
}
