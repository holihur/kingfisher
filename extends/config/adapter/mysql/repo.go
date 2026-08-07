package adapter

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"kingfisher/core/query"
	"kingfisher/extends/config/domain"
)

type ConfigRepo struct{ db *gorm.DB }

func NewConfigRepo(db *gorm.DB) *ConfigRepo { return &ConfigRepo{db: db} }

// List 分页 + 结构化查询配置列表
func (r *ConfigRepo) List(ctx context.Context, q *query.Query) ([]domain.SystemConfig, int64, error) {
	var pos []configPO
	total, err := q.Find(r.db.WithContext(ctx).Model(&configPO{}), &pos)
	if err != nil {
		return nil, 0, err
	}
	return toConfigs(ctx, r.db, pos), total, nil
}

// GetPublicAll 返回全部公开配置（未登录可读）
func (r *ConfigRepo) GetPublicAll(ctx context.Context) ([]domain.SystemConfig, error) {
	var pos []configPO
	err := r.db.WithContext(ctx).Where("is_public = ?", true).Find(&pos).Error
	if err != nil {
		return nil, err
	}
	return toConfigs(ctx, r.db, pos), nil
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

func (r *ConfigRepo) Set(ctx context.Context, key, value string, isPublic bool, version, render, renderOptions string, groupID uint) error {
	// 显式 upsert：存在则更新（map 更新保证 is_public=false 等零值写入），不存在则插入。
	// 不用 FirstOrCreate——它对 Assign + 空条件的行为不可靠（曾导致 key 为空 / 误更新其他记录）。
	var po configPO
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&po).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(&configPO{
			Key: key, Value: value, IsPublic: isPublic, Version: version, Render: render, RenderOptions: renderOptions, GroupID: groupID,
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
		"group_id":       groupID,
	}).Error
}

func (r *ConfigRepo) Delete(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("key = ?", key).Delete(&configPO{}).Error
}

// DeleteBatch 批量删除配置（按 key）
func (r *ConfigRepo) DeleteBatch(ctx context.Context, keys []string) error {
	return r.db.WithContext(ctx).Where("key IN ?", keys).Delete(&configPO{}).Error
}

// ConfigGroupRepo 配置分组 CRUD
type ConfigGroupRepo struct{ db *gorm.DB }

func NewConfigGroupRepo(db *gorm.DB) *ConfigGroupRepo { return &ConfigGroupRepo{db: db} }

func (r *ConfigGroupRepo) List(ctx context.Context) ([]domain.ConfigGroup, error) {
	var pos []configGroupPO
	err := r.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&pos).Error
	if err != nil {
		return nil, err
	}
	groups := make([]domain.ConfigGroup, len(pos))
	for i, p := range pos {
		groups[i] = domain.ConfigGroup{ID: p.ID, Name: p.Name, Sort: p.Sort, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
	}
	return groups, nil
}

func (r *ConfigGroupRepo) Create(ctx context.Context, name string, sort int) (*domain.ConfigGroup, error) {
	po := configGroupPO{Name: name, Sort: sort}
	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return nil, err
	}
	return &domain.ConfigGroup{ID: po.ID, Name: po.Name, Sort: po.Sort, CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt}, nil
}

func (r *ConfigGroupRepo) Update(ctx context.Context, id uint, name string, sort int) error {
	return r.db.WithContext(ctx).Model(&configGroupPO{}).Where("id = ?", id).Updates(map[string]any{
		"name": name,
		"sort": sort,
	}).Error
}

func (r *ConfigGroupRepo) Delete(ctx context.Context, id uint) error {
	// 删除分组时，将其下配置移回未分组（group_id=0）
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&configPO{}).Where("group_id = ?", id).Update("group_id", 0).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&configGroupPO{}).Error
	})
}

// toConfigs 批量转换并填充 group_name（一次查全量分组，避免 N+1）
func toConfigs(ctx context.Context, db *gorm.DB, pos []configPO) []domain.SystemConfig {
	configs := make([]domain.SystemConfig, len(pos))
	for i, p := range pos {
		configs[i] = *toConfig(&p)
	}
	// 收集用到的 group_id 并批量查名称
	ids := map[uint]bool{}
	for _, p := range pos {
		if p.GroupID > 0 {
			ids[p.GroupID] = true
		}
	}
	if len(ids) > 0 {
		var gpos []configGroupPO
		if err := db.WithContext(ctx).Select("id", "name").Where("id IN ?", keys(ids)).Find(&gpos).Error; err == nil {
			names := map[uint]string{}
			for _, g := range gpos {
				names[g.ID] = g.Name
			}
			for i := range configs {
				configs[i].GroupName = names[configs[i].GroupID]
			}
		}
	}
	return configs
}

func keys(m map[uint]bool) []uint {
	out := make([]uint, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// toConfig 转换单个配置（不含 group_name；列表场景用 toConfigs 批量填充）
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
		GroupID:       p.GroupID,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
