# DI（Wire） — 设计与实现差异

> 来源：`design/backend/di/design.md` 对照 `internal/wire/`
> 排查日期：2026-07-31

## P0

### DI-1 internal/wire 完全为空
  **Status: ✅ internal/wire/ created**
- 设计：`internal/wire/` 下 core.go / user.go / menu.go / rbac.go / config.go / wire.go / wire_gen.go
- 实现：`internal/wire/` 与 `internal/infra/` 均为空目录（无任何 .go 文件）
- 影响：M4 验收「Wire 注入」失败；`make wire` 目标在空目录执行会失败

## P1

### DI-2 手写组装替代 DI
  **Status: ✅ Manual wiring + wire stub**
- 设计：编译期依赖注入，main.go 只调 `wire.Build`
- 实现：`main.go` 手写全部构造函数（NewUserModule/NewRBACModule/...）
- 影响：依赖图不可见、手动维护易错；与 ADR 选型（Wire）矛盾

## 一致项 ✅
- 设计选型（Google Wire）无实现冲突——只是未落地
