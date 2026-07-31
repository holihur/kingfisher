# M2 — 用户注册登录

## 目标

`curl POST /api/v1/auth/login` → 拿到 access_token + refresh_token。

## 覆盖模块

| 模块 | 设计文档 | 说明 |
|------|----------|------|
| 用户领域 | [extends/user](../backend/extends/user/design.md) | User 实体 |
| 接口定义 | [interface](../backend/interface/design.md) | UserRepository 接口 |
| 数据层 | [mysql](../backend/mysql/design.md) | adapter/mysql PO + repo |
| 业务层 | [extends/user](../backend/extends/user/design.md) | Register + Login + RefreshToken |
| 接口层 | [extends/user](../backend/extends/user/design.md) | Handler + swagger 注解 |
| 路由注册 | [extends/user](../backend/extends/user/design.md) | 实现 core.Module |
| 安全 | [security](../backend/security/design.md) | 登录限流 + 密码策略 |

## 验证

```bash
# 注册
curl -X POST localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"test","password":"Abcd1234","email":"test@example.com"}'

# 登录
curl -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Abcd1234"}'
# → {"code":0,"data":{"access_token":"eyJ...","refresh_token":"eyJ...","user":{...}}}

# 用 token 访问受保护接口
curl localhost:8080/api/v1/users/1 \
  -H 'Authorization: Bearer <access_token>'
```

## 产出文件

| 文件 | 说明 |
|------|------|
| `extends/user/domain/user.go` | User 实体 |
| `extends/user/port/repository.go` | UserRepository 接口 |
| `extends/user/port/service.go` | AuthService + UserService 接口 |
| `extends/user/adapter/mysql/model.go` | GORM PO |
| `extends/user/adapter/mysql/repo.go` | 数据访问实现 |
| `extends/user/app/service.go` | 注册/登录/刷新逻辑 |
| `extends/user/transport/handler.go` | Handler |
| `extends/user/transport/register.go` | 路由注册 |
| `extends/user/wire.go` | Provider Set |
## 验收

→ [acceptance 验收清单](../acceptance/design.md)
