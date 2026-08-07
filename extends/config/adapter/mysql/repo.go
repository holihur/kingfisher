package adapter

import (
	"context"
	"errors"

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
	return toConfigs(pos), nil
}

// GetPublicAll 返回全部公开配置（未登录可读）
func (r *ConfigRepo) GetPublicAll(ctx context.Context) ([]domain.SystemConfig, error) {
	var pos []configPO
	err := r.db.WithContext(ctx).Where("is_public = ?", true).Find(&pos).Error
	if err != nil {
		return nil, err
	}
	return toConfigs(pos), nil
}

func (r *ConfigRepo) GetByKey(ctx context.Context, key string) (*domain.SystemConfig, error) {
	var po configPO
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&po).Error
	if err != nil {
		return nil, err
	}
	return toConfig(&po), nil
}

// GetPublicByKey 返回指定 key 的公开配置；非公开项视为不存在
func (r *ConfigRepo) GetPublicByKey(ctx context.Context, key string) (*domain.SystemConfig, error) {
	var po configPO
	err := r.db.WithContext(ctx).Where("key = ? AND is_public = ?", key, true).First(&po).Error
	if err != nil {
		return nil, err
	}
	return toConfig(&po), nil
}

func (r *ConfigRepo) Set(ctx context.Context, key, value string, isPublic bool, version, render, renderOptions string) error {
	// 显式 upsert：存在则更新（map 更新保证 is_public=false 等零值写入），不存在则插入。
	// 不用 FirstOrCreate——它对 Assign + 空条件的行为不可靠（曾导致 key 为空 / 误更新其他记录）。
	var po configPO
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&po).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(&configPO{
			Key: key, Value: value, IsPublic: isPublic, Version: version, Render: render, RenderOptions: renderOptions,
		}).Error
	}
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&po).Updates(map[string]any{
		"value":          value,
		"is_public":      isPublic,
		"version":        version,
		"render":         render,
		"render_options": renderOptions,
	}).Error
}

func (r *ConfigRepo) Delete(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("key = ?", key).Delete(&configPO{}).Error
}

func toConfigs(pos []configPO) []domain.SystemConfig {
	configs := make([]domain.SystemConfig, len(pos))
	for i, p := range pos {
		configs[i] = *toConfig(&p)
	}
	return configs
}

func toConfig(p *configPO) *domain.SystemConfig {
	return &domain.SystemConfig{
		ID:            p.ID,
		Key:           p.Key,
		Value:         p.Value,
		Remark:        p.Remark,
		IsPublic:      p.IsPublic,
		Version:       p.Version,
		Render:        p.Render,
		RenderOptions: p.RenderOptions,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
