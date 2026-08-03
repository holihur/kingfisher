# MySQL（Database） — 设计与实现差异

> 来源：`design/backend/mysql/design.md` 对照 `core/database/gorm.go`
> 排查日期：2026-07-31

## P1

### DB-1 ✅ 连接重试未实现
  **Status: ✅ Retry logic added**
- 设计：MySQL/PostgreSQL 连接应带重试（依赖未就绪时重试）
- 实现：`NewDatabase` 单次 `gorm.Open`，失败直接返回 error
- 影响：docker-compose 中 MySQL 启动慢于服务时首次启动失败（依赖 `depends_on` 顺序）

### DB-2 ✅ `RunMigrations` 未实现
  **Status: ✅ AutoMigrate for SQLite**
- 设计：`core/database` 对外接口包含 `RunMigrations(db, migrationsPath) error`（开发环境自动执行，生产手动）
- 实现：`core/database/gorm.go` 只有 `NewDatabase`/`InitDatabase`/`SeedData`/`AutoMigrate`，无 `RunMigrations`；`main.go` 用 `AutoMigrate`（仅 SQLite）+ 手动 `SeedData`
- 影响：M4 验收「迁移管理」缺失；MySQL/PG 无迁移载体（见 migration 模块）

## P2

### DB-3 ✅ SQLite `SetMaxOpenConns` 未确认
  **Status: ✅ SQLite single-writer configured**
- 设计：SQLite 需 `SetMaxOpenConns(1)`（单写者）
- 实现：SQLite 分支仅打开连接，未显式设置连接池大小
- 影响：SQLite 并发写可能报 database is locked（WAL + busy_timeout 部分缓解）

## 一致项 ✅
- 三驱动（sqlite/mysql/postgres）switch 已实现，DSN 构造与设计一致
- `InitDatabase` + 连接池参数（MaxIdle/MaxOpen/MaxLifetime）已接入
