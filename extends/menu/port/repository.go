package port

import (
	"context"

	"kingfisher/extends/menu/domain"
)

type MenuRepository interface {
	FindAll(ctx context.Context) ([]domain.Menu, error)
	FindByID(ctx context.Context, id uint) (*domain.Menu, error)
	FindByParentID(ctx context.Context, parentID uint) ([]domain.Menu, error)
	FindByRoleIDs(ctx context.Context, roleIDs []uint) ([]domain.Menu, error)
	Create(ctx context.Context, menu *domain.Menu) error
	Update(ctx context.Context, id uint, updates map[string]any) error
	Delete(ctx context.Context, id uint) error
	DeleteBatch(ctx context.Context, ids []uint) error
	UpdateStatusBatch(ctx context.Context, ids []uint, status int) error
	HasChildren(ctx context.Context, parentID uint) (bool, error)
}
