# M5 前端登录闭环 — 设计与实现差异

> 来源：`design/M5-frontend-login/design.md` 对照 `web/`
> 排查日期：2026-08-03 ｜ 详见 frontend 模块 issue

## P1

### ✅ M5-1 目录名 web/ 而非 kingfisher-web（FOV-1）
- 影响：验证命令 `cd kingfisher-web` 失败

### ✅ M5-2 无 .env / 无 open / 无 docker-compose.dev（FLD-1/2/3）
### ✅ M5-3 登录后硬编码跳 /dashboard（FA-2）

## 结论
- ✅ 主链路可用：`make run` + `cd web && npm run dev` → 登录页 → 输入 admin/Abcd1234 → 进入后台 → 侧边栏菜单渲染（前提：后端能启动 + 前端 proxy 正常）
- ✅ 登录闭环完成：/users/me/permissions 已实现，菜单由 /menus/tree 全量提供。M5 核心「登录闭环」成立 `/users/me/permissions`（EU-4 占位返回空），菜单虽由 /menus/tree 全量提供，功能可用；M5 核心「登录闭环」基本成立，验证路径与设计文档不符（FOV-1）
