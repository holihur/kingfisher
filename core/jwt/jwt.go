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
	UserID         uint   `json:"user_id"`
	Role           string `json:"role"`
	JTI            string `json:"jti"`
	Type           string `json:"type"` // access | refresh
	SessionVersion int    `json:"sv"`
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

func (m *JWTManager) GenerateToken(ctx context.Context, userID uint, role string, sessionVersion int) (string, string, error) {
	jti := uuid.New().String()
	now := time.Now()

	access := Claims{
		UserID:         userID,
		Role:           role,
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
		Role:           role,
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

func (m *JWTManager) ParseToken(ctx context.Context, tokenStr string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenStr, &Claims{}, func(t *jwtlib.Token) (any, error) {
		return []byte(m.secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	// Check revocation
	if m.cache != nil {
		revoked, _ := m.cache.Exists(ctx, "blacklist:token:"+claims.JTI)
		if revoked {
			return nil, fmt.Errorf("token revoked")
		}
	}
	return claims, nil
}

func (m *JWTManager) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := m.ParseToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}
	if claims.Type != "refresh" {
		return "", fmt.Errorf("not a refresh token")
	}
	jti := uuid.New().String()
	now := time.Now()
	access := Claims{
		UserID:         claims.UserID,
		Role:           claims.Role,
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
	claims, err := m.ParseToken(ctx, tokenStr)
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
