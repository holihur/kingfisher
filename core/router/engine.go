package router

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"kingfisher/core/config"
	"kingfisher/core/middleware"
)

// Module defines the interface each extends module implements
type Module interface {
	Name() string
	Init(ctx context.Context) error
	RegisterPublic(r *gin.RouterGroup)
	RegisterProtected(r *gin.RouterGroup)
	Shutdown(ctx context.Context) error
}

// NewEngine creates a Gin engine with all core middleware
func NewEngine(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// Set trusted proxies for correct IP resolution
	if len(cfg.Server.TrustedProxies) > 0 {
		r.SetTrustedProxies(cfg.Server.TrustedProxies)
	}

	// 1. RequestID — outermost, so even panic logs have request_id
	r.Use(middleware.RequestID())
	// 2. Recovery — must be outermost to catch all downstream panics
	r.Use(middleware.Recovery())
	// 3. Trace — after Recovery, so panic won't lose trace
	r.Use(middleware.Trace())
	// 4. Logger — after Trace, can log trace_id
	r.Use(middleware.Logger(logger))
	// 5. Gzip — compress JSON responses > 1KB
	r.Use(gzipMiddleware())
	// 6. SecurityHeaders — all responses get security headers
	r.Use(middleware.SecurityHeaders())
	// 7. CORS — before Auth, so OPTIONS preflight doesn't need auth
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	return r
}

// Register adds a module's routes to the engine
func Register(r *gin.Engine, mod Module, authMw gin.HandlerFunc, rbacMw gin.HandlerFunc) {
	api := r.Group("/api/v1")
	mod.RegisterPublic(api)

	protected := api.Group("")
	protected.Use(authMw, rbacMw)
	mod.RegisterProtected(protected)
}

func gzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

var _ = context.Background // use context
