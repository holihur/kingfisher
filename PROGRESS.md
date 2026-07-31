# 实现进度

> 最后更新：2026-07-31
> 策略：每个里程碑完成后即可运行验证，不等到全部结束

---

## 🏁 里程碑总览

| 里程碑 | 内容 | 可验证 |
|--------|------|--------|
| **M1** | Core 框架启动 | `curl localhost:8080/health` 返回 200 |
| **M2** | 用户注册登录 | `curl POST /api/v1/auth/login` 拿到 token |
| **M3** | RBAC 鉴权 | 无权限用户访问管理接口返回 403 |
| **M4** | 后端 API 全量 | Swagger UI 可见全部接口，curl 全部通过 |
| **M5** | 前端登录闭环 | 浏览器输入账号密码 → 登录 → 看到后台布局 |
| **M6** | 全栈 CRUD 闭环 | 在浏览器上完成用户/菜单/角色/配置的增删改查 |
| **M7** | 生产就绪 | Docker 一键启动，CI 通过 |

---

## M1 — Core 框架启动 ✅ 0/19

> 验证：`curl localhost:8080/health` → `{"status":"ok"}`

- [ ] core/config/config.go
- [ ] core/logger/logger.go
- [ ] core/errcode/errcode.go
- [ ] core/response/response.go
- [ ] core/database/gorm.go
- [ ] core/cache/redis.go
- [ ] core/jwt/jwt.go
- [ ] core/middleware/request_id.go
- [ ] core/middleware/trace.go
- [ ] core/middleware/recovery.go
- [ ] core/middleware/logger.go
- [ ] core/middleware/gzip.go
- [ ] core/middleware/security_headers.go
- [ ] core/middleware/cors.go
- [ ] core/middleware/ratelimit.go
- [ ] core/router/engine.go
- [ ] core/telemetry/tracer.go
- [ ] core/telemetry/metrics.go
- [ ] config/config.yaml
- [ ] go.mod
- [ ] cmd/server/main.go

## M2 — 用户注册登录 ✅ 0/9

> 验证：注册 → 登录 → 拿到 access_token + refresh_token

- [ ] extends/user/domain/user.go
- [ ] extends/user/port/repository.go
- [ ] extends/user/port/service.go
- [ ] extends/user/adapter/mysql/model.go
- [ ] extends/user/adapter/mysql/repo.go
- [ ] extends/user/app/service.go
- [ ] extends/user/transport/handler.go
- [ ] extends/user/transport/register.go
- [ ] extends/user/wire.go

> 依赖 M1 完成

## M3 — RBAC 鉴权 ✅ 0/15

> 验证：不同角色登录后访问同一接口，一个 200 一个 403

- [ ] extends/rbac/domain/role.go
- [ ] extends/rbac/domain/permission.go
- [ ] extends/rbac/port/role_repo.go
- [ ] extends/rbac/port/perm_repo.go
- [ ] extends/rbac/port/service.go
- [ ] extends/rbac/adapter/mysql/model.go
- [ ] extends/rbac/adapter/mysql/role_repo.go
- [ ] extends/rbac/adapter/mysql/perm_repo.go
- [ ] extends/rbac/app/role_service.go
- [ ] extends/rbac/app/permission_service.go
- [ ] extends/rbac/transport/role_handler.go
- [ ] extends/rbac/transport/permission_handler.go
- [ ] extends/rbac/transport/middleware.go
- [ ] extends/rbac/transport/register.go
- [ ] extends/rbac/wire.go

> 依赖 M2 完成

## M4 — 后端 API 全量 ✅ 0/20

> 验证：`make swagger` → 浏览器 `/swagger/index.html` 可见全部接口

- [ ] extends/menu/domain/menu.go
- [ ] extends/menu/port/repository.go
- [ ] extends/menu/port/service.go
- [ ] extends/menu/adapter/mysql/model.go
- [ ] extends/menu/adapter/mysql/repo.go
- [ ] extends/menu/app/service.go
- [ ] extends/menu/transport/handler.go
- [ ] extends/menu/transport/register.go
- [ ] extends/menu/wire.go
- [ ] extends/config/domain/config.go
- [ ] extends/config/port/repository.go
- [ ] extends/config/port/service.go
- [ ] extends/config/adapter/mysql/model.go
- [ ] extends/config/adapter/mysql/repo.go
- [ ] extends/config/app/service.go
- [ ] extends/config/transport/handler.go
- [ ] extends/config/transport/register.go
- [ ] extends/config/wire.go
- [ ] internal/wire/wire.go + wire_gen.go

- [ ] extends/audit/domain/audit.go
- [ ] extends/audit/port/repository.go
- [ ] extends/audit/adapter/mysql/model.go
- [ ] extends/audit/adapter/mysql/repo.go
- [ ] extends/audit/app/service.go
- [ ] extends/audit/transport/handler.go
- [ ] extends/audit/transport/middleware.go
- [ ] extends/audit/transport/register.go
- [ ] extends/audit/wire.go
- [ ] migrations/ (10 个 SQL 文件：8 建表 + 1 种子 + 1 ALTER)

> 依赖 M3 完成

## M5 — 前端登录闭环 ✅ 0/14

> 验证：浏览器访问 `localhost:5173` → 跳转登录页 → 输入 admin/Abcd1234 → 进入后台 → 看到侧边栏菜单

- [ ] package.json + vite.config.ts + tsconfig.json
- [ ] src/main.tsx + src/App.tsx
- [ ] src/api/request.ts
- [ ] src/api/auth.ts
- [ ] src/types/api.ts
- [ ] src/utils/token.ts
- [ ] src/stores/auth.ts
- [ ] src/stores/menu.ts
- [ ] src/router/guard.tsx
- [ ] src/router/index.tsx
- [ ] src/layouts/AdminLayout.tsx
- [ ] src/pages/login/index.tsx
- [ ] src/components/PermissionBtn.tsx
- [ ] .env.development

> 依赖 M4 完成

## M6 — 全栈 CRUD 闭环 ✅ 0/17

> 验证：在浏览器上完整操作 用户/菜单/角色/配置 的增删改查

- [ ] src/api/user.ts
- [ ] src/api/menu.ts
- [ ] src/api/role.ts
- [ ] src/api/config.ts
- [ ] src/pages/user/UserList.tsx
- [ ] src/pages/user/UserForm.tsx
- [ ] src/pages/menu/MenuManage.tsx
- [ ] src/pages/menu/MenuForm.tsx
- [ ] src/pages/role/RoleList.tsx
- [ ] src/pages/role/RoleForm.tsx
- [ ] src/pages/role/PermissionModal.tsx
- [ ] src/pages/role/MenuAssignModal.tsx
- [ ] src/pages/config/ConfigManage.tsx
- [ ] src/pages/config/ConfigEditModal.tsx
- [ ] src/pages/dashboard/index.tsx
- [ ] src/pages/audit/AuditLogList.tsx

> 依赖 M5 完成

## M7 — 生产就绪 ✅ 0/8

> 验证：`docker-compose up -d` → 浏览器完整可用

- [ ] deploy/Dockerfile (backend)
- [ ] deploy/Dockerfile (frontend)
- [ ] deploy/nginx.conf
- [ ] deploy/docker-compose.yaml
- [ ] deploy/docker-compose.dev.yaml
- [ ] Makefile
- [ ] .golangci.yml
- [ ] .github/workflows/ci.yaml

> 依赖 M6 完成

---

**总进度：0 / 82** &nbsp;|&nbsp; **当前里程碑：M1**
