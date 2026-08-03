// @title           Kingfisher Admin API
// @version         1.0
// @description     后台管理系统 API 文档
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// Package server implements server logic.

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	"kingfisher/core/config"
	"kingfisher/core/database"
	"kingfisher/core/jwt"
	"kingfisher/core/logger"
	"kingfisher/core/router"

	auditTransport "kingfisher/extends/audit/transport"
	configTransport "kingfisher/extends/config/transport"
	menuTransport "kingfisher/extends/menu/transport"
	rbacTransport "kingfisher/extends/rbac/transport"
	userTransport "kingfisher/extends/user/transport"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	// 1. Load config
	configPath := "config/config.yaml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		configPath = p
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load failed: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize logger
	zapLog, err := logger.New(logger.Config{
		Level: cfg.Log.Level, Format: cfg.Log.Format, Output: cfg.Log.Output,
		FilePath: cfg.Log.FilePath, MaxSize: cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups, MaxAge: cfg.Log.MaxAge,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger init failed: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = zapLog.Sync() }()
	logger.ReplaceGlobals(zapLog)
	zapLog.Info("config loaded", zap.String("mode", cfg.Server.Mode))

	// 3. Initialize database
	db, err := database.InitDatabase(cfg.Database, zapLog)
	if err != nil {
		zapLog.Fatal("database init failed", zap.Error(err))
	}
	defer closeDB(db)

	// Run seed for SQLite
	if cfg.Database.Driver == "sqlite" {
		zapLog.Info("sqlite: seeding...")
		if err := database.SeedData(db); err != nil {
			zapLog.Warn("seed skipped", zap.Error(err))
		} else {
			zapLog.Info("seed data written")
		}
	}

	// 4. Initialize Redis
	rdb, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		// Retry once more then Fatal — Redis is required per design
		time.Sleep(2 * time.Second)
		rdb, err = cache.NewRedisClient(cfg.Redis)
		if err != nil {
			zapLog.Fatal("redis init failed", zap.Error(err))
		}
	}
	defer func() { _ = rdb.Close() }()
	redisCache := cache.NewRedisCache(rdb)

	// 5. Initialize JWT
	jwtMgr := jwt.NewJWTManager(cfg.JWT, redisCache)

	// 6. Build Gin engine
	r := router.NewEngine(cfg, zapLog)

	// 7. Register /version, /health, /ready
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": version, "commit": commit, "build_time": buildTime, "go_version": runtime.Version()})
	})
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/ready", readyHandler(db, rdb))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.StaticFile("/swagger/doc.json", "./docs/swagger.json")
	r.GET("/swagger/*any", swaggerUI())

	// 8. Build auth/rbac middleware
	authMw := rbacTransport.AuthMiddleware(jwtMgr)
	// Create RBAC service for permission lookups — shared between RBAC middleware and User module
	rbacSvc := rbacTransport.NewRoleService(db, redisCache)
	rbacMw := rbacTransport.RBACMiddlewareWith(rbacSvc)

	// 9. Register all extends modules
	auditMod := auditTransport.NewAuditModule(db)
	mods := []router.Module{
		userTransport.NewUserModule(db, redisCache, jwtMgr, rbacSvc.GetUserPermissions),
		rbacTransport.NewRBACModule(db, redisCache),
		menuTransport.NewMenuModule(db, redisCache),
		configTransport.NewConfigModule(db, redisCache),
		auditMod,
	}
	r.Use(auditMod.Middleware()) // audit all write operations

	ctx := context.Background()
	for _, m := range mods {
		if err := m.Init(ctx); err != nil {
			zapLog.Fatal("module init failed", zap.String("module", m.Name()), zap.Error(err))
		}
		router.Register(r, m, authMw, rbacMw)
		zapLog.Info("module registered", zap.String("name", m.Name()))
	}

	// 10. Start HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr: addr, Handler: r,
		ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 120 * time.Second,
	}

	go func() {
		zapLog.Info("server starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLog.Fatal("server failed", zap.Error(err))
		}
	}()

	// 11. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	zapLog.Info("shutting down...", zap.String("signal", sig.String()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLog.Error("forced shutdown", zap.Error(err))
	}

	// Shutdown modules in reverse order
	for i := len(mods) - 1; i >= 0; i-- {
		if err := mods[i].Shutdown(shutdownCtx); err != nil {
			zapLog.Error("module shutdown error", zap.String("module", mods[i].Name()), zap.Error(err))
		}
	}

	zapLog.Info("server stopped")
}

func swaggerUI() gin.HandlerFunc {
	const html = `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>Kingfisher API</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<style>html{box-sizing:border-box}*,*:before,*:after{box-sizing:inherit}body{margin:0}</style></head>
<body><div id="swagger-ui"></div><script>
SwaggerUIBundle({url:"/swagger/doc.json",dom_id:"#swagger-ui",deepLinking:true,presets:[SwaggerUIBundle.presets.apis]})
</script></body></html>`
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, html)
	}
}

func readyHandler(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		dbOK := db.WithContext(ctx).Raw("SELECT 1").Error == nil
		redisOK := rdb != nil && rdb.Ping(ctx).Err() == nil
		s := gin.H{"status": "ready"}
		if dbOK {
			s["mysql"] = "ok"
		} else {
			s["mysql"] = "down"
		}
		if redisOK {
			s["redis"] = "ok"
		} else {
			s["redis"] = "down"
		}
		c.JSON(200, s)
	}
}

func closeDB(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
