package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"kingfisher/core/cache"
	"kingfisher/extends/config/domain"
)

// cfgMockCache 内存实现 cache.Cache，测 ConfigService 缓存路径。
type cfgMockCache struct {
	store map[string]string
}

func (c *cfgMockCache) Get(_ context.Context, key string) (string, error) {
	if c.store == nil {
		return "", nil
	}
	return c.store[key], nil
}
func (c *cfgMockCache) Set(_ context.Context, key string, val any, _ time.Duration) error {
	if c.store == nil {
		c.store = map[string]string{}
	}
	c.store[key] = val.(string)
	return nil
}
func (c *cfgMockCache) Delete(_ context.Context, keys ...string) error {
	if c.store == nil {
		return nil
	}
	for _, k := range keys {
		delete(c.store, k)
	}
	return nil
}
func (c *cfgMockCache) DeleteByPattern(_ context.Context, _ string) error { return nil }
func (c *cfgMockCache) Exists(_ context.Context, _ string) (bool, error)  { return false, nil }
func (c *cfgMockCache) Incr(_ context.Context, _ string) (int64, error)   { return 1, nil }
func (c *cfgMockCache) Expire(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

// TestConfigGetAllPublicCache 校验公开配置走缓存读。
func TestConfigGetAllPublicCache(t *testing.T) {
	repo := &mockConfigRepo{configs: map[string]*domain.SystemConfig{}}
	cacheInst := &cfgMockCache{}
	// 预置一个过期缓存值，确认命中缓存而非查库
	_ = cacheInst.Set(context.Background(), "config:public", `[{"key":"cached"}]`, 0)
	svc := NewConfigService(repo, cacheInst)

	configs, err := svc.GetAllPublic(context.Background())
	if err != nil {
		t.Fatal("get all public:", err)
	}
	if len(configs) != 1 || configs[0].Key != "cached" {
		t.Errorf("expected cache hit, got %+v (repo has %d configs)", configs, len(repo.configs))
	}
}

func TestConfigGetAllPublicNoCache(t *testing.T) {
	repo := &mockConfigRepo{configs: map[string]*domain.SystemConfig{
		"a": {Key: "a", Value: "1", IsPublic: true},
		"b": {Key: "b", Value: "2", IsPublic: false},
	}}
	c := &cfgMockCache{}
	svc := NewConfigService(repo, c)
	configs, err := svc.GetAllPublic(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 1 {
		t.Fatalf("want only public, got %d", len(configs))
	}
	// 若已缓存则二次调用命中缓存
	_ = c.Set(context.Background(), "config:public", `[{"key":"a"}]`, 0)
	configs2, _ := svc.GetAllPublic(context.Background())
	if !strings.HasPrefix(configs2[0].Key, "a") {
		t.Errorf("unexpected cached value: %+v", configs2)
	}
}

func TestConfigGetCachedHit(t *testing.T) {
	repo := &mockConfigRepo{}
	c := &cfgMockCache{}
	_ = c.Set(context.Background(), "config:site_name", "缓存值", 0)
	svc := NewConfigService(repo, c)
	cfg, err := svc.Get(context.Background(), "site_name")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Value != "缓存值" {
		t.Errorf("want cached value, got %q", cfg.Value)
	}
}

func TestConfigSetInvalidatesCache(t *testing.T) {
	repo := &mockConfigRepo{configs: map[string]*domain.SystemConfig{}}
	c := &cfgMockCache{}
	_ = c.Set(context.Background(), "config:all", "x", 0)
	_ = c.Set(context.Background(), "config:site_name", "x", 0)
	svc := NewConfigService(repo, c)
	if err := svc.Set(context.Background(), "site_name", "值", true, "", "", "", 1); err != nil {
		t.Fatal(err)
	}
	// 写后应清空相关缓存
	if _, ok := c.store["config:site_name"]; ok {
		t.Error("config:site_name cache should be invalidated")
	}
}

var _ cache.Cache = (*cfgMockCache)(nil)
