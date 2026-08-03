# Frontend Layout — 设计与实现差异

> 来源：`design/frontend/layout/design.md` 对照 `web/src/layouts/AdminLayout.tsx`
> 排查日期：2026-07-31

## P1

### FL-1 ✅ 面包屑无层级
- 设计：Breadcrumb 按菜单层级生成（首页 / 系统管理 / 用户管理）
- 实现：`Breadcrumb items=[{首页}, {pathname 最后一段}]`——只有两级且为路径名
- 影响：深层页面（system/users）面包屑显示「users」而非「用户管理」

### FL-2 ✅ 无响应式抽屉
- 设计：移动端/窄屏 Sider 变 Drawer
- 实现：仅 `collapsible` 折叠（PC 语义），无 Drawer/断点
- 影响：窄屏布局不可用

### FL-3 ✅ 图标映射占位
- 设计：菜单 icon 名映射到 AntD 图标
- 实现：`icons` map 中 `MenuOutlined: <span />` 空占位；未命中 icon 回退 `QuestionOutlined`
- 影响：菜单图标显示异常（种子数据若含 MenuOutlined 则显示空白）

## P2

### FL-4 ✅ 无 flatMenus 工具
- 设计：菜单扁平化辅助（面包屑/权限用）
- 实现：AdminLayout 内联 buildItems 递归，无独立 flatMenus
- 影响：复用性差

## 一致项 ✅
- Sider 折叠 + Logo（折叠显示 K）+ 动态菜单（status==1 过滤、sort 排序）✅
- Header（折叠按钮 + 面包屑 + 用户头像下拉退出）✅
- 菜单 key 用 path、selectedKeys=[location.pathname] ✅
- 退出登录跳 /login ✅
- 登录后 fetchMenus/fetchUserInfo ✅
