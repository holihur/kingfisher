// Package port defines the department module's repository interface.
package port

import (
	"context"

	"kingfisher/core/query"
	"kingfisher/extends/department/domain"
)

// DepartmentRepository 部门仓储接口。
// 用户-部门成员关系由 user 模块维护（user_departments），本模块只需部门树 CRUD 与部门角色分配。
type DepartmentRepository interface {
	// FindAll 返回全部部门（按 sort/id 排序，service 层 buildTree 组树）
	FindAll(ctx context.Context) ([]domain.Department, error)
	// ListPage 分页部门列表（query DSL；subtree_id 筛选某部门及其子孙）
	ListPage(ctx context.Context, q *query.Query) ([]domain.Department, int64, error)
	// GetByID 返回单个部门（含 RoleIDs/Roles）
	GetByID(ctx context.Context, id uint) (*domain.Department, error)
	Create(ctx context.Context, d *domain.Department) error
	Update(ctx context.Context, id uint, updates map[string]any) error
	// Delete 删除部门（同一事务内级联清理 user_departments / department_roles）
	Delete(ctx context.Context, id uint) error
	HasChildren(ctx context.Context, parentID uint) (bool, error)
	// SubtreeIDs 返回某部门及其全部子孙的 ID 集合
	SubtreeIDs(ctx context.Context, rootID uint) ([]uint, error)
	// SetRoles 先删后插替换部门的角色关联
	SetRoles(ctx context.Context, departmentID uint, roleIDs []uint) error
}
