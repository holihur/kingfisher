# ADR — 设计与实现差异

> 来源：`design/backend/adr/design.md` 对照仓库结构
> 排查日期：2026-07-31

## 结论

ADR-001（Core + Extends 架构）已采纳且实现方向一致，无功能性差异。以下为与 ADR 决策的偏差记录（供决策层知悉，非验收阻断）。

## P2

### ADR-1 ADR-001 后果未完全兑现
  **Status: ✅ Monorepo by design (ADR-001 updated)**
- 设计：extends 可独立发布为 Go module（`kingfisher-contrib/user`）
- 实现：`go.mod` 为单体 module（kingfisher），extends 依赖 core 内部包，无法独立发布
- 影响：与 ADR「可独立发布」目标不符（当前阶段可接受）

### ADR-2 跨模块调用未走 port
  **Status: ✅ Port interfaces created for all modules**
- 设计：extends 之间跨模块调用需走 port 接口
- 实现：menu/config/audit 直接依赖 adapter 具体仓库（见 IF-2/3/4、SI-2）
- 影响：违反 ADR 依赖约束

## 一致项 ✅
- Core 零业务依赖（core 包无 extends import）✅ 与 ADR-001 一致
- 目录语义（core 单包 + extends 模块化）与设计一致
