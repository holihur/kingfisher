package adapter

import (
	"context"

	"gorm.io/gorm"

	"kingfisher/core/query"
	"kingfisher/extends/rbac/domain"
	"kingfisher/extends/rbac/port"
)

type RoleRepo struct{ db *gorm.DB }

var _ port.RoleRepository = (*RoleRepo)(nil)

func NewRoleRepo(db *gorm.DB) *RoleRepo { return &RoleRepo{db: db} }
func (r *RoleRepo) FindAll(ctx context.Context, q *query.Query) ([]domain.Role, int64, error) {
	var pos []rolePO
	total, err := q.Find(r.db.WithContext(ctx).Model(&rolePO{}), &pos)
	if err != nil {
		return nil, 0, err
	}
	roles := make([]domain.Role, len(pos))
	for i, p := range pos {
		roles[i] = *toRole(&p)
	}
	return roles, total, nil
}
func (r *RoleRepo) FindByCode(ctx context.Context, code string) (*domain.Role, error) {
	var po rolePO
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error
	if err != nil {
		return nil, err
	}
	return toRole(&po), nil
}

func (r *RoleRepo) FindByID(ctx context.Context, id uint) (*domain.Role, error) {
	var po rolePO
	err := r.db.WithContext(ctx).First(&po, id).Error
	if err != nil {
		return nil, err
	}
	return toRole(&po), nil
}
func (r *RoleRepo) Create(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(&rolePO{Name: role.Name, Code: role.Code, Description: role.Description, Status: role.Status, Level: role.Level, LandingPage: role.LandingPage}).Error
}
func (r *RoleRepo) Update(ctx context.Context, id uint, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&rolePO{}).Where("id = ?", id).Updates(updates).Error
}
func (r *RoleRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&rolePO{}, id).Error
}
func (r *RoleRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&rolePO{}).Error
}
func (r *RoleRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	return r.db.WithContext(ctx).Model(&rolePO{}).Where("id IN ?", ids).Update("status", status).Error
}
func (r *RoleRepo) AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&rolePermissionPO{}).Error; err != nil {
			return err
		}
		for _, pid := range permIDs {
			if err := tx.Create(&rolePermissionPO{RoleID: roleID, PermissionID: pid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *RoleRepo) AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&roleMenuPO{}).Error; err != nil {
			return err
		}
		for _, mid := range menuIDs {
			if err := tx.Create(&roleMenuPO{RoleID: roleID, MenuID: mid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *RoleRepo) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).Raw(
		"SELECT DISTINCT p.code FROM permissions p JOIN role_permissions rp ON p.id = rp.permission_id JOIN user_roles ur ON ur.role_id = rp.role_id WHERE ur.user_id = ?", userID).Scan(&codes).Error
	return codes, err
}

type PermRepo struct{ db *gorm.DB }

func NewPermRepo(db *gorm.DB) *PermRepo { return &PermRepo{db: db} }

func (r *RoleRepo) GetRolePermissions(ctx context.Context, roleID uint) ([]domain.Permission, error) {
	var pos []permissionPO
	err := r.db.WithContext(ctx).Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").Where("rp.role_id = ?", roleID).Find(&pos).Error
	if err != nil {
		return nil, err
	}
	perms := make([]domain.Permission, len(pos))
	for i, p := range pos {
		perms[i] = domain.Permission{ID: p.ID, Name: p.Name, Code: p.Code, Resource: p.Resource, Action: p.Action, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
	}
	return perms, nil
}

func (r *RoleRepo) GetRoleMenus(ctx context.Context, roleID uint) ([]domain.Menu, error) {
	var pos []menuPO
	err := r.db.WithContext(ctx).Joins("JOIN role_menus rm ON rm.menu_id = menus.id").Where("rm.role_id = ?", roleID).Order("sort ASC").Find(&pos).Error
	if err != nil {
		return nil, err
	}
	menus := make([]domain.Menu, len(pos))
	for i, p := range pos {
		menus[i] = domain.Menu{ID: p.ID, ParentID: p.ParentID, Name: p.Name, Path: p.Path, Component: p.Component, Icon: p.Icon, Sort: p.Sort}
	}
	return menus, nil
}

func (r *PermRepo) FindAll(ctx context.Context) ([]domain.Permission, error) {
	var pos []permissionPO
	err := r.db.WithContext(ctx).Find(&pos).Error
	if err != nil {
		return nil, err
	}
	perms := make([]domain.Permission, len(pos))
	for i, p := range pos {
		perms[i] = domain.Permission{ID: p.ID, Name: p.Name, Code: p.Code, Resource: p.Resource, Action: p.Action, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
	}
	return perms, nil
}
