package port

import (
	"context"

	"kingfisher/core/query"
	"kingfisher/extends/config/domain"
)

type ConfigRepository interface {
	List(ctx context.Context, q *query.Query) ([]domain.SystemConfig, int64, error)
	GetPublicAll(ctx context.Context) ([]domain.SystemConfig, error)
	GetByKey(ctx context.Context, key string) (*domain.SystemConfig, error)
	GetPublicByKey(ctx context.Context, key string) (*domain.SystemConfig, error)
	Set(ctx context.Context, key, value string, isPublic bool, version, render, renderOptions string, groupID uint) error
	Delete(ctx context.Context, key string) error
}

type ConfigGroupRepository interface {
	List(ctx context.Context) ([]domain.ConfigGroup, error)
	Create(ctx context.Context, name string, sort int) (*domain.ConfigGroup, error)
	Update(ctx context.Context, id uint, name string, sort int) error
	Delete(ctx context.Context, id uint) error
}
