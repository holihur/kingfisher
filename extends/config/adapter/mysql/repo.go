package adapter

import (
	"context"

	"gorm.io/gorm"

	"kingfisher/extends/config/domain"
)

type ConfigRepo struct{ db *gorm.DB }

func NewConfigRepo(db *gorm.DB) *ConfigRepo { return &ConfigRepo{db: db} }
func (r *ConfigRepo) GetAll(ctx context.Context) ([]domain.SystemConfig, error) {
	var pos []configPO
	err := r.db.WithContext(ctx).Find(&pos).Error
	if err != nil {
		return nil, err
	}
	configs := make([]domain.SystemConfig, len(pos))
	for i, p := range pos {
		configs[i] = domain.SystemConfig{ID: p.ID, Key: p.Key, Value: p.Value, Remark: p.Remark, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
	}
	return configs, nil
}
func (r *ConfigRepo) GetByKey(ctx context.Context, key string) (*domain.SystemConfig, error) {
	var po configPO
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&po).Error
	if err != nil {
		return nil, err
	}
	return &domain.SystemConfig{ID: po.ID, Key: po.Key, Value: po.Value, Remark: po.Remark, CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt}, nil
}
func (r *ConfigRepo) Set(ctx context.Context, key, value string) error {
	return r.db.WithContext(ctx).Where("key = ?", key).Assign(configPO{Value: value}).FirstOrCreate(&configPO{}).Error
}
func (r *ConfigRepo) Delete(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("key = ?", key).Delete(&configPO{}).Error
}
