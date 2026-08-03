package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"kingfisher/core/cache"
	"kingfisher/extends/rbac/domain"
	"kingfisher/extends/rbac/port"
)

type RoleService struct {
	repo  port.RoleRepository
	cache cache.Cache
}

func NewRoleService(repo port.RoleRepository, c cache.Cache) *RoleService {
	return &RoleService{repo: repo, cache: c}
}
func (s *RoleService) List(ctx context.Context) ([]domain.Role, error) { return s.repo.FindAll(ctx) }
func (s *RoleService) GetByID(ctx context.Context, id uint) (*domain.Role, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *RoleService) Create(ctx context.Context, role *domain.Role) error {
	return s.repo.Create(ctx, role)
}
func (s *RoleService) Update(ctx context.Context, id uint, updates map[string]any) error {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if role.Code == "admin" {
		return fmt.Errorf("cannot modify admin role")
	}
	return s.repo.Update(ctx, id, updates)
}
func (s *RoleService) Delete(ctx context.Context, id uint) error {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if role.Code == "admin" {
		return fmt.Errorf("cannot delete admin role")
	}
	return s.repo.Delete(ctx, id)
}
func (s *RoleService) AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error {
	if err := s.repo.AssignPermissions(ctx, roleID, permIDs); err != nil {
		return err
	}
	if s.cache != nil {
		// Invalidate all user permission caches — SCAN-based in production
		_ = s.cache.Delete(ctx, "user:perms:*")
	}
	return nil
}
func (s *RoleService) AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	if err := s.repo.AssignMenus(ctx, roleID, menuIDs); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, "menu:role:"+strconv.Itoa(int(roleID)))
		_ = s.cache.Delete(ctx, "menu:tree")
	}
	return nil
}
func (s *RoleService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	key := "user:perms:" + strconv.Itoa(int(userID))
	if s.cache != nil {
		if val, err := s.cache.Get(ctx, key); err == nil && val != "" {
			return strSlice(val), nil
		}
	}
	codes, err := s.repo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, key, strings.Join(codes, ","), 30*60*1e9)
	} // 30min in ns
	return codes, nil
}
func (s *RoleService) GetRolePermissions(ctx context.Context, roleID uint) ([]domain.Permission, error) {
	rp, err := s.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return rp, nil
}
func (s *RoleService) GetRoleMenus(ctx context.Context, roleID uint) ([]domain.Menu, error) {
	rm, err := s.repo.GetRoleMenus(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return rm, nil
}
func strSlice(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
type PermService struct{ repo port.PermissionRepository }

func NewPermService(repo port.PermissionRepository) *PermService { return &PermService{repo: repo} }
func (s *PermService) List(ctx context.Context) ([]domain.Permission, error) {
	return s.repo.FindAll(ctx)
}
