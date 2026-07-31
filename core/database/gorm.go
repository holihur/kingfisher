package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"kingfisher/core/config"
)

// NewDatabase creates a database connection based on driver type.
func NewDatabase(cfg config.DatabaseConfig, logger *zap.Logger) (*gorm.DB, error) {
	gl := NewGormLogger(logger, cfg.Driver == "sqlite")
	gc := &gorm.Config{
		Logger: gl,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	switch cfg.Driver {
	case "sqlite":
		path := cfg.SQLite.Path
		if path == "" {
			path = "kingfisher.db"
		}
		dsn := path + "?_journal_mode=WAL&_busy_timeout=5000"
		db, err := gorm.Open(sqlite.Open(dsn), gc)
		if err != nil {
			return nil, fmt.Errorf("sqlite open: %w", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("sqlite underlying db: %w", err)
		}
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
		return db, nil

	case "mysql":
		return nil, nil // placeholder for driver
	case "postgres":
		return nil, nil // placeholder for driver
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}

// InitDatabase creates connection, runs AutoMigrate + Seed for SQLite mode.
func InitDatabase(cfg config.DatabaseConfig, logger *zap.Logger) (*gorm.DB, error) {
	db, err := NewDatabase(cfg, logger)
	if err != nil {
		return nil, err
	}
	if cfg.Driver == "sqlite" {
		db.Exec("PRAGMA journal_mode=WAL")
		db.Exec("PRAGMA foreign_keys=ON")
		if err := autoMigrate(db); err != nil {
			return nil, fmt.Errorf("automigrate: %w", err)
		}
		// Seed is called after InitDatabase in main.go to avoid circular deps
	}
	return db, nil
}

// autoMigrate creates tables for SQLite dev mode.
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&UserPO{},
		&RolePO{},
		&PermissionPO{},
		&RolePermissionPO{},
		&MenuPO{},
		&RoleMenuPO{},
		&SystemConfigPO{},
		&AuditLogPO{},
	)
}

// SeedData writes initial data for SQLite dev mode. Idempotent.
func SeedData(db *gorm.DB) error {
	// Skip if already seeded
	var count int64
	if err := db.Model(&UserPO{}).Count(&count).Error; err == nil && count > 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		// Permissions
		perms := []PermissionPO{
			{ID: 1, Name: "查看用户", Code: "user:list", Resource: "user", Action: "read"},
			{ID: 2, Name: "创建用户", Code: "user:create", Resource: "user", Action: "create"},
			{ID: 3, Name: "更新用户", Code: "user:update", Resource: "user", Action: "update"},
			{ID: 4, Name: "删除用户", Code: "user:delete", Resource: "user", Action: "delete"},
			{ID: 5, Name: "查看菜单", Code: "menu:list", Resource: "menu", Action: "read"},
			{ID: 6, Name: "创建菜单", Code: "menu:create", Resource: "menu", Action: "create"},
			{ID: 7, Name: "更新菜单", Code: "menu:update", Resource: "menu", Action: "update"},
			{ID: 8, Name: "删除菜单", Code: "menu:delete", Resource: "menu", Action: "delete"},
			{ID: 9, Name: "查看角色", Code: "role:list", Resource: "role", Action: "read"},
			{ID: 10, Name: "创建角色", Code: "role:create", Resource: "role", Action: "create"},
			{ID: 11, Name: "更新角色", Code: "role:update", Resource: "role", Action: "update"},
			{ID: 12, Name: "删除角色", Code: "role:delete", Resource: "role", Action: "delete"},
			{ID: 13, Name: "查看配置", Code: "config:list", Resource: "config", Action: "read"},
			{ID: 14, Name: "更新配置", Code: "config:update", Resource: "config", Action: "update"},
			{ID: 15, Name: "查看审计", Code: "audit:list", Resource: "audit", Action: "read"},
		}
		if err := tx.Create(&perms).Error; err != nil {
			return fmt.Errorf("seed permissions: %w", err)
		}

		// Roles
		roles := []RolePO{
			{ID: 1, Name: "超级管理员", Code: "admin", Level: 0, Status: 1},
			{ID: 3, Name: "编辑", Code: "editor", Level: 1, Status: 1},
			{ID: 4, Name: "访客", Code: "viewer", Level: 2, Status: 1},
		}
		if err := tx.Create(&roles).Error; err != nil {
			return fmt.Errorf("seed roles: %w", err)
		}

		// Role-Permission links
		type RP struct{ RoleID, PermissionID uint }
		rp := []RP{
			{1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8},
			{1, 9}, {1, 10}, {1, 11}, {1, 12}, {1, 13}, {1, 14}, {1, 15},
			{3, 1}, {3, 2}, {3, 3}, {3, 5}, {3, 6}, {3, 7}, {3, 9}, {3, 13},
			{4, 1}, {4, 5}, {4, 9}, {4, 13},
		}
		if err := tx.Table("role_permissions").Create(&rp).Error; err != nil {
			return fmt.Errorf("seed role_permissions: %w", err)
		}

		// Menus
		menus := []MenuPO{
			{ID: 1, ParentID: 0, Name: "Dashboard", Path: "/dashboard", Component: "pages/Dashboard", Icon: "DashboardOutlined", Sort: 0, Type: 2},
			{ID: 2, ParentID: 0, Name: "系统管理", Path: "/system", Icon: "SettingOutlined", Sort: 1, Type: 1},
			{ID: 3, ParentID: 2, Name: "用户管理", Path: "/system/users", Component: "pages/User/UserList", Icon: "UserOutlined", Sort: 1, Type: 2, Permission: "user:list"},
			{ID: 4, ParentID: 3, Name: "新增用户", Sort: 1, Type: 3, Permission: "user:create"},
			{ID: 5, ParentID: 3, Name: "编辑用户", Sort: 2, Type: 3, Permission: "user:update"},
			{ID: 6, ParentID: 3, Name: "删除用户", Sort: 3, Type: 3, Permission: "user:delete"},
			{ID: 7, ParentID: 2, Name: "菜单管理", Path: "/system/menus", Component: "pages/Menu/MenuManage", Icon: "MenuOutlined", Sort: 2, Type: 2, Permission: "menu:list"},
			{ID: 8, ParentID: 7, Name: "新增菜单", Sort: 1, Type: 3, Permission: "menu:create"},
			{ID: 9, ParentID: 7, Name: "编辑菜单", Sort: 2, Type: 3, Permission: "menu:update"},
			{ID: 10, ParentID: 7, Name: "删除菜单", Sort: 3, Type: 3, Permission: "menu:delete"},
			{ID: 11, ParentID: 2, Name: "角色管理", Path: "/system/roles", Component: "pages/Role/RoleList", Icon: "SafetyOutlined", Sort: 3, Type: 2, Permission: "role:list"},
			{ID: 12, ParentID: 11, Name: "新增角色", Sort: 1, Type: 3, Permission: "role:create"},
			{ID: 13, ParentID: 11, Name: "编辑角色", Sort: 2, Type: 3, Permission: "role:update"},
			{ID: 14, ParentID: 11, Name: "删除角色", Sort: 3, Type: 3, Permission: "role:delete"},
			{ID: 15, ParentID: 2, Name: "系统配置", Path: "/system/configs", Component: "pages/Config/ConfigManage", Icon: "ControlOutlined", Sort: 4, Type: 2, Permission: "config:list"},
		}
		if err := tx.Create(&menus).Error; err != nil {
			return fmt.Errorf("seed menus: %w", err)
		}

		// Role-Menu links
		type RM struct{ RoleID, MenuID uint }
		rm := []RM{
			{1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8},
			{1, 9}, {1, 10}, {1, 11}, {1, 12}, {1, 13}, {1, 14}, {1, 15},
			{3, 1}, {3, 3}, {3, 7},
			{4, 1},
		}
		if err := tx.Table("role_menus").Create(&rm).Error; err != nil {
			return fmt.Errorf("seed role_menus: %w", err)
		}

		// System Configs
		configs := []SystemConfigPO{
			{Key: "site_name", Value: "Kingfisher Admin", Remark: "系统名称"},
			{Key: "site_logo", Value: "/logo.png", Remark: "Logo"},
			{Key: "max_login_attempts", Value: "5", Remark: "最大登录失败次数"},
			{Key: "lockout_duration", Value: "15m", Remark: "锁定时间"},
			{Key: "session_timeout", Value: "30m", Remark: "会话超时"},
		}
		if err := tx.Create(&configs).Error; err != nil {
			return fmt.Errorf("seed configs: %w", err)
		}

		// Admin user
		admin := UserPO{
			ID:       1,
			Username: "admin",
			Password: "$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO",
			Email:    "admin@example.com",
			Status:   1,
			RoleID:   1,
		}
		return tx.Create(&admin).Error
	})
}

// NewGormLogger creates a Zap-based GORM logger
func NewGormLogger(logger *zap.Logger, debug bool) gormlogger.Interface {
	lvl := gormlogger.Warn
	if debug {
		lvl = gormlogger.Info
	}
	return gormlogger.New(
		&zapWriter{logger: logger},
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  lvl,
			IgnoreRecordNotFoundError: true,
			Colorful:                  debug,
		},
	)
}

type zapWriter struct {
	logger *zap.Logger
}

func (w *zapWriter) Printf(format string, args ...any) {
	w.logger.Info(fmt.Sprintf(format, args...))
}

var _ context.Context = nil // import context
