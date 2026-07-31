# Cache — 设计与实现差异

> 来源：`design/backend/cache/design.md` 对照 `core/cache/`、`extends/*`
> 排查日期：2026-07-31

## P0

### CA-1 缓存使用模式大面积缺失
- 设计：Cache-Aside（菜单树/权限列表）、Write-Through（系统配置）、空值缓存防穿透、布隆过滤器
- 实现：仅 RBAC 的 `GetUserPermissions` 用了 Cache-Aside（且因 `strSlice` 占位返回 nil 而失效，见 ER-1）；菜单树 `GetTree`、配置 `GetAll` 均无缓存
- 影响：M4 验收「配置修改立即生效」「菜单树缓存」缺失；高并发下全部打 DB

### CA-2 配置写后缓存失效未实现
- 设计：`PUT /configs/:key` 后应删除/更新缓存
- 实现：`extends/config` 完全无缓存依赖（Service 无 cache 字段）
- 影响：即使未来加读缓存，写路径也未联动失效

## P1

### CA-3 RBAC 缓存失效用通配符 Delete
- 设计：`Delete(ctx, keys ...string)` 精确 key
- 实现：`AssignPermissions` 调用 `cache.Delete(ctx, "user:perms:*")`——Redis 的 DEL 不支持通配符，只会删字面 key `user:perms:*`
- 影响：权限变更后用户缓存不失效，权限更新不生效（配合 ER-1 实际未触发）

### CA-4 空值缓存/布隆过滤器未实现
- 设计：防缓存穿透（空值缓存、布隆过滤器）
- 实现：无任何实现
- 影响：热点空 key 请求穿透到 DB

## 一致项 ✅
- `core/cache.Cache` 接口（Get/Set/Delete/Exists/Incr/Expire）存在，JWT 黑名单模式（`blacklist:token:{jti}`）与设计一致
