package adapter

import (
	"context"

	"gorm.io/gorm"

	"kingfisher/core/query"
	"kingfisher/extends/department/domain"
	"kingfisher/extends/department/port"
)

type DepartmentRepo struct{ db *gorm.DB }

var _ port.DepartmentRepository = (*DepartmentRepo)(nil)

func NewDepartmentRepo(db *gorm.DB) *DepartmentRepo { return &DepartmentRepo{db: db} }

// FindAll 返回全部部门（含挂载角色），按 sort/id 排序，由 service 层 buildTree 组树。
func (r *DepartmentRepo) FindAll(ctx context.Context) ([]domain.Department, error) {
	var pos []departmentPO
	if err := r.db.WithContext(ctx).Order("sort ASC").Order("id ASC").Find(&pos).Error; err != nil {
		return nil, err
	}
	depts := make([]domain.Department, len(pos))
	for i, p := range pos {
		depts[i] = *p.toDepartment()
	}
	if err := r.attachRoles(ctx, depts); err != nil {
		return nil, err
	}
	return depts, nil
}

// ListPage 分页部门列表（query DSL）。支持 subtree_id 筛选：返回该部门及其全部子孙。
func (r *DepartmentRepo) ListPage(ctx context.Context, q *query.Query) ([]domain.Department, int64, error) {
	// subtree_id 从通用 filter 抽出来，解析为子孙 ID 集合，再交给 q.Find（避免不存在的列）
	var subtreeID uint
	var subtreeSet bool
	rest := make([]query.Condition, 0, len(q.Filters))
	for _, f := range q.Filters {
		if f.Field == "subtree_id" && f.Op == query.OpEq {
			if v, ok := toUint(f.Value); ok {
				subtreeID = v
				subtreeSet = true
			}
			continue
		}
		rest = append(rest, f)
	}
	base := r.db.WithContext(ctx).Model(&departmentPO{})
	if subtreeSet {
		ids, err := r.SubtreeIDs(ctx, subtreeID)
		if err != nil {
			return nil, 0, err
		}
		if len(ids) == 0 {
			return []domain.Department{}, 0, nil
		}
		base = base.Where("id IN ?", ids)
	}
	q2 := *q
	q2.Filters = rest
	var pos []departmentPO
	total, err := q2.Find(base, &pos)
	if err != nil {
		return nil, 0, err
	}
	depts := make([]domain.Department, len(pos))
	for i, p := range pos {
		depts[i] = *p.toDepartment()
	}
	if err := r.attachRoles(ctx, depts); err != nil {
		return nil, 0, err
	}
	return depts, total, nil
}

// toUint 从 query 条件值中提取 uint（coerce 对 TypeUint 返回 uint64，这里做兼容转换）。
// 仅接受 uint64/uint：query 包对 TypeUint 恒产生 uint64，其他类型视为非法。
func toUint(v any) (uint, bool) {
	switch n := v.(type) {
	case uint64:
		return uint(n), true
	case uint:
		return n, true
	}
	return 0, false
}

// SubtreeIDs 返回 rootID 及其全部子孙的 ID 集合（部门树通常很小，全量扫一次）。
func (r *DepartmentRepo) SubtreeIDs(ctx context.Context, rootID uint) ([]uint, error) {
	var pos []departmentPO
	if err := r.db.WithContext(ctx).Select("id", "parent_id").Find(&pos).Error; err != nil {
		return nil, err
	}
	children := make(map[uint][]uint)
	for _, p := range pos {
		children[p.ParentID] = append(children[p.ParentID], p.ID)
	}
	ids := []uint{rootID}
	queue := []uint{rootID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, c := range children[cur] {
			ids = append(ids, c)
			queue = append(queue, c)
		}
	}
	return ids, nil
}

// attachRoles 为一批部门挂载角色（一次性查询，避免 N+1）。
func (r *DepartmentRepo) attachRoles(ctx context.Context, depts []domain.Department) error {
	roleIDs, err := r.allDepartmentRoleIDs(ctx)
	if err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	roles, err := r.rolesByIDs(ctx, roleIDs)
	if err != nil {
		return err
	}
	for i := range depts {
		if ids, ok := roleIDs[depts[i].ID]; ok {
			depts[i].RoleIDs = ids
			for _, id := range ids {
				if rl, ok := roles[id]; ok {
					depts[i].Roles = append(depts[i].Roles, &domain.Role{ID: rl.ID, Name: rl.Name, Code: rl.Code})
				}
			}
		}
	}
	return nil
}

// GetByID 返回单个部门（含挂载角色）。
func (r *DepartmentRepo) GetByID(ctx context.Context, id uint) (*domain.Department, error) {
	var po departmentPO
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error; err != nil {
		return nil, err
	}
	d := po.toDepartment()
	var drs []departmentRolePO
	if err := r.db.WithContext(ctx).Where("department_id = ?", id).Find(&drs).Error; err != nil {
		return nil, err
	}
	for _, dr := range drs {
		d.RoleIDs = append(d.RoleIDs, dr.RoleID)
	}
	if len(drs) > 0 {
		roles, err := r.rolesByIDs(ctx, map[uint][]uint{id: d.RoleIDs})
		if err != nil {
			return nil, err
		}
		for _, rid := range d.RoleIDs {
			if rl, ok := roles[rid]; ok {
				d.Roles = append(d.Roles, &domain.Role{ID: rl.ID, Name: rl.Name, Code: rl.Code})
			}
		}
	}
	return d, nil
}

func (r *DepartmentRepo) Create(ctx context.Context, d *domain.Department) error {
	po := departmentPO{
		ParentID: d.ParentID, Name: d.Name, Sort: d.Sort, Status: d.Status, Remark: d.Remark,
	}
	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return err
	}
	// 回填自增 ID 与时间戳，供上层/前端使用
	d.ID = po.ID
	d.CreatedAt = po.CreatedAt
	d.UpdatedAt = po.UpdatedAt
	return nil
}

func (r *DepartmentRepo) Update(ctx context.Context, id uint, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&departmentPO{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除部门，并在同一事务内级联清理 user_departments 与 department_roles 关联。
func (r *DepartmentRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&departmentPO{}).Error; err != nil {
			return err
		}
		if err := tx.Where("department_id = ?", id).Delete(&userDepartmentPO{}).Error; err != nil {
			return err
		}
		return tx.Where("department_id = ?", id).Delete(&departmentRolePO{}).Error
	})
}

func (r *DepartmentRepo) HasChildren(ctx context.Context, parentID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&departmentPO{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count > 0, err
}

// SetRoles 先删后插替换部门的角色关联（copy RoleRepo.AssignPermissions 模式）。
func (r *DepartmentRepo) SetRoles(ctx context.Context, departmentID uint, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("department_id = ?", departmentID).Delete(&departmentRolePO{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			if err := tx.Create(&departmentRolePO{DepartmentID: departmentID, RoleID: rid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// allDepartmentRoleIDs 返回 {departmentID: []roleID}。
func (r *DepartmentRepo) allDepartmentRoleIDs(ctx context.Context) (map[uint][]uint, error) {
	var drs []departmentRolePO
	if err := r.db.WithContext(ctx).Find(&drs).Error; err != nil {
		return nil, err
	}
	out := make(map[uint][]uint, len(drs))
	for _, dr := range drs {
		out[dr.DepartmentID] = append(out[dr.DepartmentID], dr.RoleID)
	}
	return out, nil
}

// rolesByIDs 返回 {roleID: rolePO}（按需查询，避免全表拉取）。
func (r *DepartmentRepo) rolesByIDs(ctx context.Context, ids map[uint][]uint) (map[uint]rolePO, error) {
	seen := make(map[uint]struct{})
	var all []uint
	for _, idList := range ids {
		for _, id := range idList {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				all = append(all, id)
			}
		}
	}
	if len(all) == 0 {
		return nil, nil
	}
	var pos []rolePO
	if err := r.db.WithContext(ctx).Where("id IN ?", all).Find(&pos).Error; err != nil {
		return nil, err
	}
	out := make(map[uint]rolePO, len(pos))
	for _, p := range pos {
		out[p.ID] = p
	}
	return out, nil
}
