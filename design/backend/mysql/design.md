# Database — 多数据库驱动

## 职责

GORM 初始化 + SQL 迁移管理，支持 **SQLite / MySQL / PostgreSQL** 三驱动切换。配置一行 `database.driver: sqlite` 即换。

## 对外接口

```go
func NewDatabase(cfg DatabaseConfig, logger *zap.Logger) (*gorm.DB, error)
func RunMigrations(db *gorm.DB, migrationsPath string) error
```

`NewDatabase` 根据 `cfg.Driver` 自动选择驱动、构造 DSN、打开连接。调用方不关心底层是 SQLite 还是 MySQL——返回的都是 `*gorm.DB`。

## 核心逻辑

### NewDatabase

```
switch cfg.Driver {
case "sqlite":
    1. path := cfg.SQLite.Path，若为空则默认 "kingfisher.db"
    2. dsn := path + "?_journal_mode=WAL&_busy_timeout=5000"
    3. gorm.Open(sqlite.Open(dsn), &gorm.Config{...})
    4. sqlDB.SetMaxOpenConns(1)  // SQLite 单写者，连接池=1

case "mysql":
    1. dsn := "user:pass@tcp(host:port)/db?charset=utf8mb4&parseTime=True&loc=Local"
    2. gorm.Open(mysql.Open(dsn), &gorm.Config{...})
    3. sqlDB.SetMaxIdleConns(cfg.MaxIdle)
    4. sqlDB.SetMaxOpenConns(cfg.MaxOpen)
    5. sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

case "postgres":
    1. dsn := "host=... port=... user=... password=... dbname=... sslmode=..."
    2. gorm.Open(postgres.Open(dsn), &gorm.Config{...})
    3. sqlDB.SetMaxIdleConns(cfg.MaxIdle)
    4. sqlDB.SetMaxOpenConns(cfg.MaxOpen)
}
```

GORM 统一接口：所有驱动共用 `&gorm.Config{Logger: ..., NowFunc: UTC}`。返回 `*gorm.DB` 后，上层 extends 模块完全无感。

## 驱动导入

```go
// core/database/gorm.go
import (
    "gorm.io/driver/mysql"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)
```

三个驱动全部 import——Go compiler 会根据实际调用路径 tree-shake 未用驱动（需确认：实际 runtime 不会加载未用驱动，但 binary 会包含）。若在意 binary 大小，可用 build tags 分别编译。

## 开发体验（SQLite 模式）

```
$ make run
# → database.driver=sqlite
# → 自动创建 kingfisher.db（项目根目录）
# → GORM AutoMigrate 自动建表（非 migrations/*.sql）
# → 种子数据通过 Go 代码写入
# → 无需 Docker、无需安装 MySQL
# → kingfisher.db 加入 .gitignore

$ sqlite3 kingfisher.db "SELECT * FROM users;"
# → 1|admin|$2a$12$...|admin@example.com|1
```

## Adapter 层：PO/Domain 分离（驱动无关）

```go
// adapter/mysql/user_model.go — GORM PO，所有驱动通用
type userPO struct {
    ID        uint
    Username  string
    Password  string
    Email     string
    Status    int
    RoleID    uint
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt
}
func (userPO) TableName() string { return "users" }
```

GORM 根据实际驱动自动翻译 Go struct tag → DDL 语句（`AUTO_INCREMENT` vs `AUTOINCREMENT` vs `SERIAL`）。PO 层零改动。

## 迁移兼容性

| DDL 特性 | SQLite | MySQL | PostgreSQL |
|----------|--------|-------|------------|
| `AUTO_INCREMENT` / `AUTOINCREMENT` | ✅ GORM 自动 | ✅ | ✅ `SERIAL` |
| `VARCHAR(n)` | ✅ | ✅ | ✅ |
| `TEXT` / `JSON` | ✅ | ✅ | ✅ `JSONB` |
| `DATETIME(3)` | ⚠️ 存为 TEXT | ✅ | ✅ |
| `ALTER COLUMN` | ❌ SQLite 不支持 | ✅ | ✅ |
| `ADD COLUMN ... AFTER` | ❌ | ✅ | ❌ |
| Foreign Key | ⚠️ 需 `PRAGMA foreign_keys=ON` | ✅ | ✅ |
| Collation (`utf8mb4_unicode_ci`) | ❌ | ✅ | ❌ |

设计决策：**迁移 SQL 写 MySQL 方言（utf8mb4 + InnoDB），SQLite 开发时用 GORM AutoMigrate 兼容模式**。生产 PG 部署时需提供 PG 版本的迁移文件（或使用 GORM AutoMigrate）。

```go
// 开发环境 (SQLite) → GORM AutoMigrate
if cfg.Driver == "sqlite" && cfg.Server.Mode != "release" {
    db.AutoMigrate(&userPO{}, &rolePO{}, ...)
}
// 生产环境 (MySQL/PG) → golang-migrate SQL
```

## 索引策略（同 round1，驱动兼容）

| 表 | 索引 | 类型 |
|------|------|------|
| users | uk_username | UNIQUE |
| roles | uk_code | UNIQUE |
| menus | idx_parent_id | INDEX |
| configs | uk_key | UNIQUE |

## 配置

```yaml
# 开发 —— 零依赖
database:
  driver: sqlite
  sqlite:
    path: kingfisher.db

# 生产 —— MySQL
# database:
#   driver: mysql
#   mysql:
#     host: ${MYSQL_HOST}
#     port: 3306
#     database: kingfisher
#     max_idle_conns: 10
#     max_open_conns: 100

# 生产 —— PostgreSQL
# database:
#   driver: postgres
#   postgres:
#     host: ${PG_HOST}
#     port: 5432
#     database: kingfisher
#     sslmode: require
```

## 设计要点

- **GORM 做抽象层**——所有 extends 只依赖 `*gorm.DB`，切换驱动不改业务代码
- **SQLite 开发零依赖**——新 clone 直接 `make run`，不需要 Docker/MySQL
- **迁移分两套**：开发 SQLite → AutoMigrate；生产 MySQL/PG → golang-migrate SQL
- **每个驱动的连接池参数不同**：SQLite=1（单写者）、MySQL=100（高并发）
- `kingfisher.db` 加入 `.gitignore`，不提交数据库文件
