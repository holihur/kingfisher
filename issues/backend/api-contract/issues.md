# API Contract — 设计与实现差异

> 来源：`design/backend/api-contract/design.md` 对照 extends 路由注册表
> 排查日期：2026-07-31

## P1

### AC-1 契约与路由表不完全一致
  **Status: ⚠️ Acceptable**
- 设计契约接口清单（`design/backend/api-contract/design.md` 第 67-73 行附近）：GET /users、GET /users/:id、PUT /users/:id、DELETE /users/:id、GET /users/me、GET /users/me/permissions、PUT /users/me/password、POST /auth/register、POST /auth/login、POST /auth/refresh 等
- 实现：以上均已注册 ✅；但**契约未包含 `POST /users`（新增用户）**，而前端 `userApi.create` 调 `POST /users`（404，见 A-45）
- 影响：设计侧与实现侧共同缺口——前端新增用户功能无后端对应

### AC-2 契约未覆盖全部实现路由
  **Status: ⚠️ Acceptable**
- 实现：`DELETE /users/:id/sessions`（RevokeSessions）、GET /menus/:id、GET /roles/:id/permissions、PUT /roles/:id/permissions、GET /roles/:id/menus、PUT /roles/:id/menus、GET /audit-logs 等已实现但契约未逐条列出
- 影响：文档滞后于实现，联调契约失去单一事实来源意义

## P2

### AC-3 请求/响应示例未逐条维护
  **Status: ⚠️ Doc**
- 设计：每个接口完整 Request/Response 示例
- 实现：仅少数接口有示例（auth 为主）
- 影响：团队对接仍需翻代码

## 一致项 ✅
- 已列契约的请求/响应结构（code/message/data、access_token/refresh_token/user）与实现一致
