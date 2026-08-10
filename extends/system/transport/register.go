package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	rbacTransport "kingfisher/extends/rbac/transport"
	"kingfisher/extends/system/app"
)

// SystemModule 系统信息模块，实现 router.Module
type SystemModule struct {
	handler *SystemHandler
}

// NewSystemModule 创建系统信息模块。ver 为后端构建版本信息（由 main 注入）。
func NewSystemModule(db *gorm.DB, rdb *redis.Client, ver app.VersionInfo) *SystemModule {
	svc := app.NewSystemService(db, rdb, ver)
	return &SystemModule{handler: NewSystemHandler(svc)}
}

func (m *SystemModule) Name() string                       { return "system" }
func (m *SystemModule) Init(ctx context.Context) error     { return nil }
func (m *SystemModule) Shutdown(ctx context.Context) error { return nil }
func (m *SystemModule) RegisterPublic(r *gin.RouterGroup)  {}

func (m *SystemModule) RegisterProtected(r *gin.RouterGroup) {
	sys := r.Group("/system")
	sys.GET("/info", rbacTransport.RequirePerm("system:list"), m.handler.GetInfo)
}
