# Frontend Auth — 设计与实现差异

> 来源：`design/frontend/auth/design.md` 对照 `web/src/pages/login/`、`web/src/stores/auth.ts`、`web/src/router/index.tsx`
> 排查日期：2026-07-31

## P1

### FA-1 ✅ 无注册页/路由
- 设计：登录页含「还没有账号？去注册」，注册页 + 注册接口对接
- 实现：登录页无注册入口；`router/index.tsx` 无 `/register` 路由
- 影响：M6「注册」闭环缺失（后端 register 接口存在但前端无入口）

### FA-2 ✅ 登录后无 redirect 回跳
- 设计：Auth Guard 记录来源路径，登录成功后回跳
- 实现：`navigate('/dashboard')` 硬编码；guard.tsx 不存在
- 影响：用户从深层页面被踢回登录后丢失原目标

### FA-3 ✅ 无独立 token 工具
- 设计：token 存取集中管理（utils/token.ts 或 auth store 封装）
- 实现：`localStorage` key（kingfisher_token/kingfisher_refresh）散落在 router/AuthGuard、request.ts、stores/auth.ts
- 影响：key 变更需多处同步

## P2

### FA-4 ✅ 登录页品牌细节
- 设计：Logo + 名称展示（🦜 Kingfisher）
- 实现：Card 内纯文本标题，无 Logo 图
- 影响：视觉验收（轻微）

## 一致项 ✅
- 登录表单（用户名/密码 + 主按钮）+ 校验 ✅
- Zustand auth store（token/refreshToken/userInfo/permissions/isLoggedIn/login/logout/fetchUserInfo）✅
- AuthGuard 检查 token（路由级）✅
- fetchUserInfo 拉取 `/users/me` ✅
