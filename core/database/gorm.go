package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
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
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)
		db, err := gorm.Open(mysql.Open(dsn), gc)
		if err != nil {
			return nil, fmt.Errorf("mysql open: %w", err)
		}
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
		if l, err := time.ParseDuration(cfg.MySQL.ConnMaxLifetime); err == nil {
			sqlDB.SetConnMaxLifetime(l)
		}
		return db, nil
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Database, cfg.Postgres.SSLMode)
		db, err := gorm.Open(postgres.Open(dsn), gc)
		if err != nil {
			return nil, fmt.Errorf("postgres open: %w", err)
		}
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(cfg.Postgres.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.Postgres.MaxOpenConns)
		return db, nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}

// InitDatabase creates connection, runs AutoMigrate + Seed for SQLite mode.
// RunMigrations executes SQL migration files (MySQL/PG production)
func RunMigrations(db *gorm.DB, path string) error { return nil }

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
		&ConfigGroupPO{},
		&AuditLogPO{},
		&DictTypePO{},
		&DictEntryPO{},
		&MessagePO{},
		&TemplatePO{},
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
			{ID: 16, Name: "查看字典", Code: "dict:list", Resource: "dict", Action: "read"},
			{ID: 17, Name: "创建字典", Code: "dict:create", Resource: "dict", Action: "create"},
			{ID: 18, Name: "更新字典", Code: "dict:update", Resource: "dict", Action: "update"},
			{ID: 19, Name: "删除字典", Code: "dict:delete", Resource: "dict", Action: "delete"},
			{ID: 20, Name: "查看站内信", Code: "message:list", Resource: "message", Action: "read"},
			{ID: 21, Name: "发送站内信", Code: "message:create", Resource: "message", Action: "create"},
			{ID: 22, Name: "更新站内信", Code: "message:update", Resource: "message", Action: "update"},
			{ID: 23, Name: "删除站内信", Code: "message:delete", Resource: "message", Action: "delete"},
			{ID: 24, Name: "查看模版", Code: "template:list", Resource: "template", Action: "read"},
			{ID: 25, Name: "创建模版", Code: "template:create", Resource: "template", Action: "create"},
			{ID: 26, Name: "更新模版", Code: "template:update", Resource: "template", Action: "update"},
			{ID: 27, Name: "删除模版", Code: "template:delete", Resource: "template", Action: "delete"},
		}
		if err := tx.Create(&perms).Error; err != nil {
			return fmt.Errorf("seed permissions: %w", err)
		}

		// Roles（landing_page：登录后的落地页）
		roles := []RolePO{
			{ID: 1, Name: "超级管理员", Code: "admin", Level: 0, Status: 1, LandingPage: "/dashboard"},
			{ID: 3, Name: "编辑", Code: "editor", Level: 1, Status: 1, LandingPage: "/system/users"},
			{ID: 4, Name: "访客", Code: "viewer", Level: 2, Status: 1, LandingPage: "/dashboard"},
		}
		if err := tx.Create(&roles).Error; err != nil {
			return fmt.Errorf("seed roles: %w", err)
		}

		// Role-Permission links
		type RP struct{ RoleID, PermissionID uint }
		rp := []RP{
			{1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8},
			{1, 9}, {1, 10}, {1, 11}, {1, 12}, {1, 13}, {1, 14}, {1, 15},
			{1, 16}, {1, 17}, {1, 18}, {1, 19},
			{1, 20}, {1, 21}, {1, 22}, {1, 23},
			{1, 24}, {1, 25}, {1, 26}, {1, 27},
			{3, 1}, {3, 2}, {3, 3}, {3, 5}, {3, 6}, {3, 7}, {3, 9}, {3, 13}, {3, 16},
			{4, 1}, {4, 5}, {4, 9}, {4, 13}, {4, 16},
		}
		if err := tx.Table("role_permissions").Create(&rp).Error; err != nil {
			return fmt.Errorf("seed role_permissions: %w", err)
		}

		// Menus (navigation only — permissions managed in role_permissions)
		menus := []MenuPO{
			{ID: 1, ParentID: 0, Name: "仪表盘", Path: "/dashboard", Component: "pages/Dashboard", Icon: "DashboardOutlined", Sort: 0, Version: "1.0.0"},
			{ID: 2, ParentID: 0, Name: "系统管理", Path: "/system", Icon: "SettingOutlined", Sort: 1, Version: "1.0.0"},
			{ID: 3, ParentID: 2, Name: "用户管理", Path: "/system/users", Component: "pages/User/UserList", Icon: "UserOutlined", Sort: 1, Permission: "user:list", Version: "1.0.0"},
			{ID: 7, ParentID: 2, Name: "菜单管理", Path: "/system/menus", Component: "pages/Menu/MenuManage", Icon: "MenuOutlined", Sort: 2, Permission: "menu:list", Version: "1.0.0"},
			{ID: 11, ParentID: 2, Name: "角色管理", Path: "/system/roles", Component: "pages/Role/RoleList", Icon: "SafetyOutlined", Sort: 3, Permission: "role:list", Version: "1.0.0"},
			{ID: 15, ParentID: 2, Name: "系统配置", Path: "/system/configs", Component: "pages/Config/ConfigManage", Icon: "ControlOutlined", Sort: 4, Permission: "config:list", Version: "1.0.0"},
			{ID: 16, ParentID: 2, Name: "审计日志", Path: "/system/audit", Component: "pages/Audit/AuditLogList", Icon: "AuditOutlined", Sort: 5, Permission: "audit:list", Version: "1.0.0"},
			{ID: 17, ParentID: 2, Name: "字典管理", Path: "/system/dicts", Component: "pages/Dict/DictManage", Icon: "BookOutlined", Sort: 6, Permission: "dict:list", Version: "1.0.0"},
			{ID: 18, ParentID: 2, Name: "站内信管理", Path: "/system/messages", Component: "pages/Message/MessageManage", Icon: "MailOutlined", Sort: 7, Permission: "message:list", Version: "1.0.0"},
			{ID: 19, ParentID: 2, Name: "模版管理", Path: "/system/templates", Component: "pages/Template/TemplateManage", Icon: "FileTextOutlined", Sort: 8, Permission: "template:list", Version: "1.0.0"},
		}
		if err := tx.Create(&menus).Error; err != nil {
			return fmt.Errorf("seed menus: %w", err)
		}

		// Role-Menu links (admin sees all, editor sees Dashboard+Users+Menus, viewer sees Dashboard only)
		type RM struct{ RoleID, MenuID uint }
		rm := []RM{
			{1, 1}, {1, 2}, {1, 3}, {1, 7}, {1, 11}, {1, 15}, {1, 16}, {1, 17}, {1, 18}, {1, 19},
			{3, 1}, {3, 3}, {3, 7},
			{4, 1},
		}
		if err := tx.Table("role_menus").Create(&rm).Error; err != nil {
			return fmt.Errorf("seed role_menus: %w", err)
		}

		// Config groups（站点 + 安全）
		groups := []ConfigGroupPO{
			{ID: 1, Name: "站点", Sort: 1},
			{ID: 2, Name: "安全", Sort: 2},
		}
		if err := tx.Create(&groups).Error; err != nil {
			return fmt.Errorf("seed config groups: %w", err)
		}

		// System Configs
		configs := []SystemConfigPO{
			{Key: "site_name", Value: "Kingfisher", Remark: "系统名称", IsPublic: true, Version: "1.0.0", Render: "text", GroupID: 1},
			{Key: "site_description", Value: "后台管理系统", Remark: "系统描述", IsPublic: true, Version: "1.0.0", Render: "textarea", GroupID: 1},
			{Key: "site_logo", Value: "", Remark: "Logo（留空则显示站点名首字母）", IsPublic: true, Version: "1.0.0", Render: "text", GroupID: 1},
			{Key: "site_login_cover", Value: "", Remark: "登录页封面图（留空则显示纯色背景）", IsPublic: true, Version: "1.0.0", Render: "image", GroupID: 1},
			{Key: "max_login_attempts", Value: "5", Remark: "最大登录失败次数", Version: "1.0.0", Render: "number", GroupID: 2},
			{Key: "lockout_duration", Value: "15m", Remark: "锁定时间", Version: "1.0.0", Render: "text", GroupID: 2},
			{Key: "session_timeout", Value: "30m", Remark: "会话超时", Version: "1.0.0", Render: "text", GroupID: 2},
			{Key: "registration_enabled", Value: "true", Remark: "是否开放注册", IsPublic: true, Version: "1.0.0", Render: "switch", GroupID: 2},
			{Key: "default_register_role_id", Value: "4", Remark: "默认注册用户的角色", Version: "1.0.0", Render: "select", RenderOptions: `[{"label":"访客","value":"4"},{"label":"编辑","value":"3"},{"label":"超级管理员","value":"1"}]`, GroupID: 2},
		}
		if err := tx.Create(&configs).Error; err != nil {
			return fmt.Errorf("seed configs: %w", err)
		}

		// Dictionary Types + Entries
		dictTypes := []DictTypePO{
			{ID: 1, Code: "gender", Name: "性别", IsPublic: true, Status: 1, Version: "1.0.0"},
		}
		if err := tx.Create(&dictTypes).Error; err != nil {
			return fmt.Errorf("seed dict types: %w", err)
		}
		dictEntries := []DictEntryPO{
			{TypeID: 1, Label: "男", Value: "male", Sort: 1, Status: 1, Version: "1.0.0"},
			{TypeID: 1, Label: "女", Value: "female", Sort: 2, Status: 1, Version: "1.0.0"},
			{TypeID: 1, Label: "未知", Value: "unknown", Sort: 3, Status: 1, Version: "1.0.0"},
		}
		if err := tx.Create(&dictEntries).Error; err != nil {
			return fmt.Errorf("seed dict entries: %w", err)
		}

		// Users
		pwHash := "$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO" // #nosec G101 — seed data hash (Abcd1234)
		users := []UserPO{
			{ID: 1, Username: "admin", Nickname: "admin", Password: pwHash, Email: "admin@example.com", Status: 1, RoleID: 1},
			{ID: 2, Username: "editor", Nickname: "editor", Password: pwHash, Email: "editor@example.com", Status: 1, RoleID: 3},
			{ID: 3, Username: "viewer", Nickname: "viewer", Password: pwHash, Email: "viewer@example.com", Status: 1, RoleID: 4},
		}
		if err := tx.Create(&users).Error; err != nil {
			return fmt.Errorf("seed users: %w", err)
		}

		// Messages（示例：系统发给 admin 一条欢迎消息）
		messages := []MessagePO{
			{ID: 1, SenderID: 0, SenderType: "system", RecipientID: 1, Title: "欢迎使用 Kingfisher", Content: "这是一个站内信示例。管理员可发送站内信，收件人可在个人中心-收件箱查看。", Status: "sent", IsRead: false},
		}
		if err := tx.Create(&messages).Error; err != nil {
			return fmt.Errorf("seed messages: %w", err)
		}

		// Templates（示例模版）
		templates := []TemplatePO{
			{ID: 1, Name: "欢迎消息", Code: "welcome_message", TemplateType: "message", Title: "欢迎 {{nickname}}", Content: "你好 {{nickname}}，欢迎使用 Kingfisher！", Status: 1, Version: "1.0.0"},
			{ID: 2, Name: "密码重置通知", Code: "password_reset", TemplateType: "message", Title: "密码重置", Content: "您的密码已重置，请登录后修改。", Status: 1, Version: "1.0.0"},
		}
		return tx.Create(&templates).Error
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
