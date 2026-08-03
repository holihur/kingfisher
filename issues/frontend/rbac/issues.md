# Frontend RBAC — 设计与实现差异

> 来源：`design/frontend/rbac/design.md` 对照 `web/src/pages/role/RoleList.tsx`
> 排查日期：2026-07-31

## P1

### FR-1 ✅ 权限 Tabs 组数与验收不符
- 设计：Tabs 按 Resource 分组 4 组（用户/菜单/角色/配置）
- 实现：按后端权限数据动态分组——种子含 audit 权限时出现 5 组（user/menu/role/config/audit）
- 影响：与验收预期 4 组不一致（设计侧种子含 audit:list 与验收 4 组矛盾，见 A-47 相关）

## P2

### FR-2 ✅ 无 403 页
- 设计：无权限访问显示 403 页面
- 实现：无 `/403` 路由；无权限按钮隐藏，但直接输入 URL 可访问页面（后端鉴权缺位时无兜底 UI）
- 影响：越权访问时空白/内容直出

## P2

### FR-3 ✅ 菜单分配树无选中状态回显优化
- 设计：菜单分配 Tree 勾选（父子联动）
- 实现：`menuApi.getTree` + Tree checkable 基本可用 ✅；半选/全选回显依赖 AntD 默认行为
- 影响：轻微

## 一致项 ✅
- 角色 CRUD + 权限分配弹窗（Checkbox 按 resource/action 列出）+ 菜单分配弹窗（Tree）✅
- 操作按权限渲染（role:update/role:delete/role:create）✅
- 权限数据来自 `GET /permissions`，已分配回显 `GET /roles/:id/permissions` ✅
