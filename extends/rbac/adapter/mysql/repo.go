package adapter
import (
    "context"; "gorm.io/gorm"
    "kingfisher/extends/rbac/domain"; "kingfisher/extends/rbac/port"
)
type RoleRepo struct{ db *gorm.DB }
var _ port.RoleRepository = (*RoleRepo)(nil)
func NewRoleRepo(db *gorm.DB) *RoleRepo { return &RoleRepo{db: db} }
func (r *RoleRepo) FindAll(ctx context.Context) ([]domain.Role, error) {
    var pos []rolePO; err := r.db.WithContext(ctx).Find(&pos).Error; if err != nil { return nil, err }
    roles := make([]domain.Role, len(pos))
    for i, p := range pos { roles[i] = domain.Role{ID: p.ID, Name: p.Name, Code: p.Code, Description: p.Description, Status: p.Status, Level: p.Level, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt} }
    return roles, nil
}
func (r *RoleRepo) FindByID(ctx context.Context, id uint) (*domain.Role, error) {
    var po rolePO; err := r.db.WithContext(ctx).First(&po, id).Error; if err != nil { return nil, err }
    return &domain.Role{ID: po.ID, Name: po.Name, Code: po.Code, Description: po.Description, Status: po.Status, Level: po.Level}, nil
}
func (r *RoleRepo) Create(ctx context.Context, role *domain.Role) error {
    return r.db.WithContext(ctx).Create(&rolePO{Name: role.Name, Code: role.Code, Description: role.Description, Status: role.Status, Level: role.Level}).Error
}
func (r *RoleRepo) Update(ctx context.Context, id uint, updates map[string]any) error {
    return r.db.WithContext(ctx).Model(&rolePO{}).Where("id = ?", id).Updates(updates).Error
}
func (r *RoleRepo) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&rolePO{}, id).Error
}
func (r *RoleRepo) AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        tx.Where("role_id = ?", roleID).Delete(&struct{ RoleID, PermissionID uint }{})
        for _, pid := range permIDs {
            tx.Create(&struct{ RoleID, PermissionID uint }{roleID, pid})
        }
        return nil
    })
}
func (r *RoleRepo) AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        tx.Where("role_id = ?", roleID).Delete(&struct{ RoleID, MenuID uint }{})
        for _, mid := range menuIDs {
            tx.Create(&struct{ RoleID, MenuID uint }{roleID, mid})
        }
        return nil
    })
}
func (r *RoleRepo) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
    var codes []string
    err := r.db.WithContext(ctx).Raw(
        "SELECT DISTINCT p.code FROM permissions p JOIN role_permissions rp ON p.id = rp.permission_id JOIN users u ON u.role_id = rp.role_id WHERE u.id = ?", userID).Scan(&codes).Error
    return codes, err
}
type PermRepo struct{ db *gorm.DB }
func NewPermRepo(db *gorm.DB) *PermRepo { return &PermRepo{db: db} }

func (r *RoleRepo) GetRolePermissions(ctx context.Context, roleID uint) ([]domain.Permission, error) {
	var pos []permissionPO
	err := r.db.WithContext(ctx).Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").Where("rp.role_id = ?", roleID).Find(&pos).Error
	if err != nil { return nil, err }
	perms := make([]domain.Permission, len(pos))
	for i, p := range pos { perms[i] = domain.Permission{ID: p.ID, Name: p.Name, Code: p.Code, Resource: p.Resource, Action: p.Action} }
	return perms, nil
}

func (r *RoleRepo) GetRoleMenus(ctx context.Context, roleID uint) ([]domain.Menu, error) {
	type menuPO struct { ID uint; ParentID uint; Name string; Path string; Component string; Icon string; Sort int; Type int; Permission string; Status int }
	var pos []menuPO
	err := r.db.WithContext(ctx).Joins("JOIN role_menus rm ON rm.menu_id = menus.id").Where("rm.role_id = ?", roleID).Order("sort ASC").Find(&pos).Error
	if err != nil { return nil, err }
	menus := make([]domain.Menu, len(pos))
	for i, p := range pos { menus[i] = domain.Menu{ID: p.ID, ParentID: p.ParentID, Name: p.Name, Path: p.Path, Component: p.Component, Icon: p.Icon, Sort: p.Sort, Type: p.Type, Permission: p.Permission} }
	return menus, nil
}

func (r *PermRepo) FindAll(ctx context.Context) ([]domain.Permission, error) {
    var pos []permissionPO; err := r.db.WithContext(ctx).Find(&pos).Error; if err != nil { return nil, err }
    perms := make([]domain.Permission, len(pos))
    for i, p := range pos { perms[i] = domain.Permission{ID: p.ID, Name: p.Name, Code: p.Code, Resource: p.Resource, Action: p.Action} }
    return perms, nil
}
