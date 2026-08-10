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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	"kingfisher/core/config"
	"kingfisher/core/database"
	"kingfisher/core/jwt"
	"kingfisher/core/logger"
	"kingfisher/core/middleware"
	"kingfisher/core/router"
	"kingfisher/core/taskqueue"
	_ "kingfisher/docs"

	auditTransport "kingfisher/extends/audit/transport"
	configTransport "kingfisher/extends/config/transport"
	dictTransport "kingfisher/extends/dict/transport"
	menuTransport "kingfisher/extends/menu/transport"
	messageTransport "kingfisher/extends/message/transport"
	rbacTransport "kingfisher/extends/rbac/transport"
	taskTransport "kingfisher/extends/task/transport"
	templateTransport "kingfisher/extends/template/transport"
	userAdapter "kingfisher/extends/user/adapter/mysql"
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
	middleware.InitValidator()
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

	// 5.5 Initialize task queue (asynq)：Redis 已必选，producer 一定可用
	asynqOpt := taskqueue.RedisClientOpt(cfg.Redis)
	producer := taskqueue.NewProducer(asynqOpt)
	defer func() {
		if cp, ok := producer.(taskqueue.ClosableProducer); ok {
			_ = cp.Close()
		}
	}()

	// 6. Build Gin engine
	r := router.NewEngine(cfg, zapLog)

	// 7. Register /version, /health, /ready
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": version, "commit": commit, "build_time": buildTime, "go_version": runtime.Version()})
	})
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/ready", readyHandler(db, rdb))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Static("/uploads", "./uploads")

	// Global rate limit — wired with the real Redis cache
	if cfg.RateLimit.Enabled {
		r.Use(middleware.RateLimit(redisCache, cfg.RateLimit.RequestsPerMinute, 1*time.Minute))
	}

	// 8. Build auth/rbac middleware
	svProvider := userAdapter.NewSessionVersionProvider(db)
	authMw := rbacTransport.AuthMiddleware(jwtMgr, svProvider)
	// Create RBAC service for permission lookups — shared between RBAC middleware and User module
	rbacSvc := rbacTransport.NewRoleService(db, redisCache)
	rbacMw := rbacTransport.RBACMiddlewareWith(rbacSvc)

	// 9. Register all extends modules
	auditMod := auditTransport.NewAuditModule(db)
	userMod := userTransport.NewUserModule(db, redisCache, jwtMgr, rbacSvc.GetUserPermissions)
	// Inject audit logger into auth handler so login/logout are recorded
	userMod.InjectAuditLogger(userTransport.AuditLogger(auditMod.AuditLogCallback()))
	userMod.InjectAuditService(auditMod.Service())
	mods := []router.Module{
		userMod,
		rbacTransport.NewRBACModule(db, redisCache),
		menuTransport.NewMenuModule(db, redisCache),
		configTransport.NewConfigModule(db, redisCache),
		dictTransport.NewDictModule(db, redisCache),
		messageTransport.NewMessageModule(db, producer),
		templateTransport.NewTemplateModule(db, redisCache),
		taskTransport.NewTaskModule(db, producer),
		auditMod,
	}
	r.Use(auditMod.Middleware()) // audit all write operations

	// 9.5 收集各模块独立 worker（注册模式：模块实现 WorkerProvider 即注册自己的 worker）
	var workers []taskqueue.WorkerModule
	for _, m := range mods {
		if wp, ok := m.(taskqueue.WorkerProvider); ok {
			workers = append(workers, wp.Worker())
		}
	}
	// 内置 nop worker（周期任务测试/占位）：不依赖任何业务模块
	workers = append(workers, taskqueue.NewNopWorker(zapLog))
	// 注入任务管理页可用的任务类型列表（各模块 worker 声明的类型）
	for _, m := range mods {
		if tm, ok := m.(interface {
			InjectTaskTypes(func() []taskqueue.TaskTypeInfo)
		}); ok {
			allTypes := make([]taskqueue.TaskTypeInfo, 0)
			for _, w := range workers {
				allTypes = append(allTypes, w.TaskTypes()...)
			}
			tm.InjectTaskTypes(func() []taskqueue.TaskTypeInfo { return allTypes })
		}
	}

	// 9.6 收集各模块周期任务 provider（注册模式：模块实现 PeriodicProviderProvider 即注册自己的周期任务）
	var periodicProviders []taskqueue.PeriodicProvider
	for _, m := range mods {
		if pp, ok := m.(taskqueue.PeriodicProviderProvider); ok {
			periodicProviders = append(periodicProviders, pp.PeriodicProvider())
		}
	}
	var taskSrv *taskqueue.Server
	var periodicMgr *taskqueue.PeriodicManager
	if cfg.TaskQueue.Enabled {
		taskSrv = taskqueue.NewServer(asynqOpt, cfg.TaskQueue, workers, zapLog)
		if err := taskSrv.Start(); err != nil {
			zapLog.Fatal("taskqueue server start failed", zap.Error(err))
		}
		zapLog.Info("taskqueue server started", zap.Int("workers", len(workers)))
		if len(periodicProviders) > 0 {
			periodicMgr, err = taskqueue.NewPeriodicManager(asynqOpt, cfg.TaskQueue, periodicProviders, zapLog)
			if err != nil {
				zapLog.Fatal("periodic manager create failed", zap.Error(err))
			}
			if err := periodicMgr.Start(); err != nil {
				zapLog.Fatal("periodic manager start failed", zap.Error(err))
			}
			zapLog.Info("periodic task manager started", zap.Int("providers", len(periodicProviders)))
		}
	}

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

	// Shutdown task queue server（停止消费，等待在途任务）
	if taskSrv != nil {
		if err := taskSrv.Shutdown(shutdownCtx); err != nil {
			zapLog.Error("taskqueue server shutdown error", zap.Error(err))
		}
	}

	// Shutdown periodic task manager（停止周期性调度）
	if periodicMgr != nil {
		if err := periodicMgr.Shutdown(shutdownCtx); err != nil {
			zapLog.Error("periodic manager shutdown error", zap.Error(err))
		}
	}

	// Shutdown modules in reverse order
	for i := len(mods) - 1; i >= 0; i-- {
		if err := mods[i].Shutdown(shutdownCtx); err != nil {
			zapLog.Error("module shutdown error", zap.String("module", mods[i].Name()), zap.Error(err))
		}
	}

	zapLog.Info("server stopped")
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
