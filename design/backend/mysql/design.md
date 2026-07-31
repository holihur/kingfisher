# MySQL — 数据库 & 迁移

## 职责

GORM 初始化 + SQL 迁移管理。**不使用 AutoMigrate**，正式环境用纯 SQL 迁移。

## 对外接口

```go
func NewMySQL(cfg MySQLConfig, logger *zap.Logger) (*gorm.DB, error)
func RunMigrations(db *gorm.DB, migrationsPath string) error
```

## 核心逻辑

### NewMySQL

```
1. dsn := "user:pass@tcp(host:port)/db?charset=utf8mb4&parseTime=True&loc=Local"
2. gorm.Open(mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{
       Logger: NewGormLogger(logger),       // 适配 Zap
       NowFunc: func() time.Time { return time.Now().UTC() },  // UTC
   })
3. sqlDB.SetMaxIdleConns(cfg.MaxIdle)
4. sqlDB.SetMaxOpenConns(cfg.MaxOpen)
5. sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
6. sqlDB.PingContext(ctx)
```

### RunMigrations

```
1. driver, _ := mysql.WithInstance(db, &mysql.Config{})
2. m, _ := migrate.NewWithDatabaseInstance("file://migrations", "mysql", driver)
3. m.Up() // 若为开发环境；生产环境提示手动执行
```

## 迁移文件示例

```sql
-- migrations/000001_create_users.up.sql
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(32) NOT NULL,
    password VARCHAR(128) NOT NULL,
    email VARCHAR(128) DEFAULT '',
    avatar VARCHAR(255) DEFAULT '',
    status TINYINT DEFAULT 1,
    role_id BIGINT UNSIGNED DEFAULT 0,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) DEFAULT NULL,
    UNIQUE INDEX idx_username (username),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

```sql
-- migrations/000001_create_users.down.sql
DROP TABLE IF EXISTS users;
```

## Adapter 层：用 GORM Model 适配 domain

```go
// adapter/mysql/user_model.go — GORM 专用
type userPO struct {          // PO = Persistent Object
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

// adapter/mysql/user_repo.go — 实现 port.UserRepository
type UserRepo struct { DB *gorm.DB }

func (r *UserRepo) FindByID(ctx context.Context, id uint) (*domain.User, error) {
    var po userPO
    err := r.DB.WithContext(ctx).First(&po, id).Error
    if err != nil { return nil, err }
    return po.toDomain(), nil    // PO → Domain 转换
}

func (p userPO) toDomain() *domain.User {
    return &domain.User{
        ID: p.ID, Username: p.Username, /* ... */
    }
}
```

## 索引策略

| 表 | 索引 | 类型 | 理由 |
|------|------|------|------|
| users | idx_username | UNIQUE | 登录唯一查询 |
| users | idx_deleted_at | INDEX | 软删除过滤 |
| menus | idx_parent_id | INDEX | 树形查询 |
| roles | idx_code | UNIQUE | 角色编码唯一 |
| configs | idx_key | UNIQUE | 配置键唯一 |
| role_menus | idx_role_id | INDEX | 角色菜单关联 |
| role_permissions | idx_role_id | INDEX | 角色权限关联 |

## 设计要点

- **PO/Domain 分离** — adapter 层有自己的 GORM model，与 domain 互不污染
- **ctx 传递** — 所有 DB 操作带 `WithContext(ctx)`，支持超时、trace
- **生产不用 AutoMigrate** — 用 `golang-migrate` 或手写 DDL
