// Package jwt implements jwt logic.

package jwt

import (
	"context"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"kingfisher/core/cache"
	"kingfisher/core/config"
)

type Claims struct {
	UserID         uint     `json:"user_id"`
	RoleIDs        []uint   `json:"role_ids"`
	Roles          []string `json:"roles"`
	Username       string   `json:"username"`
	JTI            string   `json:"jti"`
	Type           string   `json:"type"` // access | refresh
	SessionVersion int      `json:"sv"`
	jwtlib.RegisteredClaims
}

type JWTManager struct {
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
	issuer     string
	cache      cache.Cache
}

func NewJWTManager(cfg config.JWTConfig, c cache.Cache) *JWTManager {
	return &JWTManager{
		secret:     cfg.Secret,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		issuer:     cfg.Issuer,
		cache:      c,
	}
}

func (m *JWTManager) GenerateToken(ctx context.Context, userID uint, roleIDs []uint, roles []string, username string, sessionVersion int) (string, string, error) {
	jti := uuid.New().String()
	now := time.Now()

	access := Claims{
		UserID:         userID,
		RoleIDs:        roleIDs,
		Roles:          roles,
		Username:       username,
		JTI:            jti,
		Type:           "access",
		SessionVersion: sessionVersion,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.accessTTL)),
			IssuedAt:  jwtlib.NewNumericDate(now),
			Issuer:    m.issuer,
		},
	}
	accessToken, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, access).SignedString([]byte(m.secret))
	if err != nil {
		return "", "", fmt.Errorf("sign access token: %w", err)
	}

	refresh := Claims{
		UserID:         userID,
		RoleIDs:        roleIDs,
		Roles:          roles,
		Username:       username,
		JTI:            jti,
		Type:           "refresh",
		SessionVersion: sessionVersion,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.refreshTTL)),
			IssuedAt:  jwtlib.NewNumericDate(now),
			Issuer:    m.issuer,
		},
	}
	refreshToken, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, refresh).SignedString([]byte(m.secret))
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// parseCore is the shared token parsing with algorithm + issuer validation.
func (m *JWTManager) parseCore(tokenStr string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenStr, &Claims{}, func(t *jwtlib.Token) (any, error) {
		return []byte(m.secret), nil
	},
		jwtlib.WithValidMethods([]string{"HS256"}),
		jwtlib.WithIssuer(m.issuer),
	)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// ParseToken parses and validates a JWT. For API auth, use ParseAccessToken.
func (m *JWTManager) ParseToken(ctx context.Context, tokenStr string) (*Claims, error) {
	claims, err := m.parseCore(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Type != "access" {
		return nil, fmt.Errorf("not an access token")
	}
	// Check revocation (fails closed on cache error)
	if m.cache != nil {
		revoked, err := m.cache.Exists(ctx, "blacklist:token:"+claims.JTI)
		if err != nil {
			return nil, fmt.Errorf("revocation check failed: %w", err)
		}
		if revoked {
			return nil, fmt.Errorf("token revoked")
		}
	}
	return claims, nil
}

// ParseRefreshToken parses a refresh token (type=refresh only).
func (m *JWTManager) ParseRefreshToken(ctx context.Context, refreshToken string) (*Claims, error) {
	claims, err := m.parseCore(refreshToken)
	if err != nil {
		return nil, err
	}
	if claims.Type != "refresh" {
		return nil, fmt.Errorf("not a refresh token")
	}
	// Check revocation
	if m.cache != nil {
		revoked, err := m.cache.Exists(ctx, "blacklist:token:"+claims.JTI)
		if err != nil {
			return nil, fmt.Errorf("revocation check failed: %w", err)
		}
		if revoked {
			return nil, fmt.Errorf("token revoked")
		}
	}
	return claims, nil
}

// GetSessionVersion returns the session version for a user. Injected per-request.
type SessionVersionProvider func(ctx context.Context, userID uint) (int, error)

func (m *JWTManager) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := m.ParseRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}
	jti := uuid.New().String()
	now := time.Now()
	access := Claims{
		UserID:         claims.UserID,
		RoleIDs:        claims.RoleIDs,
		Roles:          claims.Roles,
		JTI:            jti,
		Type:           "access",
		SessionVersion: claims.SessionVersion,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.accessTTL)),
			IssuedAt:  jwtlib.NewNumericDate(now),
			Issuer:    m.issuer,
		},
	}
	return jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, access).SignedString([]byte(m.secret))
}

func (m *JWTManager) RevokeToken(ctx context.Context, tokenStr string) error {
	claims, err := m.parseCore(tokenStr)
	if err != nil {
		return nil //nolint:nilerr // token already invalid
	}
	remainingTTL := time.Until(claims.ExpiresAt.Time)
	if remainingTTL <= 0 {
		return nil
	}
	if m.cache != nil {
		return m.cache.Set(ctx, "blacklist:token:"+claims.JTI, "1", remainingTTL)
	}
	return nil
}
