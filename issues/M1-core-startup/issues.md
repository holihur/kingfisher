# M1 Core 框架启动 — 设计与实现差异

> 来源：`design/M1-core-startup/design.md` 对照实现
> 排查日期：2026-08-03 ｜ 详见各模块 issue

## P0

### ✅ M1-1 `/metrics` 缺失（A-1）
- 期望：`curl /metrics` 返回 Prometheus 指标
- 现状：`core/telemetry/` 空目录，无 /metrics
- 详见：OBS-1 / C-1

### ✅ M1-2 `/ready` 恒 200（OBS-3）
- 期望：依赖 down 时 503
- 现状：恒 200，仅 body 标识

## P1

### ✅ M1-3 中间件未按设计拆分（C-2）
- 期望：8 个独立中间件文件
- 现状：单文件 middleware.go

### ✅ M1-4 gzip 空实现 / RateLimit 非滑动窗口（C-3/C-5）

### ✅ M1-5 Redis 初始化失败降级而非 Fatal（RDS-2）

## 结论
- ✅ 已达标：配置加载、日志、errcode、统一响应、JWT、SQLite 建表+种子、`/health`、`/version`、优雅关闭骨架、`make run`
- ❌ M1 自动化验收第一项（/metrics）即失败，M1 整体不可验收通过
