package adapter

import (
	"time"

	"gorm.io/gorm"

	"kingfisher/extends/user/domain"
)

type userPO struct {
	ID             uint `gorm:"primaryKey"`
	Username       string
	Nickname       string
	Password       string
	Email          string
	Avatar         string
	Status         int
	Roles          []rolePO       `gorm:"many2many:user_roles;joinForeignKey:UserID;joinReferences:RoleID"`
	Departments    []departmentPO `gorm:"many2many:user_departments;joinForeignKey:UserID;joinReferences:DepartmentID"`
	SessionVersion int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (userPO) TableName() string { return "users" }

type rolePO struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Code string
}

func (rolePO) TableName() string { return "roles" }

func (p userPO) toDomain() *domain.User {
	u := &domain.User{
		ID: p.ID, Username: p.Username, Nickname: p.Nickname, Password: p.Password,
		Email: p.Email, Avatar: p.Avatar, Status: p.Status,
		SessionVersion: p.SessionVersion,
		CreatedAt:      p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
	for _, r := range p.Roles {
		u.RoleIDs = append(u.RoleIDs, r.ID)
		u.Roles = append(u.Roles, &domain.Role{ID: r.ID, Name: r.Name, Code: r.Code})
		u.DirectRoleIDs = append(u.DirectRoleIDs, r.ID)
	}
	for _, d := range p.Departments {
		u.DeptIDs = append(u.DeptIDs, d.ID)
		u.Departments = append(u.Departments, &domain.Department{ID: d.ID, Name: d.Name})
	}
	return u
}

// userRolePO 用户-角色关联（与 core/database 的 UserRolePO 同构，供本模块读写 user_roles 表）
type userRolePO struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

func (userRolePO) TableName() string { return "user_roles" }

// departmentPO 部门表（用户所属部门的轻量读取）
type departmentPO struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func (departmentPO) TableName() string { return "departments" }

// userDepartmentPO 用户-部门关联（复合主键；供本模块读写 user_departments 表）
type userDepartmentPO struct {
	UserID       uint `gorm:"primaryKey"`
	DepartmentID uint `gorm:"primaryKey"`
}

func (userDepartmentPO) TableName() string { return "user_departments" }

// departmentRolePO 部门-角色关联（读成员用户的部门继承角色用）
type departmentRolePO struct {
	DepartmentID uint `gorm:"primaryKey"`
	RoleID       uint `gorm:"primaryKey"`
}

func (departmentRolePO) TableName() string { return "department_roles" }
