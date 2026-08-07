package app

import (
	"context"
	"encoding/json"
	"time"

	coreCache "kingfisher/core/cache"
	"kingfisher/extends/config/domain"
	"kingfisher/extends/config/port"
)

type ConfigService struct {
	repo  port.ConfigRepository
	cache coreCache.Cache
}

func NewConfigService(repo port.ConfigRepository, cache coreCache.Cache) *ConfigService {
	return &ConfigService{repo: repo, cache: cache}
}

func (s *ConfigService) GetAll(ctx context.Context) ([]domain.SystemConfig, error) {
	if s.cache != nil {
		if val, err := s.cache.Get(ctx, "config:all"); err == nil && val != "" {
			var configs []domain.SystemConfig
			if json.Unmarshal([]byte(val), &configs) == nil {
				return configs, nil
			}
		}
	}
	configs, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		if data, err := json.Marshal(configs); err == nil {
			_ = s.cache.Set(ctx, "config:all", string(data), 5*time.Minute)
		}
	}
	return configs, nil
}

// GetAllPublic 返回全部公开配置（未登录可读）
func (s *ConfigService) GetAllPublic(ctx context.Context) ([]domain.SystemConfig, error) {
	if s.cache != nil {
		if val, err := s.cache.Get(ctx, "config:public"); err == nil && val != "" {
			var configs []domain.SystemConfig
			if json.Unmarshal([]byte(val), &configs) == nil {
				return configs, nil
			}
		}
	}
	configs, err := s.repo.GetPublicAll(ctx)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		if data, err := json.Marshal(configs); err == nil {
			_ = s.cache.Set(ctx, "config:public", string(data), 5*time.Minute)
		}
	}
	return configs, nil
}

func (s *ConfigService) Get(ctx context.Context, key string) (*domain.SystemConfig, error) {
	if s.cache != nil {
		if val, err := s.cache.Get(ctx, "config:"+key); err == nil && val != "" {
			return &domain.SystemConfig{Key: key, Value: val}, nil
		}
	}
	return s.repo.GetByKey(ctx, key)
}

// GetPublic 返回指定 key 的公开配置；非公开项视为不存在
func (s *ConfigService) GetPublic(ctx context.Context, key string) (*domain.SystemConfig, error) {
	return s.repo.GetPublicByKey(ctx, key)
}

func (s *ConfigService) Set(ctx context.Context, key, value string, isPublic bool, version, render, renderOptions string) error {
	if err := s.repo.Set(ctx, key, value, isPublic, version, render, renderOptions); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, "config:all")
		_ = s.cache.Delete(ctx, "config:public")
		_ = s.cache.Delete(ctx, "config:"+key)
	}
	return nil
}

func (s *ConfigService) Delete(ctx context.Context, key string) error {
	if err := s.repo.Delete(ctx, key); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, "config:all")
	}
	return nil
}
