package port

import (
	"context"

	"kingfisher/extends/config/domain"
)

type ConfigRepository interface {
	GetAll(ctx context.Context) ([]domain.SystemConfig, error)
	GetByKey(ctx context.Context, key string) (*domain.SystemConfig, error)
	Set(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}
