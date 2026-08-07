package adapter

import (
	"context"
	"time"

	"gorm.io/gorm"

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
	err := r.db.WithContext(ctx).First(&po, id).Error
	if err != nil {
		return nil, err
	}
	return po.toDomain(), nil
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var po userPO
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&po).Error
	if err != nil {
		return nil, err
	}
	return po.toDomain(), nil
}

func (r *UserRepo) FindAll(ctx context.Context, q *query.Query) ([]domain.User, int64, error) {
	var pos []userPO
	total, err := q.Find(r.db.WithContext(ctx).Model(&userPO{}), &pos)
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
	po := userPO{Username: u.Username, Nickname: u.Nickname, Password: u.Password, Email: u.Email, Status: u.Status, RoleID: u.RoleID}
	err := r.db.WithContext(ctx).Create(&po).Error
	if err == nil {
		u.ID = po.ID
		u.CreatedAt = po.CreatedAt
		u.UpdatedAt = po.UpdatedAt
	}
	return err
}

func (r *UserRepo) Update(ctx context.Context, id uint, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&userPO{}).Where("id = ?", id).Updates(updates).Error
}

func (r *UserRepo) Delete(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&userPO{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

func (r *UserRepo) IncrementSessionVersion(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&userPO{}).Where("id = ?", id).
		Update("session_version", gorm.Expr("session_version + 1")).Error
}

func (r *UserRepo) GetSessionVersion(ctx context.Context, id uint) (int, error) {
	var sv int
	err := r.db.WithContext(ctx).Model(&userPO{}).Select("session_version").Where("id = ?", id).Scan(&sv).Error
	return sv, err
}
