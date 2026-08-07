package adapter

import "time"

type rolePO struct {
	ID          uint
	Name        string
	Code        string
	Description string
	Status      int
	Level       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (rolePO) TableName() string { return "roles" }

type permissionPO struct {
	ID        uint
	Name      string
	Code      string
	Resource  string
	Action    string
	CreatedAt time.Time
}

func (permissionPO) TableName() string { return "permissions" }

type rolePermissionPO struct {
	RoleID       uint
	PermissionID uint
}

func (rolePermissionPO) TableName() string { return "role_permissions" }

type roleMenuPO struct {
	RoleID uint
	MenuID uint
}

func (roleMenuPO) TableName() string { return "role_menus" }

type menuPO struct {
	ID         uint
	ParentID   uint
	Name       string
	Path       string
	Component  string
	Icon       string
	Sort       int
	Type       int
	Permission string
	Status     int
}

func (menuPO) TableName() string { return "menus" }
