# Security — 安全设计

## 职责

从网络层到应用层的安全防护，覆盖限流、超时、请求体限制、CORS、SQL 注入防护、XSS 防护。

## 防护清单

| 威胁 | 防护措施 | 实现位置 |
|------|----------|----------|
| 暴力破解 | 登录限流 5 次/min + 失败锁定 15min | RateLimit 中间件 + Redis |
| 垃圾注册 | 注册限流 1 次/5min per IP + 同邮箱 3 次/h | RateLimit 中间件 + Redis |
| DDoS | 全局限流 60 req/min per IP，`SetTrustedProxies` 防 IP 伪造 | RateLimit 中间件 + Gin Engine |
| SQL 注入 | GORM 参数化查询 + 禁止 raw SQL | repository 层规范 |
| XSS | JSON 响应 + nohtml validator + CSP header | Gin JSON + validation + SecurityHeaders |
| Clickjacking | `X-Frame-Options: DENY` | SecurityHeaders 中间件 |
| MIME 嗅探 | `X-Content-Type-Options: nosniff` | SecurityHeaders 中间件 |
| CSRF | JWT Bearer token（非 Cookie） + `SameSite=Lax`（如有 Cookie） | Auth 中间件 |
| 超大请求 | `c.Request.Body = http.MaxBytesReader(..., 10MB)` | Recovery 中间件 |
| 超时攻击 | Server ReadTimeout/WriteTimeout | http.Server 配置 |
| 敏感数据泄露 | 密码 json:"-", 日志脱敏 | model + logger |
| 越权 | RBAC 中间件 + 角色层级保护 | RBAC 中间件 |
| 权限提升 | 禁止低角色修改高角色、禁止修改自己角色 | role_service.go |
| Token 泄露 | 短 TTL + 黑名单 + 修改密码后自动踢出所有 session | JWT + ChangePassword |
| IP 伪造 | `SetTrustedProxies([]string{"127.0.0.1","10.0.0.0/8"})` + X-Forwarded-For | Gin Engine 初始化 |
| 用户枚举 | 登录统一返回 `code:10103`（不区分用户名不存在 vs 密码错误） | extends/user/app/service.go |

## RateLimit 设计

```go
// 全局限流
func RateLimit(cache port.Cache, limit int, window time.Duration) gin.HandlerFunc

// 登录专用限流（更严格）
func LoginRateLimit(cache port.Cache) gin.HandlerFunc
```

**滑动窗口算法**：
```
key = "ratelimit:{ip}:{path}"
1. now_ms := time.Now().UnixMilli()
2. window_start := now_ms - window_ms
3. Redis ZREMRANGEBYSCORE key 0 window_start    // 删除窗口外的记录
4. count := Redis ZCARD key
5. if count >= limit → 429
6. Redis ZADD key now_ms {now_ms}_{uuid}
7. Redis EXPIRE key window
```

## 请求体限制

```go
// 在 Recovery 中间件中
c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20) // 10MB
```

## CORS 白名单

```yaml
# 配置化，非 hardcode
cors:
  allowed_origins:
    - http://localhost:5173
    - https://admin.example.com
  allowed_methods: ["GET","POST","PUT","DELETE","OPTIONS"]
  allowed_headers: ["Authorization","Content-Type","X-Request-ID"]
  allow_credentials: true
  max_age: 12h
```

## 密码策略

```go
// Service 层
func validatePassword(password string) error {
    if len(password) < 8 { return ErrPasswordTooShort }
    if len(password) > 64 { return ErrPasswordTooLong }
    // 必须包含大小写+数字
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
    if !hasUpper || !hasLower || !hasDigit { return ErrPasswordWeak }
    return nil
}
// bcrypt cost = 12（约 250ms，平衡安全性和体验）
```

### 用户枚举防护

**Login 统一返回** `code:10103`，不区分用户名不存在和密码错误。

```go
// extends/user/app/service.go
func (s *UserService) Login(ctx context.Context, username, password string) (...) {
    user, err := s.repo.FindByUsername(ctx, username)
    // 用固定假 hash 执行 bcrypt 比对，即使用户不存在
    dummyHash := "$2a$12$..." // 固定值
    hashToCheck := dummyHash
    if err == nil {
        hashToCheck = user.Password
    }
    if bcrypt.CompareHashAndPassword([]byte(hashToCheck), []byte(password)) != nil {
        // 统一返回密码错误——不透露用户是否存在
        return "", "", nil, domain.ErrPasswordWrong
    }
    if err != nil { // 比对通过但用户其实不存在（概率极低）
        return "", "", nil, domain.ErrPasswordWrong
    }
    // ... 正常登录流程
}
```

- `ErrUserNotFound` (10102) 只出现在用户主动查询接口（`GET /users/:id`）
- 登录接口绝不返回 10102

### 角色层级约束

`roles` 表加 `level` 字段，防权限提升：

```yaml
种子数据:
  admin  → level: 0   # 不可被任何人删除/降权
  editor → level: 1
  viewer → level: 2
```

```go
// extends/rbac/app/role_service.go
func (s *RoleService) Update(ctx context.Context, callerUserID uint, targetRoleID uint, updates map[string]any) error {
    caller := s.getUser(ctx, callerUserID)
    targetRole := s.getRole(ctx, targetRoleID)

    if caller.Role.Level >= targetRole.Level && caller.Role.Code != "admin" {
        return domain.ErrForbidden  // 低 level 不能改高 level
    }
    if newLevel, ok := updates["level"]; ok && newLevel.(int) <= caller.Role.Level {
        return domain.ErrForbidden  // 不能把自己或别人提到比自己更高的 level
    }
    return s.repo.Update(ctx, targetRoleID, updates)
}
```

### 注册限流

`POST /api/v1/auth/register` — 2 次/5min per IP：

```go
auth.POST("/register", middleware.RateLimit(m.cache, 2, 5*time.Minute), m.handler.Register)
```

### 修改密码自动踢出

`ChangePassword` 成功后自动 `increment session_version`，旧 token 全部失效：

```go
func (s *UserService) ChangePassword(ctx context.Context, userID uint, oldPwd, newPwd string) error {
    // ... 校验 + 更新密码 ...
    s.repo.IncrementSessionVersion(userID)         // 踢出所有旧 session
    s.cache.Delete(ctx, "user:sv:"+strconv.Itoa(int(userID)))
    return nil
}
```

### Config Value 类型校验

带类型标记的配置项在后端校验值类型：

```go
// extends/config/app/service.go
var typedConfigs = map[string]string{
    "max_login_attempts": "int",
    "lockout_duration":   "duration",
    "session_timeout":    "duration",
}

func (s *ConfigService) Set(ctx context.Context, key, value string) error {
    if expectedType, ok := typedConfigs[key]; ok {
        if !validateType(value, expectedType) {
            return domain.ErrInvalidParam
        }
    }
    return s.repo.Set(ctx, key, value)
}
```

### TrustedProxies

```yaml
# config/config.yaml
server:
  trusted_proxies:
    - "127.0.0.1"       # docker-compose 中的 nginx
    - "10.0.0.0/8"      # K8s 内部网络
```

```go
r.SetTrustedProxies(cfg.Server.TrustedProxies)
```

### Refresh Token Rotation 决策

v1 **不做自动 rotation**。权衡：
- 对管理后台而言，refresh token 泄露场景概率低（管理员设备而非公共 kiosk）
- Rotation 需要额外的 refresh token family 管理（父 token 失效时全部子孙失效）
- 定期改密码自动踢出所有有效 token 是更简单的替代方案

## 设计要点

- 限流计数器存在 Redis，多实例共享
- RBAC 中间件从 DB 查权限？（否）→ 从 JWT claims 取角色，从 Redis 缓存取权限列表
- 所有外部输入（query、path param、body）经过 Gin binding/validation
