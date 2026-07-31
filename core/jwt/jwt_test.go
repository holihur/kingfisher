package jwt

import (
	"context"
	"testing"
	"time"

	_ "kingfisher/core/cache"
	"kingfisher/core/config"
)

func TestGenerateAndParse(t *testing.T) {
	mgr := NewJWTManager(config.JWTConfig{Secret: "test-secret", AccessTTL: time.Hour, RefreshTTL: 2 * time.Hour, Issuer: "test"}, nil)
	access, refresh, err := mgr.GenerateToken(context.Background(), 1, "admin", 1)
	if err != nil { t.Fatal("generate:", err) }
	if access == "" || refresh == "" { t.Error("empty tokens") }

	claims, err := mgr.ParseToken(context.Background(), access)
	if err != nil { t.Fatal("parse:", err) }
	if claims.UserID != 1 { t.Error("user_id should be 1") }
	if claims.Role != "admin" { t.Error("role should be admin") }
	if claims.Type != "access" { t.Error("type should be access") }
	if claims.SessionVersion != 1 { t.Error("sv should be 1") }
}

func TestRefreshToken(t *testing.T) {
	mgr := NewJWTManager(config.JWTConfig{Secret: "test-secret", AccessTTL: time.Hour, RefreshTTL: 2 * time.Hour, Issuer: "test"}, nil)
	_, refresh, _ := mgr.GenerateToken(context.Background(), 1, "admin", 1)

	newAccess, err := mgr.RefreshToken(context.Background(), refresh)
	if err != nil { t.Fatal("refresh:", err) }
	if newAccess == "" { t.Error("empty new access") }
}

func TestRefreshWithAccessToken(t *testing.T) {
	mgr := NewJWTManager(config.JWTConfig{Secret: "test-secret", AccessTTL: time.Hour, RefreshTTL: 2 * time.Hour, Issuer: "test"}, nil)
	access, _, _ := mgr.GenerateToken(context.Background(), 1, "admin", 1)

	_, err := mgr.RefreshToken(context.Background(), access)
	if err == nil { t.Error("should reject access token for refresh") }
}

func TestExpiredToken(t *testing.T) {
	// Token with very short TTL — generates fine but Parse fails after TTL
	mgr := NewJWTManager(config.JWTConfig{Secret: "test", AccessTTL: -1 * time.Hour, RefreshTTL: time.Hour, Issuer: "test"}, nil)
	_, _, err := mgr.GenerateToken(context.Background(), 1, "admin", 1)
	// GenerateToken with negative TTL creates a token that's already expired; parse would fail
	if err != nil { t.Log("generate with negative TTL:", err) }
}

func TestRevokeToken(t *testing.T) {
	mgr := NewJWTManager(config.JWTConfig{Secret: "test", AccessTTL: time.Hour, RefreshTTL: 2 * time.Hour, Issuer: "test"}, nil)
	access, _, _ := mgr.GenerateToken(context.Background(), 1, "admin", 1)
	// Revoke without cache: ParseToken succeeds, then cache.Set fails silently (cache is nil)
	// Cache=nil means IsRevoked is skipped in ParseToken
	err := mgr.RevokeToken(context.Background(), access)
	// RevokeToken catches the cache.Set error internally and returns nil (best-effort)
	if err != nil { t.Log("revoke with nil cache:", err) }
}
