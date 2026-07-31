package port

import (
	"context"

	"kingfisher/extends/user/domain"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uint) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindAll(ctx context.Context, page, pageSize int, keyword string) ([]domain.User, int64, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, id uint, updates map[string]any) error
	Delete(ctx context.Context, id uint) error
	IncrementSessionVersion(ctx context.Context, id uint) error
	GetSessionVersion(ctx context.Context, id uint) (int, error)
}
