// Package database implements database logic.

package database

import (
	"time"

	"gorm.io/gorm"
)

// Persistent Objects (PO) — used by AutoMigrate + Seed in SQLite dev mode.
// Production MySQL/PG uses migrations/*.sql instead.

type UserPO struct {
	ID             uint   `gorm:"primaryKey"`
	Username       string `gorm:"size:32;uniqueIndex;not null"`
	Nickname       string `gorm:"size:64"`
	Password       string `gorm:"size:128;not null"`
	Email          string `gorm:"size:128"`
	Avatar         string `gorm:"size:255"`
	Status         int    `gorm:"default:1"`
	SessionVersion int    `gorm:"default:1"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (UserPO) TableName() string { return "users" }

type RolePO struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:32;not null"`
	Code        string `gorm:"size:32;uniqueIndex;not null"`
	Description string `gorm:"size:255"`
	Status      int    `gorm:"default:1"`
	Level       int    `gorm:"default:2"`
	// LandingPage 角色登录后的落地页（如 /dashboard）
	LandingPage string `gorm:"size:128"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (RolePO) TableName() string { return "roles" }

type PermissionPO struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:64;not null"`
	Code      string `gorm:"size:64;uniqueIndex;not null"`
	Resource  string `gorm:"size:32;not null"`
	Action    string `gorm:"size:16;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (PermissionPO) TableName() string { return "permissions" }

type MenuPO struct {
	ID         uint   `gorm:"primaryKey"`
	ParentID   uint   `gorm:"default:0;index"`
	Name       string `gorm:"size:32;not null"`
	Path       string `gorm:"size:128"`
	Component  string `gorm:"size:128"`
	Icon       string `gorm:"size:64"`
	Sort       int    `gorm:"default:0"`
	Type       int    `gorm:"default:2"`
	Permission string `gorm:"size:64"`
	Status     int    `gorm:"default:1"`
	// Version 表示该菜单由哪个版本新增
	Version   string `gorm:"size:32"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (MenuPO) TableName() string { return "menus" }

type SystemConfigPO struct {
	ID     uint   `gorm:"primaryKey"`
	Key    string `gorm:"size:64;uniqueIndex;not null"`
	Value  string `gorm:"type:text;not null"`
	Remark string `gorm:"size:255"`
	// IsPublic 是否公开：公开项可在未登录状态下通过 /api/v1/public/configs 读取
	IsPublic bool `gorm:"default:false;not null"`
	// Version 表示该配置由哪个版本新增
	Version string `gorm:"size:32"`
	// Render 前端渲染组件：text|number|switch|select|textarea
	Render string `gorm:"size:32"`
	// RenderOptions 渲染组件配置（JSON），如 select 的选项 [{"label":"开启","value":"1"}]
	RenderOptions string `gorm:"type:text"`
	// GroupID 配置分组（关联 config_groups.id）
	GroupID   uint `gorm:"default:0;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SystemConfigPO) TableName() string { return "system_configs" }

type ConfigGroupPO struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:64;uniqueIndex;not null"`
	Sort      int    `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ConfigGroupPO) TableName() string { return "config_groups" }

type AuditLogPO struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"index;not null"`
	Username   string    `gorm:"size:32;not null"`
	Action     string    `gorm:"size:16;not null;index"`
	Resource   string    `gorm:"size:32;not null;index:idx_resource"`
	ResourceID uint      `gorm:"default:0;index:idx_resource"`
	Detail     string    `gorm:"type:text"`
	Result     string    `gorm:"size:16;not null;default:success;index"`
	Latency    int64     `gorm:"default:0"`
	Message    string    `gorm:"size:255"`
	IP         string    `gorm:"size:45"`
	UserAgent  string    `gorm:"size:512"`
	CreatedAt  time.Time `gorm:"index"`
}

func (AuditLogPO) TableName() string { return "audit_logs" }

// Junction tables need explicit models for AutoMigrate
type RoleMenuPO struct {
	RoleID uint `gorm:"primaryKey"`
	MenuID uint `gorm:"primaryKey"`
}

func (RoleMenuPO) TableName() string { return "role_menus" }

type RolePermissionPO struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

func (RolePermissionPO) TableName() string { return "role_permissions" }

// UserRolePO 用户-角色关联（多对多，多角色支持）
type UserRolePO struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

func (UserRolePO) TableName() string { return "user_roles" }

type DictTypePO struct {
	ID       uint   `gorm:"primaryKey"`
	Code     string `gorm:"size:64;uniqueIndex;not null"`
	Name     string `gorm:"size:128;not null"`
	IsPublic bool   `gorm:"default:false;not null"`
	Status   int    `gorm:"default:1"`
	Remark   string `gorm:"size:255"`
	// Version 表示该字典类型由哪个版本新增
	Version   string `gorm:"size:32"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (DictTypePO) TableName() string { return "dict_types" }

type DictEntryPO struct {
	ID     uint   `gorm:"primaryKey"`
	TypeID uint   `gorm:"index;not null"`
	Label  string `gorm:"size:128;not null"`
	Value  string `gorm:"size:128;not null"`
	Sort   int    `gorm:"default:0"`
	Status int    `gorm:"default:1"`
	Remark string `gorm:"size:255"`
	// Version 表示该字典条目由哪个版本新增
	Version   string `gorm:"size:32"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (DictEntryPO) TableName() string { return "dict_entries" }

// MessagePO 站内信消息
type MessagePO struct {
	ID          uint   `gorm:"primaryKey"`
	SenderID    uint   `gorm:"index"`
	SenderType  string `gorm:"size:16;default:admin"` // admin | system
	RecipientID uint   `gorm:"index;not null"`
	Title       string `gorm:"size:128;not null"`
	Content     string `gorm:"type:text"`
	Status      string `gorm:"size:16;default:sent"` // draft | sent
	IsRead      bool   `gorm:"default:false;index"`
	ReadAt      *time.Time
	CreatedAt   time.Time `gorm:"index"`
	UpdatedAt   time.Time
}

func (MessagePO) TableName() string { return "messages" }

// TemplatePO 模版（消息/通知/通用，通过 TemplateType 区分）
type TemplatePO struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"size:64;not null"`
	Code         string `gorm:"size:64;uniqueIndex;not null"`
	TemplateType string `gorm:"size:32;default:general;index"`
	Title        string `gorm:"size:255;not null"`
	Content      string `gorm:"type:text"`
	Status       int    `gorm:"default:1"`
	Remark       string `gorm:"size:255"`
	// Version 表示该模版由哪个版本新增
	Version   string `gorm:"size:32"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (TemplatePO) TableName() string { return "templates" }

// ScheduledTaskPO 周期任务配置（后台管理 → asynq PeriodicTaskManager 周期性同步调度）
type ScheduledTaskPO struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:64;not null"`
	TaskType  string `gorm:"size:64;not null"`
	CronSpec  string `gorm:"size:64;not null"`
	Payload   string `gorm:"type:text"`
	Enabled   int    `gorm:"default:1;index"`
	Remark    string `gorm:"size:255"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ScheduledTaskPO) TableName() string { return "scheduled_tasks" }

// DocDirectoryPO 文档目录（树形）。可见性 shared/private（shared 公开，private 仅 admin）
type DocDirectoryPO struct {
	ID         uint   `gorm:"primaryKey"`
	ParentID   uint   `gorm:"default:0;index"`
	Name       string `gorm:"size:64;not null"`
	Sort       int    `gorm:"default:0"`
	Status     int    `gorm:"default:1"`
	Version    string `gorm:"size:32"`
	Visibility string `gorm:"size:16;default:shared"` // shared | private
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (DocDirectoryPO) TableName() string { return "doc_directories" }

// DocDirRolePO 目录-角色授权（历史遗留表，已弃用：目录可见性改为 shared/private）
type DocDirRolePO struct {
	DirID  uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

func (DocDirRolePO) TableName() string { return "doc_dir_roles" }

// DocumentPO 文档
type DocumentPO struct {
	ID             uint   `gorm:"primaryKey"`
	DirID          uint   `gorm:"index;not null"`
	Title          string `gorm:"size:128;not null"`
	Content        string `gorm:"type:text"`
	OwnerID        uint   `gorm:"index;not null"`
	Visibility     string `gorm:"size:16;default:shared"`
	Status         string `gorm:"size:16;default:draft;index"`
	CurrentVersion int    `gorm:"default:1"`
	Sort           int    `gorm:"default:0"`
	PublishedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (DocumentPO) TableName() string { return "documents" }

// DocVersionPO 文档版本历史（append-only；UNIQUE(doc_id, version_no)）
type DocVersionPO struct {
	ID        uint   `gorm:"primaryKey"`
	DocID     uint   `gorm:"index;not null"`
	VersionNo int    `gorm:"not null"`
	Title     string `gorm:"size:128;not null"`
	Content   string `gorm:"type:text"`
	OwnerID   uint   `gorm:"default:0"`
	Note      string `gorm:"size:255"`
	CreatedAt time.Time
}

func (DocVersionPO) TableName() string { return "doc_versions" }

// DepartmentPO 部门（树形）。一个部门可挂多个角色（department_roles），一个用户可属于多个部门（user_departments）
type DepartmentPO struct {
	ID        uint   `gorm:"primaryKey"`
	ParentID  uint   `gorm:"default:0;index"` // 0=根
	Name      string `gorm:"size:64;not null"`
	Sort      int    `gorm:"default:0"`
	Status    int    `gorm:"default:1"` // 1=启用 0=停用
	Remark    string `gorm:"size:255"`
	Version   string `gorm:"size:32"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (DepartmentPO) TableName() string { return "departments" }

// UserDepartmentPO 用户-部门关联（多对多，一个用户可属于多个部门）
type UserDepartmentPO struct {
	UserID       uint `gorm:"primaryKey"`
	DepartmentID uint `gorm:"primaryKey"`
}

func (UserDepartmentPO) TableName() string { return "user_departments" }

// DepartmentRolePO 部门-角色关联（多对多，一个部门可有多个角色）
type DepartmentRolePO struct {
	DepartmentID uint `gorm:"primaryKey"`
	RoleID       uint `gorm:"primaryKey"`
}

func (DepartmentRolePO) TableName() string { return "department_roles" }
