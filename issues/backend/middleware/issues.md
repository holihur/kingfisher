# Middleware — 设计与实现差异

> 来源：`design/backend/middleware/design.md` 对照 `core/middleware/middleware.go`、`core/router/engine.go`
> 排查日期：2026-08-03

## P1

### MW-1 ✅ 中间件未拆分文件
- 设计：8 个独立文件（request_id/recovery/trace/logger/gzip/cors/security_headers/ratelimit）+ extends/rbac 3 个文件（auth_middleware/rbac_middleware/require_perm）
- 实现：`core/middleware/middleware.go` 单文件承载全部 8 个中间件；extends/rbac 为 `transport/middleware.go` 单文件
- 影响：M1「目录结构与设计一致」验收失败（同 C-2）

### MW-2 ✅ Gzip 空实现
- 设计：JSON 响应 >1KB 压缩，排除 /metrics、/health
- 实现：`core/router/engine.go` 的 `gzipMiddleware()` 只调 `c.Next()`，无压缩逻辑
- 影响：大响应传输带宽无优化（A-31 相关）

### MW-3 ✅ RateLimit 非滑动窗口
- 设计：ZSET 滑动窗口 + `X-RateLimit-Limit/Remaining/Retry-After` 头；全局限流 60 req/min 挂 engine
- 实现：`RateLimit` 用 INCR 固定窗口（仅 Retry-After 头）；全局限流未挂 engine，仅登录/注册路由手动挂载
- 影响：窗口边界可突刺；`rate_limit` 配置项未生效（A-20/A-23）

### MW-4 ✅ Trace 无 OTel span
- 设计：`otel.GetTextMapPropagator().Extract` + `tracer.Start` + `traceparent` 传播
- 实现：`Trace()` 仅生成/透传 `X-Trace-ID`（注释自认 placeholder for OTel）
- 影响：链路追踪验收失败（OBS-2）

### MW-5 ✅ Recovery 无请求体限制与堆栈
- 设计：`http.MaxBytesReader`（10MB → 413）+ panic 日志含 `debug.Stack()`
- 实现：Recovery 无 MaxBytesReader；panic 日志仅 error + request_id，无 stack
- 影响：超大请求体无上限（SEC-3）；panic 排障信息不足

### MW-6 ✅ SecurityHeaders 缺 CSP/HSTS 等
- 设计：CSP、Cache-Control、Pragma、HSTS、X-Frame-Options、X-Content-Type-Options、X-XSS-Protection、Referrer-Policy
- 实现：仅 X-Frame-Options(DENY)、X-Content-Type-Options(nosniff)、X-XSS-Protection、Referrer-Policy 四项
- 影响：CSP/XSS 纵深防御缺失（C-4 同源）

### MW-7 ✅ InitValidator 定义但从未调用
- 设计：Recovery 中间件中初始化 validator（trim + phone/idcard/nohtml/password/sort 五个校验器）
- 实现：`InitValidator()` 已定义（仅注册 `password` 一个），全仓库无任何调用；`RegisterReq` 等也未使用 `binding:"password"`
- 影响：自定义校验器形同虚设（VAL-1 同源，密码强度仍缺失）

## 一致项 ✅
- 中间件链顺序：RequestID → Recovery → Trace → Logger → Gzip → SecurityHeaders → CORS（engine 注册顺序与设计一致；RateLimit 按设计为路由级可选挂载）
- RequestID 透传/生成 X-Request-ID ✅
- Logger 记录 method/path/status/latency/ip/request_id/trace_id/body_size ✅
- CORS 白名单（allowedOrigins 配置化）✅
- Auth/RBAC/RequirePerm 三层职责分离（认证/授权/粒度）✅
