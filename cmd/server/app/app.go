package app

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
	"kingfisher/core/middleware"
	"kingfisher/core/router"
	"kingfisher/core/taskqueue"
	_ "kingfisher/docs"

	rbacTransport "kingfisher/extends/rbac/transport"
	userAdapter "kingfisher/extends/user/adapter/mysql"
)

func Run(cfg *config.Config, zapLog *zap.Logger, vi VersionInfo) error {
	middleware.InitValidator()

	db, err := database.InitDatabase(cfg.Database, zapLog)
	if err != nil {
		return fmt.Errorf("database init: %w", err)
	}
	defer closeDB(db)

	if cfg.Database.Driver == "sqlite" {
		zapLog.Info("sqlite: seeding...")
		if err := database.SeedData(db); err != nil {
			zapLog.Warn("seed skipped", zap.Error(err))
		} else {
			zapLog.Info("seed data written")
		}
	}

	rdb, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		time.Sleep(2 * time.Second)
		rdb, err = cache.NewRedisClient(cfg.Redis)
		if err != nil {
			return fmt.Errorf("redis init: %w", err)
		}
	}
	defer func() { _ = rdb.Close() }()
	redisCache := cache.NewRedisCache(rdb)

	jwtMgr := jwt.NewJWTManager(cfg.JWT, redisCache)
	asynqOpt := taskqueue.RedisClientOpt(cfg.Redis)
	producer := taskqueue.NewProducer(asynqOpt)
	defer func() {
		if cp, ok := producer.(taskqueue.ClosableProducer); ok {
			_ = cp.Close()
		}
	}()

	r := router.NewEngine(cfg, zapLog)
	setupInfraRoutes(r, db, rdb, vi)
	setupSPA(r, cfg)
	if cfg.RateLimit.Enabled {
		r.Use(middleware.RateLimit(redisCache, cfg.RateLimit.RequestsPerMinute, 1*time.Minute))
	}

	svProvider := userAdapter.NewSessionVersionProvider(db)
	authMw := rbacTransport.AuthMiddleware(jwtMgr, svProvider)
	bundle := buildModules(db, rdb, redisCache, jwtMgr, producer, cfg, zapLog, vi)
	rbacMw := rbacTransport.RBACMiddlewareWith(bundle.rbacSvc)
	r.Use(bundle.auditMod.Middleware())

	workers, taskSrv, periodicMgr := setupTaskQueue(bundle.mods, cfg, asynqOpt, zapLog)
	_ = workers
	_ = taskSrv
	_ = periodicMgr

	ctx := context.Background()
	for _, m := range bundle.mods {
		if err := m.Init(ctx); err != nil {
			return fmt.Errorf("module %s init: %w", m.Name(), err)
		}
		router.Register(r, m, authMw, rbacMw)
		zapLog.Info("module registered", zap.String("name", m.Name()))
	}
	if taskSrv != nil {
		if err := taskSrv.Start(); err != nil {
			return fmt.Errorf("taskqueue start: %w", err)
		}
		zapLog.Info("taskqueue server started")
		if periodicMgr != nil {
			if err := periodicMgr.Start(); err != nil {
				return fmt.Errorf("periodic manager start: %w", err)
			}
			zapLog.Info("periodic task manager started")
		}
	}

	return serveAndWait(r, cfg, zapLog, bundle.mods, taskSrv, periodicMgr)
}

func setupInfraRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, vi VersionInfo) {
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": vi.Version, "commit": vi.Commit, "build_time": vi.BuildTime, "go_version": runtime.Version()})
	})
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/ready", readyHandler(db, rdb))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Static("/uploads", "./uploads")
}

func serveAndWait(r *gin.Engine, cfg *config.Config, zapLog *zap.Logger, mods []router.Module, taskSrv *taskqueue.Server, periodicMgr *taskqueue.PeriodicManager) error {
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr: addr, Handler: r,
		ReadTimeout: 10 * time.Second, WriteTimeout: 300 * time.Second, IdleTimeout: 120 * time.Second,
	}
	go func() {
		zapLog.Info("server starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLog.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	zapLog.Info("shutting down...", zap.String("signal", sig.String()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLog.Error("forced shutdown", zap.Error(err))
	}
	if taskSrv != nil {
		if err := taskSrv.Shutdown(shutdownCtx); err != nil {
			zapLog.Error("taskqueue server shutdown error", zap.Error(err))
		}
	}
	if periodicMgr != nil {
		if err := periodicMgr.Shutdown(shutdownCtx); err != nil {
			zapLog.Error("periodic manager shutdown error", zap.Error(err))
		}
	}
	for i := len(mods) - 1; i >= 0; i-- {
		if err := mods[i].Shutdown(shutdownCtx); err != nil {
			zapLog.Error("module shutdown error", zap.String("module", mods[i].Name()), zap.Error(err))
		}
	}
	zapLog.Info("server stopped")
	return nil
}

func setupTaskQueue(mods []router.Module, cfg *config.Config, asynqOpt interface{}, zapLog *zap.Logger) ([]taskqueue.WorkerModule, *taskqueue.Server, *taskqueue.PeriodicManager) {
	var workers []taskqueue.WorkerModule
	for _, m := range mods {
		if wp, ok := m.(taskqueue.WorkerProvider); ok {
			workers = append(workers, wp.Worker())
		}
	}
	workers = append(workers, taskqueue.NewNopWorker(zapLog))
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
	var periodicProviders []taskqueue.PeriodicProvider
	for _, m := range mods {
		if pp, ok := m.(taskqueue.PeriodicProviderProvider); ok {
			periodicProviders = append(periodicProviders, pp.PeriodicProvider())
		}
	}
	if !cfg.TaskQueue.Enabled {
		return workers, nil, nil
	}
	// asynqOpt is taskqueue.RedisClientOpt result; type assert to expected opt
	// To avoid import cycle, we accept interface{} and cast via type switch using the known concrete type.
	// Fallback: reconstruct opt from cfg if cast fails (should not happen in production).
	opt := asynqOpt
	_ = opt
	// Direct construction using typed helper to satisfy compiler: we know asynqOpt is asynq.RedisClientOpt
	// So we reconstruct via taskqueue.RedisClientOpt(cfg.Redis) with correct type.
	// This keeps app.go decoupled from asynq concrete type.
	srvOpt := taskqueue.RedisClientOpt(cfg.Redis)
	taskSrv := taskqueue.NewServer(srvOpt, cfg.TaskQueue, workers, zapLog)
	var periodicMgr *taskqueue.PeriodicManager
	if len(periodicProviders) > 0 {
		var err error
		periodicMgr, err = taskqueue.NewPeriodicManager(srvOpt, cfg.TaskQueue, periodicProviders, zapLog)
		if err != nil {
			zapLog.Fatal("periodic manager create failed", zap.Error(err))
		}
	}
	return workers, taskSrv, periodicMgr
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
