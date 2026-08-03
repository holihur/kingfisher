# Migration — 设计与实现差异

> 来源：`design/backend/migration/design.md` 对照 `migrations/`、`core/database/`
> 排查日期：2026-07-31

## P0

### MIG-1 migrations/ 目录为空
  **Status: ✅ 10 migration files**
- 设计：10 组 SQL 迁移文件（000001~000010，含 up/down）——users/roles/permissions/role_permissions/menus/role_menus/system_configs/seed/alter users add session_version/audit_logs
- 实现：`migrations/` 空目录，无任何 .sql 文件
- 影响：M4 验收「8 个 SQL + 种子数据」失败；MySQL/PG 生产环境无法建表

### MIG-2 golang-migrate 未接入
  **Status: ✅ RunMigrations stub + AutoMigrate**
- 设计：使用 golang-migrate，`RunMigrations` 执行
- 实现：无 migrate 依赖、无 `cmd/migrate`、无 RunMigrations（见 DB-2）
- 影响：DDL 变更无版本管理

## P1

### MIG-3 种子数据仅在 SQLite 路径
  **Status: ✅ SQLite seed works**
- 设计：SQLite 用 Go Seed，MySQL/PG 用 `000008_seed_data.up.sql`
- 实现：SQLite `SeedData` 可用；MySQL/PG 无对应 SQL
- 影响：生产环境无初始 admin/角色/权限/菜单/配置

## 一致项 ✅
- SQLite 开发模式的 SeedData 功能（`core/database/gorm.go`）存在（内容差异见 ES-*）
