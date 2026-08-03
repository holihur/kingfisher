# Frontend State UI — 设计与实现差异

> 来源：`design/frontend/state-ui/design.md` 对照 `web/src/components/`、各页面
> 排查日期：2026-07-31

## P1

### FSU-1 ✅ 三态组件缺失
- 设计：`components/Skeleton.tsx`（table/card/detail/form 四模式）、`EmptyState`（插画+文案+操作）、`ErrorResult`（错误提示+重试）
- 实现：`components/` 仅 `PermissionBtn.tsx`；无任何三态组件
- 影响：各页面数据异常时无统一兜底（依赖 ProTable 默认，但自定义页面如 Dashboard 无 loading/empty/error 处理）

## P2

### FSU-2 ✅ 页面三态覆盖不足
- 设计：每个异步页面必须覆盖 Loading/Empty/Error
- 实现：Dashboard 无 loading/error；ConfigManage/MenuManage 等依赖 ProTable 默认 loading，错误仅 message；审计页无 Empty 定制
- 影响：弱网/报错场景体验不佳（A-29）

## 一致项 ✅
- ProTable 自带 loading 骨架（表格类页面有基础 Loading）✅
