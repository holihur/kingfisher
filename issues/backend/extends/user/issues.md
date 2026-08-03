# Extends/User — 设计与实现差异

> 来源：`design/backend/extends/user/design.md` 对照 `extends/user/`
> 排查日期：2026-07-31

## P0

### EU-1 登录硬编码角色 viewer
- 设计：`Login` 从用户记录取角色（`user.Role`），签发 token 角色应与用户实际角色一致
- 实现：`Login` 调用 `jwtMgr.GenerateToken(ctx, user.ID, "viewer", ...)` 硬编码 `"viewer"`
- 影响：admin 登录拿到的也是 viewer 角色声明 → 前端权限/路由按角色判断会错乱；M3 角色差异化验收失败

### EU-2 `POST /users` 未注册
- 设计：用户管理应支持创建用户（前端 UserForm/新增用户）
- 实现：`register.go` 无 `users.POST("")`；前端 `userApi.create` → 404（见 AC-1、A-45）
- 影响：M6 验收「新增用户」失败

### EU-3 软删除未实现但 SQL 写 deleted_at
- 设计：用户删除（软删除 `gorm.DeletedAt` 或物理删除需明确）
- 实现：`UserRepo.Delete` 执行 `UPDATE users SET deleted_at=?`，但 `userPO` 无 `DeletedAt` 字段 → SQLite 报「no such column: deleted_at」→ 500
- 影响：删除用户接口必坏（A-46）

### EU-4 `GetUserPermissions` 占位返回 nil
- 设计：`GET /users/me/permissions` 返回当前用户权限
- 实现：`UserService.GetUserPermissions` 注释「Placeholder: returns empty list」直接 `return nil, nil`
- 影响：前端权限渲染（PermissionBtn/菜单过滤）拿不到权限；RBAC 中间件走 RBAC 模块路径不受影响，但 `/users/me/permissions` 接口验收失败

## P1

### EU-5 session_version 校验未接入
- 设计：`ChangePassword`/`RevokeSessions` 递增 session_version 后旧 token 立即失效
- 实现：递增已实现，但 AuthMiddleware 不比对 `sv`（见 JWT-2）
- 影响：改密/踢出后旧 token 仍可用（A-19）

### EU-6 无 logout 接口
- 设计：注销（RevokeToken 黑名单）
- 实现：无 `POST /auth/logout`；仅 `DELETE /users/:id/sessions`（管理员踢出）
- 影响：前端登出仅清本地 token，服务端黑名单机制无入口

### EU-7 注册密码强度缺失
- 设计：8-64 位 + 大写/小写/数字
- 实现：`binding:"required,min=8,max=64"` 仅长度（见 SEC-4、VAL-1）
- 影响：弱密码可注册

### EU-8 注册邮箱校验缺失
- 设计：邮箱格式校验 + 唯一性
- 实现：`Email` 无 binding（可选）、注册时未查重
- 影响：重复邮箱可注册

## P2

### EU-9 无 `port/service.go`
- 设计：AuthService/UserService 接口化
- 实现：仅 `port/repository.go`（见 SI-1）

### EU-10 错误映射粗糙
- 设计：区分 ErrUserExists/ErrPasswordWrong/ErrTokenExpired 等
- 实现：Service 返回 `fmt.Errorf` 字符串，Handler 按字符串匹配（`err.Error() == "too many attempts"`），其余一律 ErrUserExists/ErrPasswordWrong
- 影响：错误码不可靠，部分场景 500 而非业务码

## 一致项 ✅
- Register/Login/RefreshToken 骨架、bcrypt cost 12、登录失败统一文案（dummyHash 防枚举）✅
- User domain 字段（含 SessionVersion）与设计一致 ✅
