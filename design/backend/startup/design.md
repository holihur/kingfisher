# Startup — 启动 & 优雅关闭

## 职责

`cmd/server/main.go` 唯一入口。负责加载 Core、组装 Extends、启动 HTTP Server、优雅关闭。

## main.go 完整流程

```go
func main() {
    // 1. 加载配置
    cfg, err := coreConfig.Load("config/config.yaml")
    if err != nil { log.Fatalf("config load failed: %v", err) }

    // 2. 初始化日志
    logger := coreLogger.New(cfg.Log)
    defer logger.Sync()

    // 3. 初始化数据库
    db, err := coreDB.NewGorm(cfg.MySQL, logger)
    if err != nil { logger.Fatal("mysql init failed", zap.Error(err)) }
    defer closeDB(db)

    // 4. 执行迁移（开发环境自动，生产环境手动）
    if cfg.Server.Mode != "release" {
        coreDB.RunMigrations(db, "migrations")
    }

    // 5. 初始化 Redis
    rdb, err := coreCache.NewRedisClient(cfg.Redis)
    if err != nil { logger.Fatal("redis init failed", zap.Error(err)) }
    defer rdb.Close()

    // 6. 初始化 JWT
    jwtMgr := coreJWT.NewJWTManager(cfg.JWT, coreCache.NewRedisCache(rdb))

    // 7. 初始化 Telemetry（生产环境）
    if cfg.Telemetry.Enabled {
        tp := telemetry.InitTracer(cfg.Telemetry)
        defer tp.Shutdown(context.Background())
    }

    // 8. 组装路由引擎
    r := coreRouter.NewEngine(cfg, logger)

    // 9. 注册 /version 端点（Build Info）
    r.GET("/version", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "version":    version,     // -ldflags -X main.version=$(git describe --tags)
            "commit":     commit,      // -ldflags -X main.commit=$(git rev-parse --short HEAD)
            "build_time": buildTime,   // -ldflags -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)
            "go_version": runtime.Version(),
        })
    })

    // 10. 组装 App（Wire 生成）
    app, err := wire.InitializeApp("config/config.yaml")
    if err != nil { logger.Fatal("wire init failed", zap.Error(err)) }

    // 9. 启动 HTTP Server
    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
        Handler:      app.Engine,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
        IdleTimeout:  120 * time.Second,
    }

    go func() {
        logger.Info("server starting", zap.Int("port", cfg.Server.Port))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("server failed", zap.Error(err))
        }
    }()

    // 10. 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    sig := <-quit
    logger.Info("shutting down...", zap.String("signal", sig.String()))

    // 11. 优雅关闭（30s 超时）
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        logger.Error("forced shutdown", zap.Error(err))
    }

    logger.Info("server stopped")
}
```

## 优雅关闭时序

```
SIGINT / SIGTERM
  │
  ├─→ Shutdown(ctx) begin
  │     ├─→ 停止接收新连接
  │     ├─→ 等待进行中的请求完成（最多 30s）
  │     │     ├─→ HTTP handler 完成
  │     │     ├─→ DB 事务提交/回滚
  │     │     └─→ Redis 操作完成
  │     └─→ 超时 → 强制关闭
  │
  ├─→ closeDB(db)    → GORM 连接池关闭
  ├─→ rdb.Close()    → Redis 连接池关闭
  └─→ tp.Shutdown()  → OTel 导出最后的 spans
```

## 启动配置校验

```go
func (c *Config) Validate() error {
    if c.Server.Port <= 0 || c.Server.Port > 65535 {
        return fmt.Errorf("invalid server port: %d", c.Server.Port)
    }
    if c.JWT.Secret == "" || c.JWT.Secret == "change-me-in-production" {
        return fmt.Errorf("JWT secret must be set via environment variable")
    }
    if c.MySQL.Password == "" {
        return fmt.Errorf("MySQL password must be set via environment variable")
    }
    return nil
}
```

## 资源关闭顺序

```
1. HTTP Server    (先停，不接收新请求)
2. OTel Tracer    (导出剩余 span)
3. Redis          (关闭连接池)
4. MySQL          (关闭连接池)
5. Logger Sync    (刷新缓冲区)
```

## 设计要点

- `defer logger.Sync()` 确保 panic 时日志不丢失
- 优雅关闭超时 30s，覆盖大部分 DB 查询 + HTTP 请求
- K8s `terminationGracePeriodSeconds` 设 ≥ 35s
- 生产环境迁移不自动执行，需要运维手动 `kubectl exec -- ./server migrate`
- 启动失败（DB/Redis 连不上）直接 Fatal，让 K8s/Docker 重启
