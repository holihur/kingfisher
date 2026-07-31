# Backend Overview — 设计与实现差异

> 来源：`design/backend/overview/design.md` 对照仓库结构
> 排查日期：2026-07-31

## P1

### BOV-1 目录结构与设计部分不符
- 设计：`core/` 含 config/logger/errcode/response/jwt/database/cache/router/middleware/telemetry
- 实现：core 各子包存在 ✅，但 `core/middleware` 为单文件（见 C-2）、`core/telemetry` 空（见 C-1）
- 影响：M1「目录结构与设计一致」验收失败

### BOV-2 extends 模块无 wire.go
- 设计：每个 extends 模块含 `wire.go`（Wire Provider）
- 实现：所有 extends 模块无 wire.go（DI 未落地，见 DI-1）
- 影响：与 overview 架构图不一致

## 一致项 ✅
- 依赖方向 `main → extends → core` 成立（core 无 extends import）✅
- 模块目录骨架（domain/port/app/adapter/transport）在 user/rbac 中成立（menu/config/audit 缺 port）
- cmd/server/main.go 作为唯一入口 ✅
