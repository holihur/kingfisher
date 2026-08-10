package adapter

import (
	"context"

	"gorm.io/gorm"

	"kingfisher/extends/menu/domain"
)

type MenuRepo struct{ db *gorm.DB }

func NewMenuRepo(db *gorm.DB) *MenuRepo { return &MenuRepo{db: db} }
func (r *MenuRepo) FindAll(ctx context.Context) ([]domain.Menu, error) {
	var pos []menuPO
	err := r.db.WithContext(ctx).Order("sort ASC").Find(&pos).Error
	if err != nil {
		return nil, err
	}
	menus := make([]domain.Menu, len(pos))
	for i, p := range pos {
		menus[i] = domain.Menu{ID: p.ID, ParentID: p.ParentID, Name: p.Name, Path: p.Path, Component: p.Component, Icon: p.Icon, Sort: p.Sort, Type: p.Type, Permission: p.Permission, Status: p.Status, Version: p.Version, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
	}
	return menus, nil
}
func (r *MenuRepo) FindByID(ctx context.Context, id uint) (*domain.Menu, error) {
	var po menuPO
	err := r.db.WithContext(ctx).First(&po, id).Error
	if err != nil {
		return nil, err
	}
	return &domain.Menu{ID: po.ID, ParentID: po.ParentID, Name: po.Name, Path: po.Path, Component: po.Component, Icon: po.Icon, Sort: po.Sort, Type: po.Type, Permission: po.Permission, Status: po.Status, Version: po.Version, CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt}, nil
}
func (r *MenuRepo) FindByParentID(ctx context.Context, parentID uint) ([]domain.Menu, error) {
	var pos []menuPO
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Order("sort ASC").Find(&pos).Error
	if err != nil {
		return nil, err
	}
	menus := make([]domain.Menu, len(pos))
	for i, p := range pos {
		menus[i] = domain.Menu{ID: p.ID, ParentID: p.ParentID, Name: p.Name, Path: p.Path, Component: p.Component, Icon: p.Icon, Sort: p.Sort, Type: p.Type, Permission: p.Permission, Status: p.Status, Version: p.Version, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
	}
	return menus, nil
}
func (r *MenuRepo) Create(ctx context.Context, m *domain.Menu) error {
	return r.db.WithContext(ctx).Create(&menuPO{ParentID: m.ParentID, Name: m.Name, Path: m.Path, Component: m.Component, Icon: m.Icon, Sort: m.Sort, Type: m.Type, Permission: m.Permission, Status: m.Status, Version: m.Version}).Error
}
func (r *MenuRepo) Update(ctx context.Context, id uint, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&menuPO{}).Where("id = ?", id).Updates(updates).Error
}
func (r *MenuRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&menuPO{}, id).Error
}
func (r *MenuRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&menuPO{}).Error
}
func (r *MenuRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	return r.db.WithContext(ctx).Model(&menuPO{}).Where("id IN ?", ids).Update("status", status).Error
}
func (r *MenuRepo) HasChildren(ctx context.Context, parentID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&menuPO{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count > 0, err
}
func (r *MenuRepo) FindByRoleIDs(ctx context.Context, roleIDs []uint) ([]domain.Menu, error) {
	var pos []menuPO
	err := r.db.WithContext(ctx).
		Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
		Where("role_menus.role_id IN ?", roleIDs).
		Distinct().
		Order("sort ASC").
		Find(&pos).Error
	if err != nil {
		return nil, err
	}
	menus := make([]domain.Menu, len(pos))
	for i, p := range pos {
		menus[i] = domain.Menu{ID: p.ID, ParentID: p.ParentID, Name: p.Name, Path: p.Path, Component: p.Component, Icon: p.Icon, Sort: p.Sort, Type: p.Type, Permission: p.Permission, Status: p.Status, Version: p.Version, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
	}
	return menus, nil
}
