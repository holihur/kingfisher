package port

import (
	"context"

	"kingfisher/core/dataaccess"
	"kingfisher/core/query"
	"kingfisher/extends/worktask/domain"
)

type Repository interface {
	List(context.Context, *query.Query, dataaccess.Scope) ([]domain.Task, int64, error)
	GetByID(context.Context, uint, dataaccess.Scope) (*domain.Task, error)
	Create(context.Context, *domain.Task) error
	Update(context.Context, uint, map[string]any, dataaccess.Scope) error
	Delete(context.Context, uint, dataaccess.Scope) error
}
