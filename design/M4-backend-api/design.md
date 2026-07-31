# M4 — 后端 API 全量就绪

## 目标

Swagger UI 可见全部接口，菜单树 + 系统配置可用，curl 全部通过。

## 覆盖模块

| 模块 | 设计文档 | 说明 |
|------|----------|------|
| 菜单 | [extends/menu](../backend/extends/menu/design.md) | 树形 CRUD，递归构建 |
| 配置 | [extends/config](../backend/extends/config/design.md) | 键值对 CRUD + 缓存 |
| 审计 | [extends/audit](../backend/extends/audit/design.md) | 操作日志异步写入 |
| 迁移 | [migration](../backend/migration/design.md) | 8 个 SQL + 种子数据 |
| DI | [di](../backend/di/design.md) | Wire 编译期注入 |
| API 契约 | [api-contract](../backend/api-contract/design.md) | 每个接口的完整请求/响应 |
| Swagger | [swagger-checklist](../backend/swagger-checklist/design.md) | 注解规范 |
| 事务 | [transaction](../backend/transaction/design.md) | UnitOfWork + ctx 传递 |
| 缓存 | [cache](../backend/cache/design.md) | Cache-Aside + 穿透防护 |
| 服务接口 | [service-interface](../backend/service-interface/design.md) | Service 接口化 |

## 验证

```bash
# 生成文档
make swagger

# 浏览器验证
open http://localhost:8080/swagger/index.html
# 可见 4 组接口：Auth / User / Menu / Role Permissions / Config

# 菜单树
curl localhost:8080/api/v1/menus/tree \
  -H "Authorization: Bearer $ADMIN_TOKEN"
# → 树形菜单 JSON

# 角色管理
curl localhost:8080/api/v1/roles \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# 系统配置
curl localhost:8080/api/v1/configs \
  -H "Authorization: Bearer $ADMIN_TOKEN"
# → {"site_name":"Kingfisher Admin",...}

# 数据库迁移
make migrate-up
mysql -e "SELECT id,username,role_id FROM users;"   # 1 | admin | 1
mysql -e "SELECT COUNT(*) FROM permissions;"          # 14
```

## 产出文件

| 文件 | 说明 |
|------|------|
| `extends/menu/*` | 菜单模块全部文件 |
| `extends/config/*` | 配置模块全部文件 |
| `internal/wire/wire.go` | Wire 主入口 |
| `internal/wire/wire_gen.go` | Wire 生成 |
| `migrations/000001~000008` | SQL 迁移 + 种子 |
## 验收

→ [acceptance 验收清单](../acceptance/design.md)
