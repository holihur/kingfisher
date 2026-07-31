# Frontend Overview — 前端总览

## 1. 技术栈

| 层面 | 选型 | 版本 | 理由 |
|------|------|------|------|
| 框架 | React | 18+ | 生态最大、TypeScript 支持最好 |
| 语言 | TypeScript | 5.5+ | 类型安全 |
| 构建 | Vite | 5+ | 秒级 HMR，替代 CRA |
| UI 库 | Ant Design | 5+ | 企业级组件，表格/表单/树开箱即用 |
| 路由 | React Router | v6 | 嵌套路由、loader/action |
| 状态管理 | Zustand | v4 | 轻量，无 boilerplate |
| HTTP | Axios | v1.7+ | 拦截器、请求取消 |
| 图标 | @ant-design/icons | — | 与 AntD 配套 |
| 表格 | @ant-design/pro-table | — | 搜索+分页+工具栏一体 |

## 2. 项目结构

```
kingfisher-web/
├── public/
│   └── favicon.ico
├── src/
│   ├── main.tsx                   # 入口
│   ├── App.tsx                    # 根组件（路由 + 权限守卫）
│   ├── router/
│   │   ├── index.tsx              # 路由配置
│   │   ├── guard.tsx              # Auth Guard（登录检查）
│   │   └── dynamic.tsx            # 动态路由（从后端菜单生成）
│   ├── stores/                    # Zustand
│   │   ├── auth.ts                # 用户状态、token
│   │   ├── menu.ts                # 菜单树
│   │   └── app.ts                 # 全局 UI 状态
│   ├── layouts/
│   │   └── AdminLayout.tsx        # 侧边栏 + 顶栏 + 内容
│   ├── pages/
│   │   ├── login/                 # 登录页
│   │   ├── dashboard/             # 首页/Dashboard
│   │   ├── user/                  # 用户管理
│   │   │   ├── UserList.tsx
│   │   │   └── UserForm.tsx
│   │   ├── menu/                  # 菜单管理
│   │   │   └── MenuManage.tsx
│   │   ├── role/                  # 角色管理
│   │   │   ├── RoleList.tsx
│   │   │   └── RoleForm.tsx
│   │   └── config/                # 系统配置
│   │       └── ConfigManage.tsx
│   ├── components/                # 通用组件
│   │   ├── PermissionBtn.tsx      # 权限按钮（无权限不渲染）
│   │   └── IconSelect.tsx         # 图标选择器
│   ├── api/                       # API 层
│   │   ├── request.ts             # Axios 实例 + 拦截器
│   │   ├── auth.ts                # 登录/注册/刷新 API
│   │   ├── user.ts                # 用户 CRUD API
│   │   ├── menu.ts                # 菜单 API
│   │   ├── role.ts                # 角色 API
│   │   └── config.ts              # 配置 API
│   ├── types/                     # TS 类型定义
│   │   ├── api.ts                 # 通用响应类型
│   │   ├── user.ts
│   │   ├── menu.ts
│   │   ├── role.ts
│   │   └── config.ts
│   └── utils/
│       ├── token.ts               # Token 存取（localStorage）
│       └── permission.ts          # 前端权限判断
├── .env                           # 环境变量
# VITE_API_TARGET — Vite proxy target（开发）
# VITE_API_BASE_URL — axios baseURL（生产直连）
├── .env.development
├── .env.production
├── vite.config.ts
├── tsconfig.json
├── package.json
└── Dockerfile
```

## 3. 数据流

```
Page → API (axios) → Backend
         ↕
    request.ts (拦截器: 注入 token、刷新 token、统一错误提示)
         ↕
    Zustand Store (全局状态: userInfo, menuTree, permissions)
         ↕
    Components (消费状态 + 渲染)
```

## 4. 路由设计

```
/login                  → LoginPage         （公开）
/                       → AdminLayout        （需登录）
  /dashboard            → Dashboard          （首页）
  /system               → （目录，不渲染页面）
    /system/users       → UserList           （用户管理）
    /system/menus       → MenuManage         （菜单管理）
    /system/roles       → RoleList           （角色管理）
    /system/configs     → ConfigManage       （系统配置）
```

## 5. 权限控制

| 层级 | 机制 |
|------|------|
| 路由级 | AuthGuard 检查 token 是否存在 |
| 菜单级 | 侧边栏从后端菜单树动态渲染 |
| 按钮级 | `<PermissionBtn code="user:delete">` 控制按钮显隐 |

## 6. 模块索引

| 模块 | 文档 | 说明 |
|------|------|------|
| 布局 | [layout](../layout/design.md) | AdminLayout：侧边栏+顶栏+内容区 |
| 状态 UI | [state-ui](../state-ui/design.md) | Skeleton/Empty/Error 三种异步状态规范 |
| 登录 | [auth](../auth/design.md) | 登录页、注册页 |
| 请求 | [request](../request/design.md) | Axios 拦截器、token 刷新、错误处理 |
| 用户 | [user](../user/design.md) | 用户管理 CRUD 页面 |
| 菜单 | [menu](../menu/design.md) | 菜单树管理 |
| 角色 | [rbac](../rbac/design.md) | 角色+权限管理 |
| 配置 | [config](../config/design.md) | 系统配置管理 |
