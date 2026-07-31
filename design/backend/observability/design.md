# Observability — 可观测性

## 职责

链路追踪（OTel）+ 指标（Prometheus）+ 健康检查。三个一起构成生产环境的可观测性基础。

## 子模块

### 1. 链路追踪（OpenTelemetry）

```go
// pkg/telemetry/tracer.go
func InitTracer(serviceName, endpoint string) (*sdktrace.TracerProvider, error)
```

**Span 覆盖点**：

| 位置 | Span 名 | 属性 |
|------|----------|------|
| HTTP 入口 | `http.{GET} /api/v1/users` | method, path, status_code |
| Service 方法 | `service.User.FindByID` | user_id |
| DB 查询 | `db.query` | db.statement, db.rows |
| Redis 操作 | `redis.{GET}` | db.key |
| 外部调用 | `http.client` | http.url, http.status |

**传播**：通过 `context.Context`，HTTP 头 `traceparent`。

```go
// adapter/mysql/user_repo.go
func (r *UserRepo) FindByID(ctx context.Context, id uint) (*domain.User, error) {
    ctx, span := tracer.Start(ctx, "db.UserRepo.FindByID")
    defer span.End()
    span.SetAttributes(attribute.Int64("user_id", int64(id)))
    // ... DB 操作
}
```

### 2. Prometheus 指标

```go
var (
    HTTPRequestsTotal = prometheus.NewCounterVec(/* method, path, status */)
    HTTPDuration      = prometheus.NewHistogramVec(/* method, path */)
    DBDuration        = prometheus.NewHistogramVec(/* operation, table */)
    CacheHits         = prometheus.NewCounter(/* hit/miss */)
    ActiveConnections = prometheus.NewGauge()
)
```

**暴露端点**：

```
GET /metrics    ← Prometheus 抓取
GET /health     ← K8s liveness probe
GET /ready      ← K8s readiness probe（含 DB/Redis 检测）
```

### 3. 健康检查

```go
// /health — 进程是否存活
func HealthHandler(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) }

// /ready — 是否可接收流量（检查 DB、Redis）
func ReadyHandler(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := db.WithContext(ctx).Raw("SELECT 1").Error; err != nil {
            c.JSON(503, gin.H{"status": "not ready", "mysql": "down"})
            return
        }
        if err := rdb.Ping(ctx).Err(); err != nil {
            c.JSON(503, gin.H{"status": "not ready", "redis": "down"})
            return
        }
        c.JSON(200, gin.H{"status": "ready"})
    }
}
```

## Docker Compose 附加服务

```yaml
jaeger:       # 本地开发用 Jaeger 看 trace
  image: jaegertracing/all-in-one:1.58
  ports: ["16686:16686", "4318:4318"]

prometheus:   # 抓 metrics
  image: prom/prometheus:v2.52
  ports: ["9090:9090"]
  volumes: [./deploy/prometheus.yaml:/etc/prometheus/prometheus.yml]

grafana:      # 可视化
  image: grafana/grafana:11.0
  ports: ["3000:3000"]
```

## 设计要点

- Trace 采样率：开发 100%，生产 10%（可配）
- Gin handler 的 span 在 `middleware/trace.go` 自动创建
- DB/Redis 的 span 在 adapter 层手动创建
- Health/Ready 端点不经过 Auth 中间件
