package app

import (
	"context"
	"testing"
	"time"

	"kingfisher/core/cache"
	"kingfisher/extends/rbac/domain"
)

// roleMockCache 内存实现 cache.Cache，覆盖 RBAC 服务的缓存分支。
type roleMockCache struct {
	store map[string]string
}

func (c *roleMockCache) Get(_ context.Context, key string) (string, error) {
	if c.store == nil {
		return "", nil
	}
	return c.store[key], nil
}
func (c *roleMockCache) Set(_ context.Context, key string, val any, _ time.Duration) error {
	if c.store == nil {
		c.store = map[string]string{}
	}
	c.store[key] = val.(string)
	return nil
}
func (c *roleMockCache) Delete(_ context.Context, keys ...string) error {
	if c.store == nil {
		return nil
	}
	for _, k := range keys {
		delete(c.store, k)
	}
	return nil
}
func (c *roleMockCache) DeleteByPattern(_ context.Context, _ string) error {
	if c.store == nil {
		return nil
	}
	c.store = map[string]string{}
	return nil
}
func (c *roleMockCache) Exists(_ context.Context, _ string) (bool, error) { return false, nil }
func (c *roleMockCache) Incr(_ context.Context, _ string) (int64, error)  { return 1, nil }
func (c *roleMockCache) Expire(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

var _ cache.Cache = (*roleMockCache)(nil)

// TestAssignPermissionsWithCache 覆盖写操作后清理 user:perms 缓存。
func TestAssignPermissionsWithCache(t *testing.T) {
	c := &roleMockCache{store: map[string]string{"user:perms:9": "user:list,user:get"}}
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{1: {ID: 1, Name: "op1", Code: "op1", Status: 1}},
	}
	svc := NewRoleService(repo, c)
	ctx := context.Background()
	if err := svc.AssignPermissions(ctx, 1, []uint{1}); err != nil {
		t.Fatal(err)
	}
	if len(c.store) != 0 {
		t.Error("AssignPermissions should clear user:perms:* caches")
	}
}

// TestAssignMenusWithCache 验证写操作后菜单缓存被清除。
func TestAssignMenusWithCache(t *testing.T) {
	c := &roleMockCache{store: map[string]string{"menu:role:1": "[]", "menu:tree": "[]"}}
	repo := &mockRoleRepo{
		roles: map[uint]*domain.Role{1: {ID: 1, Name: "op1", Code: "op1", Status: 1}},
	}
	svc := NewRoleService(repo, c)
	ctx := context.Background()
	if err := svc.AssignMenus(ctx, 1, []uint{1}); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.store["menu:role:1"]; ok {
		t.Error("menu:role:1 should be cleared")
	}
	if _, ok := c.store["menu:tree"]; ok {
		t.Error("menu:tree should be cleared")
	}
}

// TestGetUserPermissionsCache 覆盖缓存命中路径（mock repo 返回 user:list,user:create）。
func TestGetUserPermissionsCache(t *testing.T) {
	repo := &mockRoleRepo{}
	c := &roleMockCache{store: map[string]string{"user:perms:1": "role:list,role:get"}}
	svc := NewRoleService(repo, c)
	prms, err := svc.GetUserPermissions(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(prms) != 2 {
		t.Errorf("want 2 cached perms, got %d", len(prms))
	}
	if prms[0] != "role:list" {
		t.Errorf("cache hit should return cached perms, got %v", prms)
	}
}

// TestGetUserPermissionsCacheMiss 覆盖缓存未命中回源并写入缓存。
func TestGetUserPermissionsCacheMiss(t *testing.T) {
	repo := &mockRoleRepo{}
	c := &roleMockCache{}
	svc := NewRoleService(repo, c)
	prms, err := svc.GetUserPermissions(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(prms) != 2 {
		t.Fatalf("want 2, got %d", len(prms))
	}
	if got := c.store["user:perms:1"]; got != "user:list,user:create" {
		t.Errorf("cache not written after miss: got %q", got)
	}
}

// TestStrSliceEmpty 覆盖逗号拆分空串返回 nil 分支。
func TestStrSliceEmpty(t *testing.T) {
	if out := strSlice(""); out != nil {
		t.Errorf("empty strSlice should be nil, got %v", out)
	}
}