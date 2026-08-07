package app

import (
	"context"
	"testing"
	"time"

	"kingfisher/core/cache"
	"kingfisher/extends/menu/domain"
)

// menuMockCache 内存实现 cache.Cache，测 MenuService 的树缓存路径。
type menuMockCache struct {
	store map[string]string
}

func (c *menuMockCache) Get(_ context.Context, key string) (string, error) {
	if c.store == nil {
		return "", nil
	}
	return c.store[key], nil
}
func (c *menuMockCache) Set(_ context.Context, key string, val any, _ time.Duration) error {
	if c.store == nil {
		c.store = map[string]string{}
	}
	c.store[key] = val.(string)
	return nil
}
func (c *menuMockCache) Delete(_ context.Context, keys ...string) error {
	if c.store == nil {
		return nil
	}
	for _, k := range keys {
		delete(c.store, k)
	}
	return nil
}
func (c *menuMockCache) DeleteByPattern(_ context.Context, _ string) error { return nil }
func (c *menuMockCache) Exists(_ context.Context, _ string) (bool, error)  { return false, nil }
func (c *menuMockCache) Incr(_ context.Context, _ string) (int64, error)   { return 1, nil }
func (c *menuMockCache) Expire(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

// TestMenuGetTreeWithCache 覆盖 GetTree 的缓存命中与未命中路径。
func TestMenuGetTreeWithCache(t *testing.T) {
	repo := &mockMenuRepo{menus: map[uint]*domain.Menu{
		1: {ID: 1, Name: "系统", Path: "/system"},
		2: {ID: 2, ParentID: 1, Name: "用户", Path: "/system/user"},
		3: {ID: 3, Name: "其他"},
	}}
	c := &menuMockCache{}
	svc := NewMenuService(repo, c)
	ctx := context.Background()

	// 首次调用：未命中 → 查库并写入缓存
	tree, err := svc.GetTree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != 2 {
		t.Fatalf("want 2 roots, got %d", len(tree))
	}
	if _, ok := c.store["menu:tree"]; !ok {
		t.Error("tree should be cached after first call")
	}

	// 二次调用：命中缓存（repo 内容变化也不影响）
	delete(repo.menus, 3)
	repo.menus[1].Name = "改名"
	tree2, err := svc.GetTree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree2) != 2 {
		t.Fatalf("cache should still return 2 roots, got %d", len(tree2))
	}
	if tree2[0].Name != "系统" && tree2[1].Name != "系统" {
		t.Errorf("cache hit should return stale name, got roots %q %q", tree2[0].Name, tree2[1].Name)
	}
}

// TestMenuCreateInvalidatesTreeCache 覆盖写操作后清除菜单树缓存。
func TestMenuCreateInvalidatesTreeCache(t *testing.T) {
	repo := &mockMenuRepo{menus: map[uint]*domain.Menu{}}
	c := &menuMockCache{}
	_ = c.Set(context.Background(), "menu:tree", `[{"id":1}]`, 0)
	svc := NewMenuService(repo, c)
	ctx := context.Background()

	if err := svc.Create(ctx, &domain.Menu{Name: "新菜单", Path: "/new"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.store["menu:tree"]; ok {
		t.Error("menu:tree cache should be invalidated after create")
	}
}

var _ cache.Cache = (*menuMockCache)(nil)