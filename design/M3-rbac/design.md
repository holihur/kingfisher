# M3 — RBAC 鉴权

## 目标

不同角色登录后访问同一接口，一个 200 一个 403。

## 覆盖模块

| 模块 | 设计文档 | 说明 |
|------|----------|------|
| RBAC 总览 | [extends/rbac](../backend/extends/rbac/design.md) | Role + Permission 实体 |
| 接口定义 | [interface](../backend/interface/design.md) | RoleRepository + PermissionRepository |
| Auth 中间件 | [extends/rbac](../backend/extends/rbac/design.md) | JWT 解析 → user_id/role 注入 ctx |
| RBAC 中间件 | [extends/rbac](../backend/extends/rbac/design.md) | 权限校验 + RequirePerm |
| 角色管理 | [extends/rbac](../backend/extends/rbac/design.md) | CRUD + 分配权限 + 分配菜单 |
| Service 接口 | [service-interface](../backend/service-interface/design.md) | RoleService 接口化 |
| 安全 | [security](../backend/security/design.md) | 越权防护 |

## 验证

```bash
# 用 admin 登录
ADMIN_TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login \
  -d '{"username":"admin","password":"Abcd1234"}' | jq -r '.data.access_token')

# 用 viewer 登录
VIEWER_TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login \
  -d '{"username":"viewer","password":"Abcd1234"}' | jq -r '.data.access_token')

# admin 可以创建用户
curl -X POST localhost:8080/api/v1/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"username":"newuser","password":"Abcd1234"}'
# → 200

# viewer 不可以创建用户
curl -X POST localhost:8080/api/v1/users \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -d '{"username":"newuser","password":"Abcd1234"}'
# → 403 {"code":10004,"message":"forbidden"}
```

## 产出文件

| 文件 | 说明 |
|------|------|
| `extends/rbac/domain/role.go` | Role 实体 |
| `extends/rbac/domain/permission.go` | Permission 实体 |
| `extends/rbac/port/role_repo.go` | RoleRepository 接口 |
| `extends/rbac/port/perm_repo.go` | PermissionRepository 接口 |
| `extends/rbac/port/service.go` | RoleService + PermissionService 接口 |
| `extends/rbac/adapter/mysql/model.go` | GORM PO |
| `extends/rbac/adapter/mysql/role_repo.go` | 角色数据访问 |
| `extends/rbac/adapter/mysql/perm_repo.go` | 权限数据访问 |
| `extends/rbac/app/role_service.go` | 角色业务逻辑 |
| `extends/rbac/app/permission_service.go` | 权限业务逻辑 |
| `extends/rbac/transport/role_handler.go` | 角色 Handler |
| `extends/rbac/transport/permission_handler.go` | 权限 Handler |
| `extends/rbac/transport/auth_middleware.go` | Auth（认证） |
| `extends/rbac/transport/rbac_middleware.go` | RBAC（授权） |
| `extends/rbac/transport/require_perm.go` | RequirePerm（粒度） |
| `extends/rbac/transport/register.go` | 路由注册 |
| `extends/rbac/wire.go` | Provider Set |
## 验收

→ [acceptance 验收清单](../acceptance/design.md)
