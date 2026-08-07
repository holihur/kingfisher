package port

import (
	"context"

	"kingfisher/core/query"
	"kingfisher/extends/rbac/domain"
)

type RoleRepository interface {
	FindAll(ctx context.Context, q *query.Query) ([]domain.Role, int64, error)
	FindByID(ctx context.Context, id uint) (*domain.Role, error)
	FindByCode(ctx context.Context, code string) (*domain.Role, error)
	Create(ctx context.Context, role *domain.Role) error
	Update(ctx context.Context, id uint, updates map[string]any) error
	Delete(ctx context.Context, id uint) error
	DeleteBatch(ctx context.Context, ids []uint) error
	UpdateStatusBatch(ctx context.Context, ids []uint, status int) error
	AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error
	AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error
	GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
	GetRolePermissions(ctx context.Context, roleID uint) ([]domain.Permission, error)
	GetRoleMenus(ctx context.Context, roleID uint) ([]domain.Menu, error)
}
type PermissionRepository interface {
	FindAll(ctx context.Context) ([]domain.Permission, error)
}
