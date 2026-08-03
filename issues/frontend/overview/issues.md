# Frontend Overview — 设计与实现差异

> 来源：`design/frontend/overview/design.md` 对照 `web/`
> 排查日期：2026-07-31

## P1

### FOV-1 目录名与设计不一致
- 设计：项目目录 `kingfisher-web/`
- 实现：目录为 `web/`
- 影响：M5 验证命令 `cd kingfisher-web` 直接失败；文档与实际路径不一致

### FOV-2 路由版本与设计不一致
- 设计：React Router v6
- 实现：`package.json` 为 `react-router-dom ^7.18.2`
- 影响：v6/v7 API 大体兼容（createBrowserRouter 均支持），但设计与实现版本声明不一致

### FOV-3 目录结构缺项
- 设计：`router/guard.tsx`、`router/dynamic.tsx`（动态路由）、`stores/app.ts`（全局 UI 状态）、`types/`（类型）、`utils/permission.ts`、`components/Skeleton.tsx` 等
- 实现：`router/` 仅 `index.tsx`；`stores/` 仅 `auth.ts`、`menu.ts`；`types/` 与 `utils/` 为空目录；`components/` 仅 `PermissionBtn.tsx`
- 影响：动态路由（按后端菜单生成）、全局 UI 状态、共享类型、权限工具均缺失

## P2

### FOV-4 遗留文件
- 实现：`web/tsconfig.app.json.new` 残留
- 影响：轻微，建议清理

## 一致项 ✅
- 技术栈（React + TS + Vite + AntD5 + Zustand + Axios + @ant-design/pro-table + @ant-design/icons）✅
- `main.tsx`/`App.tsx` 入口、`layouts/AdminLayout.tsx`、pages 目录（login/dashboard/user/menu/role/config/audit）✅
- `@` alias 已配置 ✅
