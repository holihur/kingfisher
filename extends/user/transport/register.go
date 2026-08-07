// Package user implements user logic.

package transport

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	"kingfisher/core/jwt"
	"kingfisher/core/middleware"
	auditApp "kingfisher/extends/audit/app"
	rbacTransport "kingfisher/extends/rbac/transport"
	adapter "kingfisher/extends/user/adapter/mysql"
	"kingfisher/extends/user/app"
)

type UserModule struct {
	authHandler *AuthHandler
	userHandler *UserHandler
	cache       cache.Cache
	authSvc     *app.AuthService
}

func NewUserModule(db *gorm.DB, c cache.Cache, jwtMgr *jwt.JWTManager, getUserPerms PermProvider) *UserModule {
	repo := adapter.NewUserRepo(db)
	authSvc := app.NewAuthService(repo, c, jwtMgr)
	userSvc := app.NewUserService(repo, c)
	uh := NewUserHandler(userSvc)
	uh.getUserPerms = getUserPerms
	return &UserModule{
		authHandler: NewAuthHandler(authSvc),
		userHandler: uh,
		cache:       c,
		authSvc:     authSvc,
	}
}

// InjectLandingPageProvider 注入角色落地页查询函数（登录后跳转页面）。
func (m *UserModule) InjectLandingPageProvider(fn func(ctx context.Context, roleID uint) (string, error)) {
	m.authSvc.SetLandingPageProvider(fn)
}

// InjectConfigProvider 注入系统配置查询函数（注册开关、默认注册角色）。
func (m *UserModule) InjectConfigProvider(fn func(ctx context.Context, key string) (string, error)) {
	m.authSvc.SetConfigProvider(fn)
}

// InjectAuditLogger wires the audit logger callback into the auth handler.
func (m *UserModule) InjectAuditLogger(fn AuditLogger) { m.authHandler.auditLog = fn }

// InjectAuditService wires the audit service into the user handler for login-log queries.
func (m *UserModule) InjectAuditService(auditSvc *auditApp.AuditService) {
	m.userHandler.SetAuditService(auditSvc)
}

func (m *UserModule) Name() string                       { return "user" }
func (m *UserModule) Init(ctx context.Context) error     { return nil }
func (m *UserModule) Shutdown(ctx context.Context) error { return nil }

func (m *UserModule) RegisterPublic(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.POST("/register", m.authHandler.Register)
	auth.POST("/login", middleware.RateLimit(m.cache, 5, time.Minute), m.authHandler.Login)
	auth.POST("/refresh", m.authHandler.Refresh)
	auth.POST("/logout", m.authHandler.Logout)
}

func (m *UserModule) RegisterProtected(r *gin.RouterGroup) {
	users := r.Group("/users")
	users.POST("", rbacTransport.RequirePerm("user:create"), m.userHandler.Create)
	users.GET("/me", m.userHandler.GetMe)
	users.PUT("/me", m.userHandler.UpdateMe)
	users.GET("/me/permissions", m.userHandler.GetMyPermissions)
	users.GET("/me/login-logs", m.userHandler.GetMyLoginLogs)
	users.POST("/me/avatar", m.userHandler.UploadAvatar)
	users.PUT("/me/password", m.userHandler.ChangePassword)
	users.GET("/:id", rbacTransport.RequirePerm("user:list"), m.userHandler.GetByID)
	users.PUT("/:id", rbacTransport.RequirePerm("user:update"), m.userHandler.Update)
	users.GET("", rbacTransport.RequirePerm("user:list"), m.userHandler.List)
	users.POST("/batch-delete", rbacTransport.RequirePerm("user:delete"), m.userHandler.BatchDelete)
	users.POST("/batch-status", rbacTransport.RequirePerm("user:update"), m.userHandler.BatchUpdateStatus)
	users.DELETE("/:id", rbacTransport.RequirePerm("user:delete"), m.userHandler.Delete)
	users.DELETE("/:id/sessions", rbacTransport.RequirePerm("user:update"), m.userHandler.RevokeSessions)
}

var _ = gin.HandlerFunc(nil)
