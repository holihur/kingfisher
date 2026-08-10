package jwt

import (
	"context"
	"errors"
	"testing"
	"time"

	"kingfisher/core/cache"
	"kingfisher/core/config"
)

// mockCache implements cache.Cache for jwt revocation-path tests.
type mockCache struct {
	blacklist  map[string]bool
	failExists bool
}

func (m *mockCache) Get(context.Context, string) (string, error)                   { return "", nil }
func (m *mockCache) Set(_ context.Context, _ string, _ any, _ time.Duration) error { return nil }
func (m *mockCache) Delete(context.Context, ...string) error                       { return nil }
func (m *mockCache) DeleteByPattern(context.Context, string) error                 { return nil }
func (m *mockCache) Exists(_ context.Context, key string) (bool, error) {
	if m.failExists {
		return false, errors.New("redis down")
	}
	return m.blacklist[key], nil
}
func (m *mockCache) Incr(context.Context, string) (int64, error)         { return 1, nil }
func (m *mockCache) Expire(context.Context, string, time.Duration) error { return nil }

func TestParseTokenRevoked(t *testing.T) {
	mgr := NewJWTManager(config.JWTConfig{Secret: "test-secret", AccessTTL: time.Hour, RefreshTTL: 2 * time.Hour, Issuer: "test"}, &mockCache{blacklist: map[string]bool{}})
	access, _, err := mgr.GenerateToken(context.Background(), 1, []uint{1}, []string{"admin"}, "admin", 1)
	if err != nil {
		t.Fatal(err)
	}
	// 正常可解析
	if _, err := mgr.ParseToken(context.Background(), access); err != nil {
		t.Fatal("shold parse:", err)
	}
	// 加入黑名单后禁止
	claims, _ := mgr.ParseToken(context.Background(), access)
	mgr.cache.(*mockCache).blacklist["blacklist:token:"+claims.JTI] = true
	if _, err := mgr.ParseToken(context.Background(), access); err == nil {
		t.Error("revoked token should fail")
	}
}

func TestParseTokenCacheError(t *testing.T) {
	mgr := NewJWTManager(config.JWTConfig{Secret: "test-secret", AccessTTL: time.Hour, RefreshTTL: 2 * time.Hour, Issuer: "test"}, &mockCache{failExists: true})
	access, _, _ := mgr.GenerateToken(context.Background(), 1, []uint{1}, []string{"admin"}, "admin", 1)
	// 缓存错误时 fails-closed
	if _, err := mgr.ParseToken(context.Background(), access); err == nil {
		t.Error("cache error should fail closed")
	}
}

func TestParseRefreshTokenRevoked(t *testing.T) {
	mgr := NewJWTManager(config.JWTConfig{Secret: "test-secret", AccessTTL: time.Hour, RefreshTTL: 2 * time.Hour, Issuer: "test"}, &mockCache{blacklist: map[string]bool{}})
	_, refresh, _ := mgr.GenerateToken(context.Background(), 1, []uint{1}, []string{"admin"}, "admin", 1)
	claims, err := mgr.ParseRefreshToken(context.Background(), refresh)
	if err != nil || claims == nil {
		t.Fatalf("should parse refresh: %v", err)
	}
	mgr.cache.(*mockCache).blacklist["blacklist:token:"+claims.JTI] = true
	if _, err := mgr.ParseRefreshToken(context.Background(), refresh); err == nil {
		t.Error("revoked refresh should fail")
	}
}

var _ cache.Cache = (*mockCache)(nil)
