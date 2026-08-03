# Extends/RBAC — 设计与实现差异

> 来源：`design/backend/extends/rbac/design.md` 对照 `extends/rbac/`
> 排查日期：2026-07-31

## P0

### ER-1 `strSlice()` 占位返回 nil → 权限缓存命中即空
- 设计：`GetUserPermissions` 缓存命中后反序列化权限列表
- 实现：`strSlice(s string) []string { return nil } // placeholder`——缓存命中（Redis 可用）时返回空列表
- 影响：Redis 可用时**所有用户权限为空** → `RequirePerm` 全部 403（A-5）；Redis 不可用反而正常（直查 DB）
- 影响面：roles 写操作（role:create/update/delete）全部 403

### ER-2 RBAC 中间件忽略错误
- 设计：权限获取失败应 500 或至少记录
- 实现：`RBACMiddleware` 中 `perms, _ := roleSvc.GetUserPermissions(...)` 忽略 error；`RequirePerm` 中 `ps, _ := c.Get("permissions")` 也忽略
- 影响：DB 故障时静默放行/误判，无日志

### ER-3 中间件未挂到业务路由
- 设计：所有受保护路由都应过 RBAC 权限校验
- 实现：`RequirePerm` 仅用于 roles 路由；users/menus/configs/audit 路由均未挂（A-2/A-3/A-38）
- 影响：viewer 可读写用户/菜单/配置

## P1

### ER-4 角色层级保护缺失
- 设计：禁止低角色修改高角色、禁止修改自己角色
- 实现：`RoleService.Update/Delete` 无调用者 userID 参数、无层级校验（见 SEC-2）
- 影响：越权修改角色

### ER-5 AssignPermissions 无校验
- 设计：分配前校验 role/perm 存在性
- 实现：直接 `repo.AssignPermissions`（先删后插），无存在性校验
- 影响：无效 role_id/perm_id 静默写入或报错

### ER-6 RBAC 缓存失效用通配符 DELETE
- 设计：`Delete(ctx, keys ...string)` 精确 key
- 实现：`AssignPermissions` 后 `cache.Delete(ctx, "user:perms:*")`——Redis DEL 不支持通配
- 影响：权限更新后缓存不失效（见 CA-3）

### ER-7 菜单分配未校验菜单存在
- 设计：AssignMenus 应校验 menu_ids 存在
- 实现：直接写入 role_menus 关联
- 影响：脏关联数据

## P2

### ER-8 文件未拆分
- 设计：role_service.go / permission_service.go / role_handler.go / permission_handler.go / auth_middleware.go / rbac_middleware.go / require_perm.go 分文件
- 实现：service 合并单文件 `app/service.go`；handler 合并 `transport/handler.go`；middleware 合并 `transport/middleware.go`
- 影响：M3「目录结构与设计一致」验收受影响（轻微）

### ER-9 `permission_service.go` 与 `perm_repo.go` 命名
- 设计：`port/role_repo.go`、`port/perm_repo.go`
- 实现：`port/repository.go` 合并定义
- 影响：轻微

## 一致项 ✅
- Role/Permission domain 字段与设计一致
- `GetUserPermissions` SQL 按 user.role_id join role_permissions ✅（缓存问题除外）
- RequirePerm 逻辑（map 包含即放行）与设计一致
