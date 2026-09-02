package domain

import "time"

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Password string `json:"-"` // bcrypt hash
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Status   int    `json:"status"` // 1=启用 0=禁用
	// ParentID 子账户的父账户 ID，nil 表示主账户（一级父子，子不能再建子）
	ParentID *uint `json:"parent_id,omitempty"`
	// RoleIDs/Roles 为「有效角色」= 直接分配 ∪ 部门继承（查询时合并，去重）
	RoleIDs []uint  `json:"role_ids"`
	Roles   []*Role `json:"roles,omitempty"`
	// DirectRoleIDs 直接分配的角色（user_roles），供编辑表单回填；部门继承角色不在其中
	DirectRoleIDs  []uint        `json:"direct_role_ids,omitempty"`
	DeptIDs        []uint        `json:"dept_ids,omitempty"`
	Departments    []*Department `json:"departments,omitempty"`
	SessionVersion int           `json:"-"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
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
