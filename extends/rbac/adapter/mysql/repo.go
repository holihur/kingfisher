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
	err := r.db.WithContext(ctx).Raw(`
SELECT DISTINCT p.code FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN (
    SELECT ur.role_id FROM user_roles ur WHERE ur.user_id = ?
    UNION
    SELECT dr.role_id FROM department_roles dr
    JOIN user_departments ud ON ud.department_id = dr.department_id
    WHERE ud.user_id = ?
) r ON r.role_id = rp.role_id`, userID, userID).Scan(&codes).Error
	if err != nil {
		return nil, err
	}
	var parentID *uint
	if err := r.db.WithContext(ctx).Table("users").Select("parent_id").Where("id = ?", userID).Scan(&parentID).Error; err == nil && parentID != nil {
		var parentCodes []string
		if err := r.db.WithContext(ctx).Raw(`
SELECT DISTINCT p.code FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN (
    SELECT ur.role_id FROM user_roles ur WHERE ur.user_id = ?
    UNION
    SELECT dr.role_id FROM department_roles dr
    JOIN user_departments ud ON ud.department_id = dr.department_id
    WHERE ud.user_id = ?
) r ON r.role_id = rp.role_id`, *parentID, *parentID).Scan(&parentCodes).Error; err == nil {
			allowed := make(map[string]bool, len(parentCodes))
			for _, c := range parentCodes {
				allowed[c] = true
			}
			filtered := make([]string, 0, len(codes))
			for _, c := range codes {
				if allowed[c] {
					filtered = append(filtered, c)
				}
			}
			return filtered, nil
		}
	}
	return codes, nil
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

func (r *RoleRepo) GetDataScopes(ctx context.Context, roleIDs []uint, resource string) (map[uint]string, error) {
	var rows []roleDataScopePO
	if err := r.db.WithContext(ctx).Where("role_id IN ? AND resource = ?", roleIDs, resource).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[uint]string, len(rows))
	for _, row := range rows {
		result[row.RoleID] = row.ScopeType
	}
	return result, nil
}

func (r *RoleRepo) SetDataScope(ctx context.Context, roleID uint, resource, scopeType string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ? AND resource = ?", roleID, resource).Delete(&roleDataScopePO{}).Error; err != nil {
			return err
		}
		return tx.Create(&roleDataScopePO{RoleID: roleID, Resource: resource, ScopeType: scopeType}).Error
	})
}

func (r *RoleRepo) GetUserDepartmentIDs(ctx context.Context, userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Table("user_departments").Where("user_id = ?", userID).Pluck("department_id", &ids).Error
	return ids, err
}

func (r *RoleRepo) GetDepartmentSubtreeIDs(ctx context.Context, roots []uint) ([]uint, error) {
	type row struct{ ID, ParentID uint }
	var rows []row
	if err := r.db.WithContext(ctx).Table("departments").Select("id, parent_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	seen := make(map[uint]bool, len(roots))
	for _, id := range roots {
		seen[id] = true
	}
	changed := true
	for changed {
		changed = false
		for _, item := range rows {
			if seen[item.ParentID] && !seen[item.ID] {
				seen[item.ID] = true
				changed = true
			}
		}
	}
	ids := make([]uint, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	return ids, nil
}
