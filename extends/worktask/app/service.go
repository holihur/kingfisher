package app

import (
	"context"

	"kingfisher/core/dataaccess"
	"kingfisher/core/query"
	"kingfisher/extends/worktask/domain"
	"kingfisher/extends/worktask/port"
)

type Service struct{ repo port.Repository }

func NewService(repo port.Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, q *query.Query, scope dataaccess.Scope) ([]domain.Task, int64, error) {
	return s.repo.List(ctx, q, scope)
}

func (s *Service) GetByID(ctx context.Context, id uint, scope dataaccess.Scope) (*domain.Task, error) {
	return s.repo.GetByID(ctx, id, scope)
}

func (s *Service) Create(ctx context.Context, task *domain.Task) error {
	return s.repo.Create(ctx, task)
}

func (s *Service) Update(ctx context.Context, id uint, updates map[string]any, scope dataaccess.Scope) error {
	return s.repo.Update(ctx, id, updates, scope)
}

func (s *Service) Delete(ctx context.Context, id uint, scope dataaccess.Scope) error {
	return s.repo.Delete(ctx, id, scope)
}
