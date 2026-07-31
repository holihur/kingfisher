package port

import (
	"context"

	"kingfisher/extends/rbac/domain"
)

type RoleRepository interface {
	FindAll(ctx context.Context) ([]domain.Role, error)
	FindByID(ctx context.Context, id uint) (*domain.Role, error)
	Create(ctx context.Context, role *domain.Role) error
	Update(ctx context.Context, id uint, updates map[string]any) error
	Delete(ctx context.Context, id uint) error
	AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error
	AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error
	GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
	GetRolePermissions(ctx context.Context, roleID uint) ([]domain.Permission, error)
	GetRoleMenus(ctx context.Context, roleID uint) ([]domain.Menu, error)
}
type PermissionRepository interface {
	FindAll(ctx context.Context) ([]domain.Permission, error)
}
