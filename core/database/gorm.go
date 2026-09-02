package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite" // 纯 Go SQLite 驱动（无 cgo），支持静态编译
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
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
		// glebarez/sqlite（modernc 纯 Go）：dsn 用 _pragma 前缀设置 PRAGMA
		dsn := "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
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

// InitDatabase 创建数据库连接并根据驱动选择初始化策略：
//   - sqlite：执行 AutoMigrate（开发模式，自动建表）
//   - mysql/postgres：执行 RunMigrations（生产模式，版本化 SQL 迁移）
func InitDatabase(cfg config.DatabaseConfig, logger *zap.Logger) (*gorm.DB, error) {
	db, err := NewDatabase(cfg, logger)
	if err != nil {
		return nil, err
	}
	switch cfg.Driver {
	case "sqlite":
		db.Exec("PRAGMA journal_mode=WAL")
		db.Exec("PRAGMA foreign_keys=ON")
		if err := autoMigrate(db); err != nil {
			return nil, fmt.Errorf("automigrate: %w", err)
		}
		// Seed 由 main.go 在 InitDatabase 之后调用，避免循环依赖
	case "mysql", "postgres":
		if err := RunMigrations(db, "migrations"); err != nil {
			// 尝试从可执行文件所在目录解析 migrations（兼容 Docker / 不同 cwd）
			if alt := resolveMigrationsDir(); alt != "" && alt != "migrations" {
				if altErr := RunMigrations(db, alt); altErr == nil {
					break
				}
			}
			return nil, fmt.Errorf("migrations: %w", err)
		}
	}
	return db, nil
}

// resolveMigrationsDir 尝试定位 migrations 目录，兼容不同 cwd / Docker 布局。
func resolveMigrationsDir() string {
	candidates := []string{
		"migrations",
		filepath.Join(filepath.Dir(os.Args[0]), "migrations"),
		filepath.Join(filepath.Dir(os.Args[0]), "../migrations"),
	}
	for _, p := range candidates {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return p
		}
	}
	return ""
}

// autoMigrate creates tables for SQLite dev mode.
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&UserPO{},
		&RolePO{},
		&PermissionPO{},
		&RolePermissionPO{},
		&RoleDataScopePO{},
		&MenuPO{},
		&RoleMenuPO{},
		&UserRolePO{},
		&SystemConfigPO{},
		&ConfigGroupPO{},
		&AuditLogPO{},
		&DictTypePO{},
		&DictEntryPO{},
		&MessagePO{},
		&TemplatePO{},
		&ScheduledTaskPO{},
		&WorkTaskPO{},
		&DocDirectoryPO{},
		&DocDirRolePO{},
		&DocumentPO{},
		&DocVersionPO{},
		&DepartmentPO{},
		&UserDepartmentPO{},
		&DepartmentRolePO{},
		&AgentConversationPO{},
		&AgentMessagePO{},
	)
}

// SeedData writes initial data for SQLite dev mode. Idempotent.
func SeedData(db *gorm.DB) error {
	// Skip if already seeded
	var count int64
	if err := db.Model(&UserPO{}).Count(&count).Error; err == nil && count > 0 {
		return ensureWorkTaskSeed(db)
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
			{ID: 28, Name: "查看任务", Code: "task:list", Resource: "task", Action: "read"},
			{ID: 29, Name: "创建任务", Code: "task:create", Resource: "task", Action: "create"},
			{ID: 30, Name: "更新任务", Code: "task:update", Resource: "task", Action: "update"},
			{ID: 31, Name: "删除任务", Code: "task:delete", Resource: "task", Action: "delete"},
			{ID: 32, Name: "查看系统信息", Code: "system:list", Resource: "system", Action: "read"},
			{ID: 33, Name: "查看文档", Code: "doc:list", Resource: "doc", Action: "read"},
			{ID: 34, Name: "创建文档", Code: "doc:create", Resource: "doc", Action: "create"},
			{ID: 35, Name: "更新文档", Code: "doc:update", Resource: "doc", Action: "update"},
			{ID: 36, Name: "删除文档", Code: "doc:delete", Resource: "doc", Action: "delete"},
			{ID: 37, Name: "查看部门", Code: "department:list", Resource: "department", Action: "read"},
			{ID: 38, Name: "创建部门", Code: "department:create", Resource: "department", Action: "create"},
			{ID: 39, Name: "更新部门", Code: "department:update", Resource: "department", Action: "update"},
			{ID: 40, Name: "删除部门", Code: "department:delete", Resource: "department", Action: "delete"},
			{ID: 41, Name: "使用Agent", Code: "agent:list", Resource: "agent", Action: "read"},
			{ID: 42, Name: "查看业务任务", Code: "worktask:list", Resource: "worktask", Action: "read"},
			{ID: 43, Name: "创建业务任务", Code: "worktask:create", Resource: "worktask", Action: "create"},
			{ID: 44, Name: "更新业务任务", Code: "worktask:update", Resource: "worktask", Action: "update"},
			{ID: 45, Name: "删除业务任务", Code: "worktask:delete", Resource: "worktask", Action: "delete"},
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
			{1, 28}, {1, 29}, {1, 30}, {1, 31}, {1, 32},
			{1, 33}, {1, 34}, {1, 35}, {1, 36},
			{1, 37}, {1, 38}, {1, 39}, {1, 40},
			{1, 41},
			{3, 1}, {3, 2}, {3, 3}, {3, 5}, {3, 6}, {3, 7}, {3, 9}, {3, 13}, {3, 16},
			{3, 33}, {3, 34}, {3, 35},
			{3, 37}, {3, 38}, {3, 39},
			{1, 42}, {1, 43}, {1, 44}, {1, 45},
			{3, 42}, {3, 43}, {3, 44},
			{4, 42},
			// viewer（访客）仅仪表盘 + 用户/角色/菜单/审计只读 + 文档只读；不授部门管理
			{4, 1}, {4, 5}, {4, 9}, {4, 13}, {4, 16},
			{4, 33},
			// agent: 所有登录角色均可用（工具权限由 RBAC 兜底）
			{3, 41}, {4, 41},
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
			{ID: 20, ParentID: 2, Name: "任务管理", Path: "/system/tasks", Component: "pages/Task/TaskManage", Icon: "ScheduleOutlined", Sort: 9, Permission: "task:list", Version: "1.0.0"},
			{ID: 21, ParentID: 2, Name: "系统信息", Path: "/system/info", Component: "pages/System/SystemInfo", Icon: "MonitorOutlined", Sort: 10, Permission: "system:list", Version: "1.0.0"},
			{ID: 22, ParentID: 2, Name: "文档管理", Path: "/system/docs", Component: "pages/Doc/DocManage", Icon: "FileTextOutlined", Sort: 11, Permission: "doc:list", Version: "1.0.0"},
			{ID: 23, ParentID: 2, Name: "部门管理", Path: "/system/departments", Component: "pages/Department/DeptManage", Icon: "ApartmentOutlined", Sort: 12, Permission: "department:list", Version: "1.0.0"},
			{ID: 24, ParentID: 0, Name: "Agent 助手", Path: "/agent", Component: "pages/Agent/AgentChat", Icon: "MessageOutlined", Sort: 2, Permission: "agent:list", Version: "1.0.0"},
			{ID: 25, ParentID: 0, Name: "演示任务", Path: "/worktasks", Component: "pages/WorkTask/WorkTaskManage", Icon: "CheckSquareOutlined", Sort: 3, Permission: "worktask:list", Version: "1.0.0"},
		}
		if err := tx.Create(&menus).Error; err != nil {
			return fmt.Errorf("seed menus: %w", err)
		}

		// Role-Menu links (admin sees all, editor sees Dashboard+Users+Menus+文档+部门, viewer sees Dashboard+用户+文档 only)
		type RM struct{ RoleID, MenuID uint }
		rm := []RM{
			{1, 1}, {1, 2}, {1, 3}, {1, 7}, {1, 11}, {1, 15}, {1, 16}, {1, 17}, {1, 18}, {1, 19}, {1, 20}, {1, 21}, {1, 22}, {1, 23},
			{1, 24}, {1, 25},
			// editor：含系统管理目录（id 2），否则其子菜单（用户/菜单/文档/部门）在 buildTree 时无父级被丢弃
			{3, 1}, {3, 2}, {3, 3}, {3, 7}, {3, 22}, {3, 23},
			// viewer（访客）：系统管理目录（id2）+ 用户管理 + 文档管理；部门管理不授予
			{4, 1}, {4, 2}, {4, 3}, {4, 22},
			// agent 顶级菜单：所有登录角色均可见
			{3, 24}, {4, 24}, {3, 25},
		}
		if err := tx.Table("role_menus").Create(&rm).Error; err != nil {
			return fmt.Errorf("seed role_menus: %w", err)
		}

		// Config groups（站点 + 安全）
		groups := []ConfigGroupPO{
			{ID: 1, Name: "站点", Sort: 1},
			{ID: 2, Name: "安全", Sort: 2},
			{ID: 3, Name: "Agent", Sort: 3},
		}
		if err := tx.Create(&groups).Error; err != nil {
			return fmt.Errorf("seed config groups: %w", err)
		}

		// System Configs
		configs := []SystemConfigPO{
			{Key: "site_name", Value: "Kingfisher", Remark: "系统名称", IsPublic: true, Version: "1.0.0", Render: "text", GroupID: 1},
			{Key: "site_description", Value: "Kingfisher 后台管理平台", Remark: "系统描述", IsPublic: true, Version: "1.0.0", Render: "textarea", GroupID: 1},
			{Key: "site_logo", Value: "", Remark: "Logo（留空则显示站点名首字母）", IsPublic: true, Version: "1.0.0", Render: "text", GroupID: 1},
			{Key: "site_login_cover", Value: "", Remark: "登录页封面图（留空则显示纯色背景）", IsPublic: true, Version: "1.0.0", Render: "image", GroupID: 1},
			{Key: "site_notice", Value: "⚠️ 这是一个测试站点，请勿用于生产环境。管理员账号：admin / 密码：Abcd1234", Remark: "站点通知（显示在页面顶部，可关闭；内容变化后重新展示）", IsPublic: true, Version: "1.0.0", Render: "textarea", GroupID: 1},
			{Key: "max_login_attempts", Value: "5", Remark: "最大登录失败次数", Version: "1.0.0", Render: "number", GroupID: 2},
			{Key: "lockout_duration", Value: "15m", Remark: "锁定时间", Version: "1.0.0", Render: "text", GroupID: 2},
			{Key: "session_timeout", Value: "30m", Remark: "会话超时", Version: "1.0.0", Render: "text", GroupID: 2},
			{Key: "registration_enabled", Value: "true", Remark: "是否开放注册", IsPublic: true, Version: "1.0.0", Render: "switch", GroupID: 2},
			{Key: "password_reset_enabled", Value: "true", Remark: "是否开启找回密码", IsPublic: true, Version: "1.0.0", Render: "switch", GroupID: 2},
			{Key: "default_register_role_id", Value: "4", Remark: "默认注册用户的角色", Version: "1.0.0", Render: "select", RenderOptions: `[{"label":"访客","value":"4"},{"label":"编辑","value":"3"},{"label":"超级管理员","value":"1"}]`, GroupID: 2},
			{Key: "watermark_enabled", Value: "false", Remark: "全局水印开关（登录后所有页面显示水印）", IsPublic: true, Version: "1.0.0", Render: "switch", GroupID: 2},
			{Key: "watermark_text", Value: "Kingfisher 内部系统", Remark: "水印文字", IsPublic: true, Version: "1.0.0", Render: "text", GroupID: 2},
			{Key: "watermark_extra", Value: "{username} {date}", Remark: "水印补充内容（支持 {username}/{date} 占位符，留空则仅水印文字）", IsPublic: true, Version: "1.0.0", Render: "text", GroupID: 2},
			{Key: "agent_system_prompt", Value: "", Remark: "Agent 系统提示词（留空用默认；可覆盖以定制 agent 行为）", IsPublic: true, Version: "1.0.0", Render: "textarea", GroupID: 3},
			{Key: "agent_allowed_methods", Value: "GET,POST,PUT,PATCH,DELETE", Remark: "Agent call_api 允许的 HTTP 方法白名单（多选；重启后生效）", IsPublic: true, Version: "1.0.0", Render: "select", RenderOptions: `{"multiple":true,"options":[{"label":"GET","value":"GET"},{"label":"POST","value":"POST"},{"label":"PUT","value":"PUT"},{"label":"PATCH","value":"PATCH"},{"label":"DELETE","value":"DELETE"}]}`, GroupID: 3},
			{Key: "mfa_enforce", Value: "optional", Remark: "MFA 强制策略（optional=自选开启, all=全员强制, admin=仅管理员强制）", Version: "1.0.0", Render: "select", RenderOptions: `[{"label":"自选开启","value":"optional"},{"label":"全员强制","value":"all"},{"label":"仅管理员强制","value":"admin"}]`, GroupID: 2},
			{Key: "mfa_allow_totp", Value: "true", Remark: "允许 TOTP 二次验证", Version: "1.0.0", Render: "switch", GroupID: 2},
			{Key: "mfa_allow_sms", Value: "true", Remark: "允许短信二次验证", Version: "1.0.0", Render: "switch", GroupID: 2},
			{Key: "mfa_allow_email", Value: "true", Remark: "允许邮箱二次验证", Version: "1.0.0", Render: "switch", GroupID: 2},
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
			{ID: 1, Username: "admin", Nickname: "admin", Password: pwHash, Email: "admin@example.com", Status: 1},
			{ID: 2, Username: "editor", Nickname: "editor", Password: pwHash, Email: "editor@example.com", Status: 1},
			{ID: 3, Username: "viewer", Nickname: "viewer", Password: pwHash, Email: "viewer@example.com", Status: 1},
			{ID: 4, Username: "multi", Nickname: "multi", Password: pwHash, Email: "multi@example.com", Status: 1},
		}
		if err := tx.Create(&users).Error; err != nil {
			return fmt.Errorf("seed users: %w", err)
		}
		// User-Role 关联（多角色支持）：multi 同时拥有 管理员+编辑+访客
		userRoles := []UserRolePO{
			{UserID: 1, RoleID: 1}, // admin → 超级管理员
			{UserID: 2, RoleID: 3}, // editor → 编辑
			{UserID: 3, RoleID: 4}, // viewer → 访客
			{UserID: 4, RoleID: 1}, // multi 多角色示例：管理员
			{UserID: 4, RoleID: 3}, // multi 多角色示例：编辑
			{UserID: 4, RoleID: 4}, // multi 多角色示例：访客
		}
		if err := tx.Create(&userRoles).Error; err != nil {
			return fmt.Errorf("seed user_roles: %w", err)
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
			{ID: 2, Name: "找回密码通知", Code: "password_reset", TemplateType: "email", Title: "找回密码 - Kingfisher", Content: `<p>你好 {{nickname}}：</p><p>我们收到了你的密码重置请求，请点击以下链接在 30 分钟内重置密码：</p><p><a href="{{reset_url}}">{{reset_url}}</a></p><p>如果这不是你的操作，请忽略此邮件。</p>`, Status: 1, Version: "1.0.0"},
		}
		if err := tx.Create(&templates).Error; err != nil {
			return fmt.Errorf("seed templates: %w", err)
		}

		// Scheduled Tasks（示例：nop 空转任务，用于测试周期调度链路）
		scheduledTasks := []ScheduledTaskPO{
			{ID: 1, Name: "nop 测试任务", TaskType: "nop:run", CronSpec: "* * * * *", Payload: `{"note":"周期任务测试"}`, Enabled: 1, Remark: "每 分钟空转一次，验证调度器→入队→worker 消费链路"},
		}
		if err := tx.Create(&scheduledTasks).Error; err != nil {
			return fmt.Errorf("seed scheduled_tasks: %w", err)
		}

		// Doc 模块示例数据：目录（含可见角色授权，演示默认拒绝）+ 示例文档
		docDirs := []DocDirectoryPO{
			{ID: 1, ParentID: 0, Name: "产品文档", Sort: 1, Status: 1, Version: "1.0.0"},
			{ID: 2, ParentID: 0, Name: "技术文档", Sort: 2, Status: 1, Version: "1.0.0"},
			{ID: 3, ParentID: 1, Name: "内部资料", Sort: 1, Status: 1, Version: "1.0.0"},
		}
		if err := tx.Create(&docDirs).Error; err != nil {
			return fmt.Errorf("seed doc dirs: %w", err)
		}
		// 目录可见角色：dir1 全角色可见；dir2 admin+editor；dir3 仅 admin（演示默认拒绝）
		type DR struct{ DirID, RoleID uint }
		dr := []DR{
			{1, 1}, {1, 3}, {1, 4},
			{2, 1}, {2, 3},
			{3, 1},
		}
		if err := tx.Table("doc_dir_roles").Create(&dr).Error; err != nil {
			return fmt.Errorf("seed doc_dir_roles: %w", err)
		}
		// 示例文档：doc1 已发布共享（admin 作者）；doc2 草稿（演示 draft 仅作者可见）
		docs := []DocumentPO{
			{ID: 1, DirID: 1, Title: "产品使用手册", Content: "<h2>欢迎使用 Kingfisher</h2><p>本文档介绍产品核心功能。</p>", OwnerID: 1, Visibility: "shared", Status: "published", CurrentVersion: 1, Sort: 0},
			{ID: 2, DirID: 2, Title: "开发规范", Content: "<p>代码风格与提交流程。</p>", OwnerID: 2, Visibility: "shared", Status: "draft", CurrentVersion: 1, Sort: 0},
		}
		if err := tx.Create(&docs).Error; err != nil {
			return fmt.Errorf("seed docs: %w", err)
		}
		docVersions := []DocVersionPO{
			{DocID: 1, VersionNo: 1, Title: "产品使用手册", Content: "<h2>欢迎使用 Kingfisher</h2><p>本文档介绍产品核心功能。</p>", OwnerID: 1, Note: "初始版本"},
			{DocID: 2, VersionNo: 1, Title: "开发规范", Content: "<p>代码风格与提交流程。</p>", OwnerID: 2, Note: "初始版本"},
		}
		if err := tx.Create(&docVersions).Error; err != nil {
			return fmt.Errorf("seed doc versions: %w", err)
		}

		// Department 模块示例数据：部门树 + 部门-角色关联 + 用户-部门关联
		departments := []DepartmentPO{
			{ID: 1, ParentID: 0, Name: "技术部", Sort: 1, Status: 1, Remark: "研发与技术支持", Version: "1.0.0"},
			{ID: 2, ParentID: 0, Name: "产品部", Sort: 2, Status: 1, Remark: "产品规划与设计", Version: "1.0.0"},
			{ID: 3, ParentID: 1, Name: "后端组", Sort: 1, Status: 1, Remark: "服务端开发", Version: "1.0.0"},
		}
		if err := tx.Create(&departments).Error; err != nil {
			return fmt.Errorf("seed departments: %w", err)
		}
		// 部门-角色关联：技术部挂「编辑」角色，产品部挂「访客」角色（演示部门角色继承）
		deptRoles := []DepartmentRolePO{
			{DepartmentID: 1, RoleID: 3},
			{DepartmentID: 2, RoleID: 4},
		}
		if err := tx.Create(&deptRoles).Error; err != nil {
			return fmt.Errorf("seed department_roles: %w", err)
		}
		// 用户-部门关联：admin→技术部+产品部；editor→技术部；viewer→产品部；multi→技术部+产品部
		userDepts := []UserDepartmentPO{
			{UserID: 1, DepartmentID: 1},
			{UserID: 1, DepartmentID: 2},
			{UserID: 2, DepartmentID: 1},
			{UserID: 3, DepartmentID: 2},
			{UserID: 4, DepartmentID: 1},
			{UserID: 4, DepartmentID: 2},
		}
		return tx.Create(&userDepts).Error
	})
}

func ensureWorkTaskSeed(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		perms := []PermissionPO{
			{ID: 42, Name: "查看业务任务", Code: "worktask:list", Resource: "worktask", Action: "read"},
			{ID: 43, Name: "创建业务任务", Code: "worktask:create", Resource: "worktask", Action: "create"},
			{ID: 44, Name: "更新业务任务", Code: "worktask:update", Resource: "worktask", Action: "update"},
			{ID: 45, Name: "删除业务任务", Code: "worktask:delete", Resource: "worktask", Action: "delete"},
		}
		for _, perm := range perms {
			if err := tx.Where("code = ?", perm.Code).FirstOrCreate(&perm).Error; err != nil {
				return fmt.Errorf("ensure worktask permission %s: %w", perm.Code, err)
			}
		}
		// 演示任务改为根菜单：ParentID=0, Path=/worktasks，兼容旧数据迁移
		var existing MenuPO
		if err := tx.Where("id = ?", 25).First(&existing).Error; err == nil {
			if existing.ParentID != 0 || existing.Path != "/worktasks" || existing.Sort != 3 {
				if err := tx.Model(&MenuPO{}).Where("id = ?", 25).Updates(map[string]interface{}{"parent_id": 0, "path": "/worktasks", "sort": 3}).Error; err != nil {
					return fmt.Errorf("migrate worktask menu: %w", err)
				}
			}
		} else {
			// 兼容旧 path 存量数据
			if err := tx.Where("path = ?", "/system/worktasks").First(&existing).Error; err == nil {
				if err := tx.Model(&MenuPO{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{"parent_id": 0, "path": "/worktasks", "sort": 3}).Error; err != nil {
					return fmt.Errorf("migrate worktask menu by path: %w", err)
				}
			}
		}
		menu := MenuPO{ID: 25, ParentID: 0, Name: "演示任务", Path: "/worktasks", Component: "pages/WorkTask/WorkTaskManage", Icon: "CheckSquareOutlined", Sort: 3, Permission: "worktask:list", Status: 1, Version: "1.0.0"}
		if err := tx.Where("id = ?", menu.ID).FirstOrCreate(&menu).Error; err != nil {
			return fmt.Errorf("ensure worktask menu: %w", err)
		}
		if menu.ParentID != 0 || menu.Path != "/worktasks" {
			if err := tx.Model(&MenuPO{}).Where("id = ?", menu.ID).Updates(map[string]interface{}{"parent_id": 0, "path": "/worktasks", "sort": 3}).Error; err != nil {
				return fmt.Errorf("ensure worktask menu fix: %w", err)
			}
		}
		rolePerms := map[uint][]string{1: {"worktask:list", "worktask:create", "worktask:update", "worktask:delete"}, 3: {"worktask:list", "worktask:create", "worktask:update"}, 4: {"worktask:list"}}
		for roleID, codes := range rolePerms {
			for _, code := range codes {
				var perm PermissionPO
				if err := tx.Where("code = ?", code).First(&perm).Error; err != nil {
					return err
				}
				if err := tx.FirstOrCreate(&RolePermissionPO{RoleID: roleID, PermissionID: perm.ID}).Error; err != nil {
					return fmt.Errorf("ensure worktask role permission: %w", err)
				}
			}
		}
		for _, roleID := range []uint{1, 3, 4} {
			if err := tx.FirstOrCreate(&RoleMenuPO{RoleID: roleID, MenuID: menu.ID}).Error; err != nil {
				return fmt.Errorf("ensure worktask role menu: %w", err)
			}
		}
		mfaConfigs := []SystemConfigPO{
			{Key: "mfa_enforce", Value: "optional", Remark: "MFA 强制策略（optional=自选开启, all=全员强制, admin=仅管理员强制）", Version: "1.0.0", Render: "select", RenderOptions: `[{"label":"自选开启","value":"optional"},{"label":"全员强制","value":"all"},{"label":"仅管理员强制","value":"admin"}]`, GroupID: 2},
			{Key: "mfa_allow_totp", Value: "true", Remark: "允许 TOTP 二次验证", Version: "1.0.0", Render: "switch", GroupID: 2},
			{Key: "mfa_allow_sms", Value: "true", Remark: "允许短信二次验证", Version: "1.0.0", Render: "switch", GroupID: 2},
			{Key: "mfa_allow_email", Value: "true", Remark: "允许邮箱二次验证", Version: "1.0.0", Render: "switch", GroupID: 2},
		}
		for _, cfg := range mfaConfigs {
			if err := tx.Where("key = ?", cfg.Key).FirstOrCreate(&cfg).Error; err != nil {
				return fmt.Errorf("ensure mfa config %s: %w", cfg.Key, err)
			}
		}
		return nil
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
