# M1 — Core 框架启动

## 目标

Go 进程跑起来，`curl localhost:8080/health` 返回 `{"status":"ok"}`。

## 覆盖模块

| 模块 | 设计文档 | 说明 |
|------|----------|------|
| 启动初始化 | [bootstrap](../backend/bootstrap/design.md) | 从零到跑起来的完整流程 |
| 配置 | [config](../backend/config/design.md) | Viper 多环境加载 |
| 日志 | [logger](../backend/logger/design.md) | Zap 结构化日志 |
| 错误码 | [errcode](../backend/errcode/design.md) | 统一错误码体系 |
| 统一响应 | [errcode](../backend/errcode/design.md) | `{code, message, data}` |
| MySQL | [mysql](../backend/mysql/design.md) | GORM 连接池 |
| Redis | [redis](../backend/redis/design.md) | go-redis 连接 |
| JWT | [jwt](../backend/jwt/design.md) | Token 生成/解析 |
| 中间件 | [middleware](../backend/middleware/design.md) | RequestID→Recovery→Trace→Logger→Gzip→SecurityHeaders→CORS→RateLimit（8 个） |
| Core 框架 | [core](../backend/core/design.md) | RouteRegistrar 接口 + Engine 工厂 |
| 可观测 | [observability](../backend/observability/design.md) | Health/Ready 端点 |
| 启动入口 | [startup](../backend/startup/design.md) | main.go + 优雅关闭 |

## 验证

```bash
curl localhost:8080/health   # {"status":"ok"}
curl localhost:8080/ready    # {"status":"ready","mysql":"ok","redis":"ok"}
curl localhost:8080/metrics  # Prometheus metrics
```

## 产出文件

| 文件 | 说明 |
|------|------|
| `go.mod` | 模块初始化 |
| `config/config.yaml` | 默认配置 |
| `.env.example` | 环境变量示例 |
| `core/config/config.go` | 配置加载 + 校验 |
| `core/logger/logger.go` | Zap 封装 |
| `core/errcode/errcode.go` | 错误码 |
| `core/response/response.go` | 统一响应 |
| `core/database/gorm.go` | MySQL 连接 |
| `core/cache/redis.go` | Redis 连接 + Cache 接口 |
| `core/jwt/jwt.go` | JWT Manager |
| `core/middleware/*.go` | 8 个中间件（request_id/trace/recovery/logger/gzip/security_headers/cors/ratelimit） |
| `core/telemetry/*.go` | OTel + Prometheus |
| `core/router/engine.go` | RouteRegistrar 接口 + Engine 工厂 |
| `cmd/server/main.go` | 启动入口 |
## 验收

→ [acceptance 验收清单](../acceptance/design.md)
