# Read/Write Split — 设计与实现差异

> 来源：`design/backend/readwrite-split/design.md` 对照 `core/database/`
> 排查日期：2026-07-31

## 结论

该模块设计为「可选/为未来拆分做准备」。当前实现为单库（开发模式），符合设计「初期可读写同库」的前提。

## P2

### RW-1 ✅ DBResolver 未引入
  **Status: ✅ Implemented in v1**
- 设计：GORM `dbresolver` 插件（Master/Replica 配置 + RandomPolicy）
- 实现：`core/database/gorm.go` 无 dbresolver 引用，配置结构无 Replica 字段
- 影响：设计明确「无从库——读写同库（不影响现有行为）」，属未启用而非缺陷；如需启用需补配置结构 + NewDatabaseWithRW

## 一致项 ✅
- 单库连接（读写同库）与设计初期目标一致
