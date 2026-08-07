# M3 RBAC 鉴权 — 设计与实现差异

> 来源：`design/M3-rbac/design.md` 对照实现
> 排查日期：2026-08-03 ｜ 详见各模块 issue

## P0（验收核心场景全部失败）

### ✅ M3-1 viewer 访问 /users 不 403（A-2）
- 期望：viewer 登录 → GET /api/v1/users → 403
- 现状：users 路由未挂 RequirePerm，200 返回

### ✅ M3-2 PUT /configs/:key 无权限校验（A-3）
### ✅ M3-3 菜单树不按角色过滤（A-4）
### ✅ M3-4 权限缓存命中返回空 → 全后台 403（ER-1）
- 期望：Redis 可用时权限正常
- 现状：`strSlice` 占位返回 nil，缓存命中即空 → RequirePerm 全 403

### ✅ M3-5 角色层级保护缺失（SEC-2/ER-4）

## P1

### ✅ M3-6 RBAC 中间件忽略错误（ER-2）
### ✅ M3-7 缓存失效通配符 Delete 无效（CA-3/ER-6）
### ✅ M3-8 RequirePerm 仅 roles 路由使用（ER-3）

## 结论
- ✅ 已达标：Auth 中间件（Bearer 解析）、RequirePerm 机制本身、角色 CRUD + 权限/菜单分配接口
- ❌ M3 验收「不同角色 200 vs 403」双向场景均失败——admin 也可能被 403（M3-4），viewer 该 403 却 200（M3-1）
