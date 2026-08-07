package adapter

import (
	"time"

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
	SessionVersion int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (userPO) TableName() string { return "users" }

func (p userPO) toDomain() *domain.User {
	return &domain.User{
		ID: p.ID, Username: p.Username, Nickname: p.Nickname, Password: p.Password,
		Email: p.Email, Avatar: p.Avatar, Status: p.Status,
		RoleID: p.RoleID, SessionVersion: p.SessionVersion,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}
