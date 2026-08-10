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
	err := r.db.WithContext(ctx).Preload("Roles").First(&po, id).Error
	if err != nil {
		return nil, err
	}
	return po.toDomain(), nil
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var po userPO
	err := r.db.WithContext(ctx).Preload("Roles").Where("username = ?", username).First(&po).Error
	if err != nil {
		return nil, err
	}
	return po.toDomain(), nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var po userPO
	err := r.db.WithContext(ctx).Preload("Roles").Where("email = ?", email).First(&po).Error
	if err != nil {
		return nil, err
	}
	return po.toDomain(), nil
}

func (r *UserRepo) FindAll(ctx context.Context, q *query.Query) ([]domain.User, int64, error) {
	var pos []userPO
	base := r.db.WithContext(ctx).Model(&userPO{}).Preload("Roles")

	// role_id 筛选 = 用户「拥有该角色」（user_roles 成员过滤），由 user_roles 关联表实现。
	// eq/in → 拥有任一；ne → 不拥有该角色。
	var roleVals []any
	var notRoleVals []any
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
		default:
			rest = append(rest, f)
		}
	}
	if len(roleVals) > 0 {
		base = base.Where("EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = users.id AND ur.role_id IN ?)", roleVals)
	}
	if len(notRoleVals) > 0 {
		base = base.Where("NOT EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = users.id AND ur.role_id IN ?)", notRoleVals)
	}
	// 拷贝 Query，避免污染调用方；role_id 条件已由上方 EXISTS 处理，不再交给 q.Find
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
	return users, total, nil
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
		return r.setUserRoles(tx, po.ID, u.RoleIDs)
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

func (r *UserRepo) Update(ctx context.Context, id uint, updates map[string]any) error {
	// 角色变更单独处理：普通字段与 user_roles 关联在同一事务内更新
	if roleIDs, ok := updates["role_ids"].([]uint); ok {
		delete(updates, "role_ids")
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&userPO{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
			return r.setUserRoles(tx, id, roleIDs)
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
