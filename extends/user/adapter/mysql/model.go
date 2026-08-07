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
	RoleID         uint
	Role           rolePO        `gorm:"foreignKey:RoleID;references:ID"`
	SessionVersion int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (userPO) TableName() string { return "users" }

type rolePO struct {
	ID   uint   `gorm:"primaryKey"`
	Name string
	Code string
}

func (rolePO) TableName() string { return "roles" }

func (p userPO) toDomain() *domain.User {
	u := &domain.User{
		ID: p.ID, Username: p.Username, Nickname: p.Nickname, Password: p.Password,
		Email: p.Email, Avatar: p.Avatar, Status: p.Status,
		RoleID: p.RoleID, SessionVersion: p.SessionVersion,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
	if p.Role.ID != 0 {
		u.Role = &domain.Role{ID: p.Role.ID, Name: p.Role.Name, Code: p.Role.Code}
	}
	return u
}
