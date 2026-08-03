# Port Interface — 设计与实现差异

> 来源：`design/backend/interface/design.md` 对照 `extends/*/port/`
> 排查日期：2026-07-31

## P1

### IF-1 UserRepository 签名与设计不一致
  **Status: ⚠️ FindAll signature acceptable**
- 设计：`FindAll(ctx, page, pageSize int, filters map[string]any)`
- 实现：`FindAll(ctx, page, pageSize int, keyword string)`（`extends/user/port/repository.go`）
- 影响：设计支持的任意过滤字段（状态/角色等）无法表达，仅关键词搜索

### IF-2 MenuRepository 接口缺失
  **Status: ✅ MenuRepository created**
- 设计：`extends/menu/port/repository.go` 定义 MenuRepository
- 实现：`extends/menu/` 下无 port 目录；`MenuService` 直接依赖 `*adapter.MenuRepo`（具体实现）
- 影响：违反依赖反转；菜单模块单测必须连真实 DB

### IF-3 ConfigRepository 接口缺失
  **Status: ✅ ConfigRepository created**
- 设计：`extends/config/port/repository.go` 定义 ConfigRepository
- 实现：`extends/config/` 无 port；`ConfigService` 依赖 `*adapter.ConfigRepo`
- 影响：同 IF-2

### IF-4 AuditRepository 接口缺失
  **Status: ✅ AuditRepository created**
- 设计：`extends/audit/port/repository.go` 定义 AuditRepository
- 实现：`extends/audit/` 无 port；`AuditService` 依赖 `*adapter.AuditRepo`
- 影响：同 IF-2

### IF-5 设计与实现均有缺口
  **Status: ⚠️ Service interfaces deferred**
- 设计：`port/service.go`（Service 接口）在各模块
- 实现：仅 user 有 `port/repository.go`；rbac 有 `port/repository.go`；无任何 `port/service.go`
- 影响：Handler 依赖具体 Service（见 SI-1）

## 一致项 ✅
- `RoleRepository`（含 AssignPermissions/GetPermissions/AssignMenus/GetMenus）接口已按设计实现于 `extends/rbac/port/repository.go`
- user 模块 `port.UserRepository` 存在（差异仅 FindAll 签名）
