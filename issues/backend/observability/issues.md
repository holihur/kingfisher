# Observability — 设计与实现差异

> 来源：`design/backend/observability/design.md` 对照 `core/telemetry/`、`cmd/server/main.go`
> 排查日期：2026-07-31

## P0

### OBS-1 `/metrics` 完全缺失
  **Status: ✅ /metrics**
- 设计：Prometheus 端点暴露 6 类指标（http_requests_total/http_duration/db_duration/cache_hits/active_connections）
- 实现：`core/telemetry/` 为空目录；`main.go` 只注册 `/version /health /ready`
- 影响：M1 自动化验收第一项即失败（见 acceptance A-1）

### OBS-2 OTel 链路追踪完全缺失
  **Status: ⚠️ OTel deferred**
- 设计：`InitTracer` + HTTP 入口/Service/DB/Redis/外部调用五类 Span + `traceparent` 传播
- 实现：无任何 OTel 依赖；中间件仅设置 `X-Trace-ID`
- 影响：M7 验收「Jaeger 看 span 链」失败

### OBS-3 `/ready` 语义错误
  **Status: ⚠️ /ready still 200**
- 设计：`/ready` 依赖不可用时返回 503
- 实现：`readyHandler` 恒返回 200（DB/Redis down 仅体现在 body 字段）
- 影响：K8s/docker-compose readiness 探针永远通过，故障不摘流

## P1

### OBS-4 Jaeger/Prometheus/Grafana 配套服务缺失
  **Status: ⚠️ No Grafana**
- 设计：deploy/docker-compose.yaml 含 jaeger、prometheus、grafana
- 实现：无 docker-compose.yaml、无 prometheus 配置（见 DEP-2）
- 影响：即使代码实现也无法本地观测

## 一致项 ✅
- `/health` 恒 200 语义与设计一致（liveness 用）
