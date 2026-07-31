# JWT — 认证令牌

## 职责

JWT 生成、解析、刷新、黑名单。支持注销后 token 立即失效。

## 对外接口

```go
type JWTManager struct { /* ... */ }
func NewJWTManager(cfg JWTConfig, cache port.Cache) *JWTManager

func (m *JWTManager) GenerateToken(ctx context.Context, userID uint, role string, sessionVersion int) (access, refresh string, err error)
func (m *JWTManager) ParseToken(ctx context.Context, tokenStr string) (*Claims, error)
func (m *JWTManager) RefreshToken(ctx context.Context, refreshToken string) (string, error)
func (m *JWTManager) RevokeToken(ctx context.Context, tokenStr string) error     // 注销
func (m *JWTManager) IsRevoked(ctx context.Context, jti string) (bool, error)
```

## Claims 结构

```go
type Claims struct {
    UserID         uint   `json:"user_id"`
    Role           string `json:"role"`
    JTI            string `json:"jti"`           // 唯一 ID，用于黑名单
    Type           string `json:"type"`          // access | refresh
    SessionVersion int    `json:"sv"`            // 用户 session 版本号，用于强制踢出
    jwt.RegisteredClaims
}
```

## 核心逻辑

### GenerateToken

```
1. jti := uuid.New().String()
2. access Claims{UserID, Role, jti, "access", SessionVersion, expires=now+2h}
3. refresh Claims{UserID, Role, jti, "refresh", SessionVersion, expires=now+7d}
4. HS256 签名，返回两个 JWT 字符串
```

### ParseToken + 黑名单检查

```
1. jwt.ParseWithClaims(token, &Claims{}, keyFunc)
2. check expiry → 过期则 ErrTokenExpired
3. m.IsRevoked(ctx, claims.JTI) → 已注销则 ErrTokenInvalid
4. check claims.Type == "access" → 拒绝 refresh token 当 access token 用
5. 返回 *Claims
```

### RevokeToken（注销）

```
1. ParseToken(token) → 提取 JTI 和 exp
2. cache.Set("blacklist:token:"+jti, "1", ttl=剩余有效时间)
3. 如果提供了 refresh token，也一并黑名单
```

## Token 刷新逻辑

```
1. ParseToken(refreshToken)
2. 校验 Type=="refresh"
3. 校验未黑名单
4. 若 refresh token 剩余时间 < 24h → 同时更新 refresh token
5. 否则只生成新 access token
```

## 配置

```yaml
jwt:
  access_ttl: 2h
  refresh_ttl: 168h
  issuer: kingfisher
  # secret 来自环境变量 JWT_SECRET
```

## 设计要点

- 黑名单用 Redis，TTL 自动过期（过期后没必要存黑名单）
- JTI 是每个 token 的唯一 ID，即使相同 user 多次登录也不冲突
- `ParseToken` 不依赖 DB，只依赖 Redis 查黑名单 + 签名验证
