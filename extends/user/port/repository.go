package port

import (
	"context"

	"kingfisher/core/query"
	"kingfisher/extends/user/domain"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uint) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindAll(ctx context.Context, q *query.Query) ([]domain.User, int64, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, id uint, updates map[string]any) error
	Delete(ctx context.Context, id uint) error
	DeleteBatch(ctx context.Context, ids []uint) error
	UpdateStatusBatch(ctx context.Context, ids []uint, status int) error
	IncrementSessionVersion(ctx context.Context, id uint) error
	GetSessionVersion(ctx context.Context, id uint) (int, error)
}
