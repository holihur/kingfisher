// Package domain implements department domain entities.
package domain

import "time"

// Department 部门（树形）。一个部门可挂多个角色（Roles），一个用户可属于多个部门。
// 部门角色会合并进成员用户的有效角色（user_roles ∪ department_roles ⋈ user_departments）。
type Department struct {
	ID        uint         `json:"id"`
	ParentID  uint         `json:"parent_id"` // 0=根
	Name      string       `json:"name"`
	Sort      int          `json:"sort"`
	Status    int          `json:"status"` // 1=启用 0=停用
	Remark    string       `json:"remark"`
	RoleIDs   []uint       `json:"role_ids,omitempty"` // 部门直接挂载的角色 ID
	Roles     []*Role      `json:"roles,omitempty"`    // 部门挂载的角色详情
	Children  []Department `json:"children,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// Role 部门挂载角色的轻量视图（与 user 模块的 domain.Role 同构）
type Role struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}
