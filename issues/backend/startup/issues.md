# Startup — 设计与实现差异

> 来源：`design/backend/startup/design.md` 对照 `cmd/server/main.go`
> 排查日期：2026-07-31

## P1

### ST-1 Wire 组装未接入
- 设计：main.go 通过 `internal/wire` 依赖注入组装
- 实现：`main.go` 手写全部构造函数；`internal/wire/` 为空目录
- 影响：DI 设计未落地（见 DI-1）；Makefile `wire` 目标会失败

### ST-2 Telemetry 初始化缺失
- 设计：`cfg.Telemetry.Enabled` 时 `InitTracer` + defer Shutdown
- 实现：无 telemetry 引用（配置也无 Telemetry 段？——`core/config` 未见）
- 影响：M1/M7 可观测性验收失败

### ST-3 迁移执行缺失
- 设计：非 release 模式执行 `RunMigrations(db, "migrations")`
- 实现：仅 SQLite `AutoMigrate` + `SeedData`；无 RunMigrations
- 影响：MySQL/PG 无表结构（A-24）

### ST-4 Redis 失败处理与设计冲突
- 设计：`rdb` 初始化失败 → `logger.Fatal`
- 实现：降级 Warn + nil cache
- 影响：与设计强依赖策略冲突（见 RDS-2）

### ST-5 优雅关闭顺序缺 Telemetry
- 设计：关闭顺序 server → modules → tracer.Shutdown
- 实现：server → modules 逆序，无 tracer（未初始化）
- 影响：无 span 可冲刷，风险面小但设计项缺失

## 一致项 ✅
- 配置加载 → 日志 → DB → Redis → JWT → Engine → 注册模块 → HTTP Server → 优雅关闭的骨架顺序与设计一致
- `ReadTimeout/WriteTimeout/IdleTimeout` 已配置（对应超时攻击防护）✅
- `/version` 通过 `-ldflags` 注入 version/commit/build_time ✅
