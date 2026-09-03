package domain

import "time"

type User struct {
	ID              uint          `json:"id"`
	Username        string        `json:"username"`
	Nickname        string        `json:"nickname"`
	Password        string        `json:"-"` // bcrypt hash
	Email           string        `json:"email"`
	Phone           string        `json:"phone"`
	Avatar          string        `json:"avatar"`
	Status          int           `json:"status"`
	ParentID        *uint         `json:"parent_id,omitempty"`
	RoleIDs         []uint        `json:"role_ids"`
	Roles           []*Role       `json:"roles,omitempty"`
	DirectRoleIDs   []uint        `json:"direct_role_ids,omitempty"`
	DeptIDs         []uint        `json:"dept_ids,omitempty"`
	Departments     []*Department `json:"departments,omitempty"`
	SessionVersion  int           `json:"-"`
	MFATOTPEnabled  bool          `json:"mfa_totp_enabled"`
	MFASMSEnabled   bool          `json:"mfa_sms_enabled"`
	MFAEmailEnabled bool          `json:"mfa_email_enabled"`
	MFAEnabled      bool          `json:"mfa_enabled"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

type Role struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// Department 用户所属部门的轻量视图
type Department struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
