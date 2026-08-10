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
	Roles          []rolePO `gorm:"many2many:user_roles;joinForeignKey:UserID;joinReferences:RoleID"`
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
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
	for _, r := range p.Roles {
		u.RoleIDs = append(u.RoleIDs, r.ID)
		u.Roles = append(u.Roles, &domain.Role{ID: r.ID, Name: r.Name, Code: r.Code})
	}
	return u
}

// userRolePO 用户-角色关联（与 core/database 的 UserRolePO 同构，供本模块读写 user_roles 表）
type userRolePO struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

func (userRolePO) TableName() string { return "user_roles" }
