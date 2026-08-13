// Package adapter implements the department module's GORM persistence layer.
package adapter

import (
	"time"

	"kingfisher/extends/department/domain"
)

// departmentPO 部门表（GORM PO；SQLite AutoMigrate 使用，生产走 migrations SQL）
type departmentPO struct {
	ID        uint `gorm:"primaryKey"`
	ParentID  uint `gorm:"default:0;index"`
	Name      string
	Sort      int
	Status    int
	Remark    string
	Version   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (departmentPO) TableName() string { return "departments" }

// departmentRolePO 部门-角色关联（复合主键）
type departmentRolePO struct {
	DepartmentID uint `gorm:"primaryKey"`
	RoleID       uint `gorm:"primaryKey"`
}

func (departmentRolePO) TableName() string { return "department_roles" }

// userDepartmentPO 用户-部门关联（复合主键；本模块只读，成员维护在 user 模块）
type userDepartmentPO struct {
	UserID       uint `gorm:"primaryKey"`
	DepartmentID uint `gorm:"primaryKey"`
}

func (userDepartmentPO) TableName() string { return "user_departments" }

// rolePO 角色表（部门挂载角色详情读取用；与 rbac 模块同表）
type rolePO struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Code string
}

func (rolePO) TableName() string { return "roles" }

func (p departmentPO) toDepartment() *domain.Department {
	return &domain.Department{
		ID: p.ID, ParentID: p.ParentID, Name: p.Name, Sort: p.Sort,
		Status: p.Status, Remark: p.Remark,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}
