# M5 — 前端登录闭环

## 目标

浏览器 `localhost:5173` → 登录页 → 输入账号密码 → 进入后台 → 看到侧边栏菜单。

## 覆盖模块

| 模块 | 设计文档 | 说明 |
|------|----------|------|
| 前端总览 | [frontend overview](../frontend/overview/design.md) | 技术栈 + 项目结构 |
| 本地联调 | [local-dev](../frontend/local-dev/design.md) | Vite proxy + 环境变量 |
| 登录 | [auth](../frontend/auth/design.md) | 登录页 + Zustand Store + Auth Guard |
| Axios | [request](../frontend/request/design.md) | 双拦截器 + token 刷新 |
| 布局 | [layout](../frontend/layout/design.md) | AdminLayout + 动态侧边栏 |

## 验证

```bash
# 启动后端
cd kingfisher && make run

# 启动前端
cd kingfisher-web && npm run dev

# 浏览器
open http://localhost:5173
# → 自动跳转 /login
# → 输入 admin / Abcd1234
# → 进入后台，左侧出现菜单树（Dashboard + 系统管理 > 用户管理/菜单管理/角色管理/系统配置）
# → 顶栏显示用户名 "admin"，右侧有退出按钮
# → 点击退出，返回登录页
```

## 产出文件

| 文件 | 说明 |
|------|------|
| `kingfisher-web/package.json` | 依赖 |
| `kingfisher-web/vite.config.ts` | Vite 配置 + proxy |
| `kingfisher-web/tsconfig.json` | TypeScript |
| `kingfisher-web/.env.development` | 环境变量 |
| `kingfisher-web/src/main.tsx` | 入口 |
| `kingfisher-web/src/App.tsx` | 根组件 |
| `kingfisher-web/src/api/request.ts` | Axios + 拦截器 |
| `kingfisher-web/src/api/auth.ts` | 登录 API |
| `kingfisher-web/src/api/menu.ts` | 菜单 API |
| `kingfisher-web/src/types/api.ts` | 类型定义 |
| `kingfisher-web/src/utils/token.ts` | Token 存取 |
| `kingfisher-web/src/stores/auth.ts` | Auth Store |
| `kingfisher-web/src/stores/menu.ts` | Menu Store |
| `kingfisher-web/src/router/guard.tsx` | Auth Guard |
| `kingfisher-web/src/router/index.tsx` | 路由配置 |
| `kingfisher-web/src/layouts/AdminLayout.tsx` | 后台布局 |
| `kingfisher-web/src/pages/login/index.tsx` | 登录页 |
| `kingfisher-web/src/components/PermissionBtn.tsx` | 权限按钮 |
## 验收

→ [acceptance 验收清单](../acceptance/design.md)
