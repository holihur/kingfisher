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
	rbacTransport "kingfisher/extends/rbac/transport"
	adapter "kingfisher/extends/user/adapter/mysql"
	"kingfisher/extends/user/app"
)

type UserModule struct {
	authHandler *AuthHandler
	userHandler *UserHandler
	cache       cache.Cache
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
	}
}

// InjectAuditLogger wires the audit logger callback into the auth handler.
func (m *UserModule) InjectAuditLogger(fn AuditLogger) { m.authHandler.auditLog = fn }

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
	users.GET("/me/permissions", m.userHandler.GetMyPermissions)
	users.PUT("/me/password", m.userHandler.ChangePassword)
	users.GET("/:id", m.userHandler.GetByID)
	users.PUT("/:id", rbacTransport.RequirePerm("user:update"), m.userHandler.Update)
	users.GET("", rbacTransport.RequirePerm("user:list"), m.userHandler.List)
	users.DELETE("/:id", rbacTransport.RequirePerm("user:delete"), m.userHandler.Delete)
	users.DELETE("/:id/sessions", rbacTransport.RequirePerm("user:update"), m.userHandler.RevokeSessions)
}

var _ = gin.HandlerFunc(nil)
