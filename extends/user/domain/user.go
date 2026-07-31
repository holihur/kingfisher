package domain

import "time"

type User struct {
	ID             uint      `json:"id"`
	Username       string    `json:"username"`
	Password       string    `json:"-"` // bcrypt hash
	Email          string    `json:"email"`
	Avatar         string    `json:"avatar"`
	Status         int       `json:"status"` // 1=启用 0=禁用
	RoleID         uint      `json:"role_id"`
	Role           *Role     `json:"role,omitempty"`
	SessionVersion int       `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Role struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}
