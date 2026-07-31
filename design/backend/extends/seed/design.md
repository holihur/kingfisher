# Seed — 种子数据（Go 代码）

## 职责

SQLite 开发模式下通过 Go 函数写入初始数据（用户/角色/权限/菜单/配置）。MySQL/PG 生产模式由 `migrations/000008_seed_data.up.sql` 负责，不重复执行此函数。

## 对外接口

```go
// internal/infra/seed.go
func Seed(db *gorm.DB) error
```

`Seed` 按依赖顺序写入：permissions → roles → role_permissions → menus → role_menus → configs → admin user。所有写入在同一事务中，任一步失败则全部回滚。

## 核心逻辑

```go
func Seed(db *gorm.DB) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // 1. 权限 —— 15 条（与 migrations/000008 一致）
        perms := []Permission{
            {ID: 1, Name: "查看用户", Code: "user:list", Resource: "user", Action: "read"},
            {ID: 2, Name: "创建用户", Code: "user:create", Resource: "user", Action: "create"},
            // ... 共 15 条
            {ID: 15, Name: "查看审计", Code: "audit:list", Resource: "audit", Action: "read"},
        }
        if err := tx.Create(&perms).Error; err != nil { return err }

        // 2. 角色 —— 3 个
        roles := []Role{
            {ID: 1, Name: "超级管理员", Code: "admin", Level: 0, Status: 1},
            {ID: 3, Name: "编辑", Code: "editor", Level: 1, Status: 1},
            {ID: 4, Name: "访客", Code: "viewer", Level: 2, Status: 1},
        }
        if err := tx.Create(&roles).Error; err != nil { return err }

        // 3. 角色-权限关联
        type RolePermission struct {
            RoleID       uint
            PermissionID uint
        }
        rp := []RolePermission{
            // admin 全部 15 权限
            {1,1},{1,2},{1,3},{1,4},{1,5},{1,6},{1,7},{1,8},
            {1,9},{1,10},{1,11},{1,12},{1,13},{1,14},{1,15},
            // editor 8 权限
            {3,1},{3,2},{3,3},{3,5},{3,6},{3,7},{3,9},{3,13},
            // viewer 4 权限
            {4,1},{4,5},{4,9},{4,13},
        }
        if err := tx.Create(&rp).Error; err != nil { return err }

        // 4. 菜单 —— 15 条
        menus := []Menu{
            {ID: 1, ParentID: 0, Name: "Dashboard", Path: "/dashboard", Component: "pages/Dashboard", Icon: "DashboardOutlined", Sort: 0, Type: 2},
            {ID: 2, ParentID: 0, Name: "系统管理", Path: "/system", Icon: "SettingOutlined", Sort: 1, Type: 1},
            // ... 共 15 条
        }
        if err := tx.Create(&menus).Error; err != nil { return err }

        // 5. 角色-菜单关联
        type RoleMenu struct {
            RoleID uint
            MenuID uint
        }
        rm := []RoleMenu{
            {1,1},{1,2},{1,3},{1,4},{1,5},{1,6},{1,7},{1,8},
            {1,9},{1,10},{1,11},{1,12},{1,13},{1,14},{1,15},
            {3,1},{3,3},{3,7},
            {4,1},
        }
        if err := tx.Create(&rm).Error; err != nil { return err }

        // 6. 系统配置 —— 5 项
        configs := []SystemConfig{
            {Key: "site_name", Value: "Kingfisher Admin", Remark: "系统名称"},
            {Key: "site_logo", Value: "/logo.png", Remark: "Logo"},
            {Key: "max_login_attempts", Value: "5", Remark: "最大登录失败次数"},
            {Key: "lockout_duration", Value: "15m", Remark: "锁定时间"},
            {Key: "session_timeout", Value: "30m", Remark: "会话超时"},
        }
        if err := tx.Create(&configs).Error; err != nil { return err }

        // 7. 管理员 —— 密码 Abcd1234 的 bcrypt hash（cost=12）
        admin := User{
            ID: 1, Username: "admin",
            Password: "$2a$12$LJ3m4ys3Lk0TSwHCpNqrIeN5U5Akn5dQUhBvPXFxFG7GqQvHCzB5q",
            Email: "admin@example.com", Status: 1, RoleID: 1,
        }
        return tx.Create(&admin).Error
    })
}
```

## 调用位置

```go
// core/database/gorm.go — InitDatabase 中
if cfg.Driver == "sqlite" {
    db.AutoMigrate(...)
    if err := seed.Seed(db); err != nil {
        logger.Fatal("seed failed", zap.Error(err))
    }
}
```

种子函数幂等（`ID` 明确赋值）。重复执行 → UNIQUE 冲突 → 事务回滚 → 日志 warn 而非 Fatal。

## 与 SQL 种子的关系

| 模式 | Driver | 种子来源 | 文件 |
|------|--------|----------|------|
| 开发 | SQLite | Go `seed.Seed(db)` | `internal/infra/seed.go` |
| 生产 | MySQL/PG | SQL `migrations/000008` | `migrations/000008_seed_data.up.sql` |

两套种子数据**维护同一份数据源**，变更时需同时更新 `internal/infra/seed.go` 和 `migrations/000008`。设计文档的 migration/design.md 为 MySQL seed SQL 的权威源。
