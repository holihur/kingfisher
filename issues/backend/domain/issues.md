# Domain — 设计与实现差异

> 来源：`design/backend/domain/design.md` 对照 `extends/*/domain/`
> 排查日期：2026-07-31

## P2

### DOM-1 ✅ User domain 额外含 Role 内嵌结构
  **Status: ✅ Acceptable — see issue detail**
- 设计：`User` 仅含基础字段（无 Role）
- 实现：`extends/user/domain/user.go` 增加 `Role *Role`（id/name/code）用于列表展示
- 影响：增强而非冲突；前端用户列表需要角色列（当前前端未使用，见 FU-1）

## P2

### DOM-2 ✅ 依赖纯度
  **Status: ✅ Clean**
- 设计：domain 零外部依赖（不 import GORM/Gin）
- 实现：所有 domain 包仅 import `time` 与内部包 ✅；但 `extends/menu/domain/menu.go`、`extends/rbac/domain/*` 需确认无框架依赖
- 影响：符合设计，无需修复

## 一致项 ✅
- `User{ID,Username,Password(json:"-"),Email,Avatar,Status,RoleID,SessionVersion(json:"-"),CreatedAt,UpdatedAt}` 与设计一致（含 SessionVersion 隐藏）
- Menu 树形结构（ParentID/Children/Type 1=目录 2=菜单 3=按钮）与设计一致
- Role/Permission/SystemConfig/AuditLog 字段与设计一致
