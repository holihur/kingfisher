package adapter

import (
	"context"
	"time"

	"gorm.io/gorm"

	"kingfisher/core/jwt"
	"kingfisher/core/query"
	"kingfisher/extends/user/domain"
	"kingfisher/extends/user/port"
)

type UserRepo struct {
	db *gorm.DB
}

var _ port.UserRepository = (*UserRepo)(nil)

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var po userPO
	err := r.db.WithContext(ctx).Preload("Roles").Preload("Departments").First(&po, id).Error
	if err != nil {
		return nil, err
	}
	u := po.toDomain()
	if err := r.mergeDeptRoles(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var po userPO
	err := r.db.WithContext(ctx).Preload("Roles").Preload("Departments").Where("username = ?", username).First(&po).Error
	if err != nil {
		return nil, err
	}
	u := po.toDomain()
	if err := r.mergeDeptRoles(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var po userPO
	err := r.db.WithContext(ctx).Preload("Roles").Preload("Departments").Where("email = ?", email).First(&po).Error
	if err != nil {
		return nil, err
	}
	u := po.toDomain()
	if err := r.mergeDeptRoles(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) FindAll(ctx context.Context, q *query.Query) ([]domain.User, int64, error) {
	var pos []userPO
	base := r.db.WithContext(ctx).Model(&userPO{}).Preload("Roles").Preload("Departments")

	// role_id 筛选 = 用户「拥有该角色」（有效角色 = 直接 ∪ 部门继承）。eq/in → 拥有任一；ne → 不拥有。
	// department_id 筛选 = 用户「属于该部门」（user_departments 成员过滤）。
	var roleVals, notRoleVals []any
	var deptVals, notDeptVals []any
	rest := make([]query.Condition, 0, len(q.Filters))
	for _, f := range q.Filters {
		switch {
		case f.Field == "role_id" && f.Op == query.OpIn:
			if arr, ok := f.Value.([]any); ok {
				roleVals = append(roleVals, arr...)
			}
		case f.Field == "role_id" && f.Op == query.OpEq:
			roleVals = append(roleVals, f.Value)
		case f.Field == "role_id" && f.Op == query.OpNe:
			notRoleVals = append(notRoleVals, f.Value)
		case f.Field == "department_id" && f.Op == query.OpIn:
			if arr, ok := f.Value.([]any); ok {
				deptVals = append(deptVals, arr...)
			}
		case f.Field == "department_id" && f.Op == query.OpEq:
			deptVals = append(deptVals, f.Value)
		case f.Field == "department_id" && f.Op == query.OpNe:
			notDeptVals = append(notDeptVals, f.Value)
		default:
			rest = append(rest, f)
		}
	}
	if len(roleVals) > 0 {
		base = base.Where(`EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = users.id AND ur.role_id IN ?)
			OR EXISTS (SELECT 1 FROM department_roles dr JOIN user_departments ud ON ud.department_id = dr.department_id WHERE ud.user_id = users.id AND dr.role_id IN ?)`, roleVals, roleVals)
	}
	if len(notRoleVals) > 0 {
		base = base.Where(`NOT EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = users.id AND ur.role_id IN ?)
			AND NOT EXISTS (SELECT 1 FROM department_roles dr JOIN user_departments ud ON ud.department_id = dr.department_id WHERE ud.user_id = users.id AND dr.role_id IN ?)`, notRoleVals, notRoleVals)
	}
	if len(deptVals) > 0 {
		base = base.Where("EXISTS (SELECT 1 FROM user_departments ud WHERE ud.user_id = users.id AND ud.department_id IN ?)", deptVals)
	}
	if len(notDeptVals) > 0 {
		base = base.Where("NOT EXISTS (SELECT 1 FROM user_departments ud WHERE ud.user_id = users.id AND ud.department_id IN ?)", notDeptVals)
	}
	// 拷贝 Query，避免污染调用方；role_id/department_id 条件已由上方 EXISTS 处理，不再交给 q.Find
	q2 := *q
	q2.Filters = rest

	total, err := q2.Find(base, &pos)
	if err != nil {
		return nil, 0, err
	}
	users := make([]domain.User, len(pos))
	for i, p := range pos {
		users[i] = *p.toDomain()
	}
	if err := r.attachDeptRoles(ctx, users); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// attachDeptRoles 为一批用户合并「部门继承角色」进有效角色（去重）。
// DirectRoleIDs 保持为直接分配（user_roles），RoleIDs/Roles 变为有效角色 = 直接 ∪ 部门继承。
func (r *UserRepo) attachDeptRoles(ctx context.Context, users []domain.User) error {
	if len(users) == 0 {
		return nil
	}
	ids := make([]uint, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	// user_departments ⋈ department_roles → {userID: [roleID]}
	var rows []struct {
		UserID uint
		RoleID uint
	}
	if err := r.db.WithContext(ctx).Table("department_roles dr").
		Joins("JOIN user_departments ud ON ud.department_id = dr.department_id").
		Where("ud.user_id IN ?", ids).
		Select("ud.user_id, dr.role_id").
		Scan(&rows).Error; err != nil {
		return err
	}
	inherited := make(map[uint][]uint)
	roleIDSet := make(map[uint]struct{})
	for _, row := range rows {
		inherited[row.UserID] = append(inherited[row.UserID], row.RoleID)
		roleIDSet[row.RoleID] = struct{}{}
	}
	if len(roleIDSet) == 0 {
		return nil
	}
	roleIDs := make([]uint, 0, len(roleIDSet))
	for rid := range roleIDSet {
		roleIDs = append(roleIDs, rid)
	}
	var rps []rolePO
	if err := r.db.WithContext(ctx).Where("id IN ?", roleIDs).Find(&rps).Error; err != nil {
		return err
	}
	nameByID := make(map[uint]domain.Role, len(rps))
	for _, rp := range rps {
		nameByID[rp.ID] = domain.Role{ID: rp.ID, Name: rp.Name, Code: rp.Code}
	}
	for i := range users {
		have := make(map[uint]bool, len(users[i].RoleIDs))
		for _, rid := range users[i].RoleIDs {
			have[rid] = true
		}
		for _, rid := range inherited[users[i].ID] {
			if have[rid] {
				continue
			}
			have[rid] = true
			users[i].RoleIDs = append(users[i].RoleIDs, rid)
			if rl, ok := nameByID[rid]; ok {
				users[i].Roles = append(users[i].Roles, &rl)
			}
		}
	}
	return nil
}

// mergeDeptRoles 单用户版：合并部门继承角色进有效角色，并写回 *u。
func (r *UserRepo) mergeDeptRoles(ctx context.Context, u *domain.User) error {
	one := []domain.User{*u}
	if err := r.attachDeptRoles(ctx, one); err != nil {
		return err
	}
	*u = one[0]
	return nil
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	po := userPO{Username: u.Username, Nickname: u.Nickname, Password: u.Password, Email: u.Email, Status: u.Status}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&po).Error; err != nil {
			return err
		}
		u.ID = po.ID
		u.CreatedAt = po.CreatedAt
		u.UpdatedAt = po.UpdatedAt
		// 写入 user_roles 关联
		if err := r.setUserRoles(tx, po.ID, u.RoleIDs); err != nil {
			return err
		}
		// 写入 user_departments 关联
		return r.setUserDepartments(tx, po.ID, u.DeptIDs)
	})
	return err
}

// setUserRoles 替换用户的角色关联（先删后插）
func (r *UserRepo) setUserRoles(tx *gorm.DB, userID uint, roleIDs []uint) error {
	if err := tx.Where("user_id = ?", userID).Delete(&userRolePO{}).Error; err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if err := tx.Create(&userRolePO{UserID: userID, RoleID: roleID}).Error; err != nil {
			return err
		}
	}
	return nil
}

// setUserDepartments 替换用户的部门关联（先删后插）
func (r *UserRepo) setUserDepartments(tx *gorm.DB, userID uint, deptIDs []uint) error {
	if err := tx.Where("user_id = ?", userID).Delete(&userDepartmentPO{}).Error; err != nil {
		return err
	}
	for _, deptID := range deptIDs {
		if err := tx.Create(&userDepartmentPO{UserID: userID, DepartmentID: deptID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *UserRepo) Update(ctx context.Context, id uint, updates map[string]any) error {
	// role_ids / dept_ids 变更单独处理：普通字段与关联表在同一事务内更新
	if roleIDs, ok := updates["role_ids"].([]uint); ok {
		delete(updates, "role_ids")
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&userPO{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
			if err := r.setUserRoles(tx, id, roleIDs); err != nil {
				return err
			}
			// 部门也可能同时变更
			if deptIDs, ok := updates["dept_ids"].([]uint); ok {
				delete(updates, "dept_ids")
				return r.setUserDepartments(tx, id, deptIDs)
			}
			return nil
		})
	}
	if deptIDs, ok := updates["dept_ids"].([]uint); ok {
		delete(updates, "dept_ids")
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&userPO{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
			return r.setUserDepartments(tx, id, deptIDs)
		})
	}
	return r.db.WithContext(ctx).Model(&userPO{}).Where("id = ?", id).Updates(updates).Error
}

func (r *UserRepo) Delete(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&userPO{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

// DeleteBatch 批量软删除（与 Delete 一致，置 deleted_at）
func (r *UserRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&userPO{}).Where("id IN ?", ids).Update("deleted_at", &now).Error
}

// UpdateStatusBatch 批量启用/禁用
func (r *UserRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	return r.db.WithContext(ctx).Model(&userPO{}).Where("id IN ?", ids).Update("status", status).Error
}

func (r *UserRepo) IncrementSessionVersion(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&userPO{}).Where("id = ?", id).
		Update("session_version", gorm.Expr("session_version + 1")).Error
}

// NewSessionVersionProvider creates a jwt.SessionVersionProvider from a db handle.
func NewSessionVersionProvider(db *gorm.DB) jwt.SessionVersionProvider {
	r := NewUserRepo(db)
	return r.GetSessionVersion
}

func (r *UserRepo) GetSessionVersion(ctx context.Context, id uint) (int, error) {
	var sv int
	err := r.db.WithContext(ctx).Model(&userPO{}).Select("session_version").Where("id = ?", id).Scan(&sv).Error
	return sv, err
}
