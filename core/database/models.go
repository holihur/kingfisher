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
	Password       string `gorm:"size:128;not null"`
	Email          string `gorm:"size:128"`
	Avatar         string `gorm:"size:255"`
	Status         int    `gorm:"default:1"`
	RoleID         uint   `gorm:"default:0"`
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
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (MenuPO) TableName() string { return "menus" }

type SystemConfigPO struct {
	ID        uint   `gorm:"primaryKey"`
	Key       string `gorm:"size:64;uniqueIndex;not null"`
	Value     string `gorm:"type:text;not null"`
	Remark    string `gorm:"size:255"`
	// IsPublic 是否公开：公开项可在未登录状态下通过 /api/v1/public/configs 读取
	IsPublic bool   `gorm:"default:false;not null"`
	// Version 表示该配置由哪个版本新增
	Version string `gorm:"size:32"`
	// Render 前端渲染组件：text|number|switch|select|textarea
	Render string `gorm:"size:32"`
	// RenderOptions 渲染组件配置（JSON），如 select 的选项 [{"label":"开启","value":"1"}]
	RenderOptions string `gorm:"type:text"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (SystemConfigPO) TableName() string { return "system_configs" }

type AuditLogPO struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"index;not null"`
	Username   string    `gorm:"size:32;not null"`
	Action     string    `gorm:"size:16;not null"`
	Resource   string    `gorm:"size:32;not null;index:idx_resource"`
	ResourceID uint      `gorm:"default:0;index:idx_resource"`
	Detail     string    `gorm:"type:text"`
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
