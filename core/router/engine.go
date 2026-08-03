// Package router implements router logic.

package router

import (
	"context"
	"time"

	"github.com/gin-contrib/gzip"
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

	if len(cfg.Server.TrustedProxies) > 0 {
		_ = r.SetTrustedProxies(cfg.Server.TrustedProxies)
	}

	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery())
	r.Use(middleware.Trace())
	r.Use(middleware.Logger(logger))
	r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/metrics", "/health", "/ready"})))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	if cfg.RateLimit.Enabled {
		r.Use(middleware.RateLimit(nil, cfg.RateLimit.RequestsPerMinute, 1*time.Minute))
	}

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
