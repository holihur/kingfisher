# Transaction — 设计与实现差异

> 来源：`design/backend/transaction/design.md` 对照 `core/database/`、extends repos
> 排查日期：2026-07-31

## P1

### TX-1 UnitOfWork 未实现
- 设计：`core/database/unit_of_work.go` 定义 `UnitOfWork` 接口 + `GormUnitOfWork` 实现（tx 注入 context，repo `getDB(ctx)` 取）
- 实现：无 `unit_of_work.go`；无 `txKey` 注入；无 `getDB(ctx)` 模式
- 影响：跨 Repository 事务无法保证原子性（如「创建用户+分配角色」类操作）

### TX-2 现有事务仅两处且不规范
- 设计：Service 层控制事务边界
- 实现：仅 `rbac.AssignPermissions`/`AssignMenus` 用 `gorm.Transaction`（repo 内部），其余写操作单条 SQL 隐式提交
- 影响：菜单/角色/配置删除等组合操作无回滚保护

## 一致项 ✅
- GORM 本身支持 `WithContext` 事务语义（未来接入成本低）
