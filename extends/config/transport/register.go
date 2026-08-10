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

type ConfigModule struct {
	handler      *ConfigHandler
	groupHandler *ConfigGroupHandler
	svc          *app.ConfigService
}

func NewConfigModule(db *gorm.DB, c cache.Cache) *ConfigModule {
	repo := adapter.NewConfigRepo(db)
	svc := app.NewConfigService(repo, c)

	groupRepo := adapter.NewConfigGroupRepo(db)
	groupSvc := app.NewConfigGroupService(groupRepo)

	return &ConfigModule{
		handler:      NewConfigHandler(svc),
		groupHandler: NewConfigGroupHandler(groupSvc),
		svc:          svc,
	}
}

// Service 暴露配置服务，供其他模块注入（如注册开关、默认注册角色）。
func (m *ConfigModule) Service() *app.ConfigService        { return m.svc }
func (m *ConfigModule) Name() string                       { return "config" }
func (m *ConfigModule) Init(ctx context.Context) error     { return nil }
func (m *ConfigModule) Shutdown(ctx context.Context) error { return nil }
func (m *ConfigModule) RegisterPublic(r *gin.RouterGroup) {
	// 公开配置：无需登录即可读取 is_public=true 的配置
	pub := r.Group("/public/configs")
	pub.GET("", m.handler.GetPublicAll)
	pub.GET("/:key", m.handler.GetPublic)
}
func (m *ConfigModule) RegisterProtected(r *gin.RouterGroup) {
	configs := r.Group("/configs")
	configs.GET("", rbacTransport.RequirePerm("config:list"), m.handler.GetAll)
	configs.GET("/:key", rbacTransport.RequirePerm("config:list"), m.handler.Get)
	configs.POST("/upload-image", rbacTransport.RequirePerm("config:update"), m.handler.UploadImage)
	configs.PUT("/:key", rbacTransport.RequirePerm("config:update"), m.handler.Set)
	configs.POST("/batch-delete", rbacTransport.RequirePerm("config:update"), m.handler.BatchDelete)
	configs.DELETE("/:key", rbacTransport.RequirePerm("config:update"), m.handler.Delete)

	// 配置分组 CRUD
	groups := r.Group("/config-groups")
	groups.GET("", rbacTransport.RequirePerm("config:list"), m.groupHandler.List)
	groups.POST("", rbacTransport.RequirePerm("config:update"), m.groupHandler.Create)
	groups.PUT("/:id", rbacTransport.RequirePerm("config:update"), m.groupHandler.Update)
	groups.DELETE("/:id", rbacTransport.RequirePerm("config:update"), m.groupHandler.Delete)
}
