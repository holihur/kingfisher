package app

import (
	"context"

	adapter "kingfisher/extends/menu/adapter/mysql"
	"kingfisher/extends/menu/domain"
)

type MenuService struct{ repo *adapter.MenuRepo }

func NewMenuService(repo *adapter.MenuRepo) *MenuService { return &MenuService{repo: repo} }
func (s *MenuService) GetTree(ctx context.Context) ([]domain.Menu, error) {
	menus, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return buildTree(menus, 0), nil
}
func (s *MenuService) GetByID(ctx context.Context, id uint) (*domain.Menu, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *MenuService) Create(ctx context.Context, m *domain.Menu) error { return s.repo.Create(ctx, m) }
func (s *MenuService) Update(ctx context.Context, id uint, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}
func (s *MenuService) Delete(ctx context.Context, id uint) error { return s.repo.Delete(ctx, id) }
func buildTree(menus []domain.Menu, parentID uint) []domain.Menu {
	var result []domain.Menu
	for _, m := range menus {
		if m.ParentID == parentID {
			m.Children = buildTree(menus, m.ID)
			result = append(result, m)
		}
	}
	return result
}
