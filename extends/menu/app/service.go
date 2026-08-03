package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	coreCache "kingfisher/core/cache"
	"kingfisher/extends/menu/domain"
	"kingfisher/extends/menu/port"
)

type MenuService struct {
	repo  port.MenuRepository
	cache coreCache.Cache
}

func NewMenuService(repo port.MenuRepository, cache coreCache.Cache) *MenuService {
	return &MenuService{repo: repo, cache: cache}
}

func (s *MenuService) GetTree(ctx context.Context) ([]domain.Menu, error) {
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, "menu:tree"); err == nil && cached != "" {
			var tree []domain.Menu
			if json.Unmarshal([]byte(cached), &tree) == nil {
				return tree, nil
			}
		}
	}
	menus, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	tree := buildTree(menus, 0)
	if s.cache != nil {
		if data, err := json.Marshal(tree); err == nil {
			_ = s.cache.Set(ctx, "menu:tree", string(data), 10*time.Minute)
		}
	}
	return tree, nil
}

func (s *MenuService) GetTreeForRole(ctx context.Context, roleID uint) ([]domain.Menu, error) {
	menus, err := s.repo.FindByRoleIDs(ctx, []uint{roleID})
	if err != nil {
		return nil, err
	}
	return buildTree(menus, 0), nil
}

func (s *MenuService) GetByID(ctx context.Context, id uint) (*domain.Menu, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *MenuService) Create(ctx context.Context, m *domain.Menu) error {
	if m.Type == 2 && m.Path != "" {
		menus, _ := s.repo.FindAll(ctx)
		for _, existing := range menus {
			if existing.Path == m.Path && existing.Path != "" {
				return fmt.Errorf("menu path already exists")
			}
		}
	}
	return s.repo.Create(ctx, m)
}

func (s *MenuService) Update(ctx context.Context, id uint, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}

func (s *MenuService) Delete(ctx context.Context, id uint) error {
	hasChildren, _ := s.repo.HasChildren(ctx, id)
	if hasChildren {
		return fmt.Errorf("menu has children")
	}
	return s.repo.Delete(ctx, id)
}

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
