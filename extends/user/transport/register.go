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
	authHandler    *AuthHandler
	userHandler    *UserHandler
	userSvc        *app.UserService
	mfaSvc         *app.MFAService
	cache          cache.Cache
	authSvc        *app.AuthService
	loginPerMinute int
}

func NewUserModule(db *gorm.DB, c cache.Cache, jwtMgr *jwt.JWTManager, getUserPerms PermProvider, loginPerMinute int) *UserModule {
	repo := adapter.NewUserRepo(db)
	authSvc := app.NewAuthService(repo, c, jwtMgr)
	userSvc := app.NewUserService(repo, c)
	mfaSvc := app.NewMFAService(repo, c)
	authSvc.SetMFAService(mfaSvc)
	userSvc.SetPermProvider(getUserPerms)
	uh := NewUserHandler(userSvc)
	uh.getUserPerms = getUserPerms
	uh.SetMFAService(mfaSvc)
	ah := NewAuthHandler(authSvc)
	ah.SetMFAService(mfaSvc)
	if loginPerMinute <= 0 {
		loginPerMinute = 5
	}
	return &UserModule{
		authHandler:    ah,
		userHandler:    uh,
		userSvc:        userSvc,
		mfaSvc:         mfaSvc,
		cache:          c,
		authSvc:        authSvc,
		loginPerMinute: loginPerMinute,
	}
}

func (m *UserModule) InjectPermProvider(fn func(ctx context.Context, userID uint) ([]string, error)) {
	if m.userSvc != nil {
		m.userSvc.SetPermProvider(fn)
	}
}

// InjectLandingPageProvider 注入角色落地页查询函数（登录后跳转页面）。
func (m *UserModule) InjectLandingPageProvider(fn func(ctx context.Context, roleID uint) (string, error)) {
	m.authSvc.SetLandingPageProvider(fn)
}

func (m *UserModule) InjectConfigProvider(fn func(ctx context.Context, key string) (string, error)) {
	m.authSvc.SetConfigProvider(fn)
	if m.mfaSvc != nil {
		m.mfaSvc.SetConfigProvider(fn)
	}
}

func (m *UserModule) InjectEmailSender(fn func(ctx context.Context, to, subject, body string) error) {
	m.authSvc.SetEmailSender(fn)
	if m.mfaSvc != nil {
		m.mfaSvc.SetEmailSender(fn)
	}
}

// InjectTemplateRenderer 注入模板渲染函数（找回密码邮件按模板渲染）。
func (m *UserModule) InjectTemplateRenderer(fn func(ctx context.Context, code string, vars map[string]string) (subject, body string, err error)) {
	m.authSvc.SetTemplateRenderer(fn)
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
	auth.POST("/login", middleware.RateLimit(m.cache, m.loginPerMinute, time.Minute), m.authHandler.Login)
	auth.POST("/mfa/verify", m.authHandler.MFAVerify)
	auth.POST("/mfa/send", m.authHandler.MFASend)
	auth.POST("/refresh", m.authHandler.Refresh)
	auth.POST("/logout", m.authHandler.Logout)
	auth.POST("/forgot-password", middleware.RateLimit(m.cache, 5, time.Minute), m.authHandler.ForgotPassword)
	auth.POST("/reset-password", m.authHandler.ResetPassword)
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
	users.GET("/me/mfa/status", m.userHandler.GetMFAStatus)
	users.POST("/me/mfa/totp/setup", m.userHandler.SetupTOTP)
	users.POST("/me/mfa/totp/verify", m.userHandler.VerifyTOTP)
	users.DELETE("/me/mfa/totp", m.userHandler.DisableTOTP)
	users.POST("/me/mfa/sms/send", m.userHandler.SendSMSForMFA)
	users.POST("/me/mfa/sms/verify", m.userHandler.VerifySMS)
	users.DELETE("/me/mfa/sms", m.userHandler.DisableSMS)
	users.POST("/me/mfa/email/send", m.userHandler.SendEmailForMFA)
	users.POST("/me/mfa/email/verify", m.userHandler.VerifyEmail)
	users.DELETE("/me/mfa/email", m.userHandler.DisableEmail)
	users.POST("/me/sub-accounts", m.userHandler.CreateSubAccount)
	users.GET("/me/sub-accounts", m.userHandler.ListSubAccounts)
	users.PUT("/me/sub-accounts/:id", m.userHandler.UpdateSubAccount)
	users.DELETE("/me/sub-accounts/:id", m.userHandler.DeleteSubAccount)
	users.GET("/sub-accounts", rbacTransport.RequirePerm("user:list"), m.userHandler.AdminListSubAccounts)
	users.GET("/:id/mfa/status", rbacTransport.RequirePerm("user:list"), m.userHandler.AdminGetMFAStatus)
	users.DELETE("/:id/mfa/reset", rbacTransport.RequirePerm("user:update"), m.userHandler.AdminResetMFA)
	users.GET("/:id", rbacTransport.RequirePerm("user:list"), m.userHandler.GetByID)
	users.PUT("/:id", rbacTransport.RequirePerm("user:update"), m.userHandler.Update)
	users.GET("", rbacTransport.RequirePerm("user:list"), m.userHandler.List)
	users.POST("/batch-delete", rbacTransport.RequirePerm("user:delete"), m.userHandler.BatchDelete)
	users.POST("/batch-status", rbacTransport.RequirePerm("user:update"), m.userHandler.BatchUpdateStatus)
	users.DELETE("/:id", rbacTransport.RequirePerm("user:delete"), m.userHandler.Delete)
	users.DELETE("/:id/sessions", rbacTransport.RequirePerm("user:update"), m.userHandler.RevokeSessions)
}

var _ = gin.HandlerFunc(nil)
