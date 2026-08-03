# Backend Core — 设计与实现差异

> 来源：`design/backend/core/design.md` 对照 `core/`
> 排查日期：2026-07-31

## P0

### C-1 `core/telemetry/` 空目录
  **Status: ✅ telemetry/metrics.go**
- 设计：`core/telemetry/tracer.go` + `metrics.go`（OTel + Prometheus）
- 实现：目录存在但无任何文件；无 `/metrics` 端点、无 OTel span
- 影响：M1 验收 `/metrics` 失败；observability 全部缺失

### C-2 `core/middleware/` 未拆分文件
  **Status: ⚠️ Single file acceptable**
- 设计：8 个独立文件（request_id/trace/recovery/logger/gzip/security_headers/cors/ratelimit）
- 实现：单文件 `middleware.go` 承载全部中间件
- 影响：M1「目录结构与设计一致」验收失败；单一职责弱化

## P1

### C-3 gzip 中间件为空实现
  **Status: ⚠️ gzip empty**
- 设计：JSON 响应 >1KB 压缩，排除 /metrics /health
- 实现：`core/router/engine.go` 的 `gzipMiddleware()` 只调 `c.Next()`，不压缩

### C-4 SecurityHeaders 缺响应头
  **Status: ✅ CSP/Cache/Pragma added**
- 设计：CSP / Cache-Control / Pragma / X-Frame-Options / X-Content-Type-Options / X-XSS-Protection / Referrer-Policy
- 实现：`core/middleware/middleware.go` 仅 X-Frame-Options、X-Content-Type-Options、X-XSS-Protection、Referrer-Policy

### C-5 RateLimit 算法与全局接入
  **Status: ⚠️ Simplified rate limit**
- 设计：ZSET 滑动窗口 + `X-RateLimit-Limit/Remaining/Retry-After` 头；全局限流 `requests_per_minute` 挂 engine
- 实现：INCR 固定窗口（`core/middleware/middleware.go`）；仅登录/注册路由挂载，全局限流未接入
- 影响：突发流量防护弱于设计；`rate_limit.enabled` 配置未生效

### C-6 Trace 无 OTel span
  **Status: ⚠️ OTel deferred**
- 设计：`otel.GetTextMapPropagator().Extract` + `tracer.Start` + `traceparent` 传播
- 实现：仅设置 `X-Trace-ID`（`core/middleware/middleware.go`）

### C-7 Recovery 无请求体限制与堆栈
  **Status: ⚠️ Minor**
- 设计：Recovery 内 `http.MaxBytesReader`（10MB→413）；panic 日志含 `debug.Stack()`
- 实现：无 MaxBytesReader；panic 日志无 stack

## P2

- C-8 `logger.WithContext(ctx)` 未实现（`core/logger/logger.go` 无此函数），设计要求从 ctx 提取 trace_id
- C-9 Cache 接口定义在 `core/cache`（`redis.go`），设计文档要求放在 port 层 + `adapter/redis` 实现——结构差异，功能等价
- C-10 `core/router/module.go` 未单独成文件（设计列出），Module 接口与 engine 混在 `engine.go`
- C-11 `UserPO` 等模型无 `DeletedAt`（`core/database/models.go`），与 M2 软删除设计冲突（见 acceptance A-9）
- C-12 seed 调用位置：设计 `InitDatabase` 内部完成 AutoMigrate+Seed；实现 `main.go` 单独调 `SeedData`——职责轻微偏移
