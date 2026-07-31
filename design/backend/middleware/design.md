# Middleware — 中间件链

## 职责

提供通用 HTTP 中间件。中间件分两层：**Core 中间件**（框架级）和 **Extends 中间件**（业务级）。

## Core 中间件

| 中间件 | 文件 | 职责 |
|--------|------|------|
| RequestID | `core/middleware/request_id.go` | 透传或生成 X-Request-ID |
| Recovery | `core/middleware/recovery.go` | 捕获 panic → log + stack + 500 |
| Trace | `core/middleware/trace.go` | 创建 OTel Span，注入 trace_id |
| Logger | `core/middleware/logger.go` | method/path/status/latency |
| Gzip | `core/middleware/gzip.go` | JSON 响应压缩（> 1KB 才压缩） |
| CORS | `core/middleware/cors.go` | 跨域白名单 |
| SecurityHeaders | `core/middleware/security_headers.go` | CSP/HSTS/X-Frame-Options/X-Content-Type-Options/Referrer-Policy |
| RateLimit | `core/middleware/ratelimit.go` | 滑动窗口限流 |

## Extends 中间件（来自 extends/rbac）

| 中间件 | 文件 | 职责 |
|--------|------|------|
| Auth | `extends/rbac/transport/auth_middleware.go` | JWT 解析 → user_id/role 注入 ctx |
| RBAC | `extends/rbac/transport/rbac_middleware.go` | 从缓存加载用户权限列表注入 ctx |
| RequirePerm | `extends/rbac/transport/require_perm.go` | 检查 ctx 中是否含指定权限 code |

> **单一职责**：Auth 只管认证（你是谁），RBAC 只管授权（你能做什么），RequirePerm 只管粒度校验（你能不能做这件事）。

## 中间件链顺序

```
RequestID → Recovery → Trace → Logger → Gzip → SecurityHeaders → CORS → Auth(可选) → RBAC(可选) → RequirePerm(可选) → RateLimit(可选) → Handler
```

## 为什么要这个顺序

Gin 中间件按注册顺序执行，`c.Next()` 前为进栈、后为出栈（洋葱模型）。Recovery 必须是最外层才能捕获所有下游的 panic。

| 顺序 | 中间件 | 位置原因 |
|------|--------|----------|
| 1 | RequestID | 最外层——即使 Recovery 触发，panic 日志也能带上 request_id |
| 2 | Recovery | 第二层——defer recover() 包裹下面一切，任意位置 panic 都能捕获 |
| 3 | Trace | Recovery 之后——即使 handler panic，defer span.End() 仍执行，trace 不丢 |
| 4 | Logger | Trace 之后——日志可带 trace_id。放在 Recovery 里面，panic 的 500 状态码也会记录 |
| 5 | Gzip | Logger 之后、SecurityHeaders 之前——压缩 JSON 响应体（>1KB），排除 /metrics /health |
| 6 | SecurityHeaders | CORS 之前——所有响应（含 OPTIONS 和 401）都带安全头 |
| 7 | CORS | Auth 之前——OPTIONS 预检请求不需要认证 |
| 8 | Auth | 认证——确认"你是谁" |
| 9 | RBAC | 授权——加载"你能做什么" |
| 10 | RequirePerm | 粒度校验——针对单一路由检查"你能不能做这件事" |
| 11 | RateLimit | 最后——按认证后的用户/IP 精确计数，不浪费未认证请求的 Redis 调用 |

### 为什么 Recovery 必须在 Trace 前面

```
❌ Trace → Recovery
   Trace 创建 span → Recovery defer recover → handler panic → recover 捕获
   但 Trace 的 span 创建代码本身不在 recover 保护内，若 Trace 初始化时 panic 则无捕获

✅ Recovery → Trace
   Recovery defer recover → Trace 创建 span → handler panic
   defer 出栈：span.End() → recover 捕获
   即使 Trace 初始化 panic，Recovery 也能捕获并记日志
```

### 为什么 RequestID 不放在 Recovery 后面

```
RequestID → Recovery
  request_id 生成 → 后续任意 middleware/handler panic
  → Recovery 从 c.Get("request_id") 取到值 → 写入 panic 日志
  方便按 request_id 在日志系统中检索
```

## 各中间件详情

### RequestID

```
1. requestID := c.GetHeader("X-Request-ID")
2. if requestID == "" → requestID = uuid.New()
3. c.Set("request_id", requestID)
4. c.Header("X-Request-ID", requestID)
5. c.Next()
```

### Recovery

```
defer:
  if err := recover(); err != nil {
      logger.Error("panic recovered",
          zap.Any("error", err),
          zap.String("stack", debug.Stack()),
          zap.String("request_id", c.GetString("request_id")),
      )
      c.AbortWithStatusJSON(500, response.Err(errcode.ErrInternal))
  }
```

### Trace

```
1. ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), ...)
2. ctx, span := tracer.Start(ctx, "HTTP "+c.Request.Method+" "+c.FullPath())
3. defer span.End()
4. c.Request = c.Request.WithContext(ctx)
5. c.Next()
6. span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))
```

### Logger（结构化日志）

```
start := time.Now()
c.Next()
latency := time.Since(start)
logger.Info("request",
    zap.String("method", c.Request.Method),
    zap.String("path", c.Request.URL.Path),
    zap.Int("status", c.Writer.Status()),
    zap.Duration("latency", latency),
    zap.String("ip", c.ClientIP()),
    zap.String("request_id", c.GetString("request_id")),
    zap.Int("body_size", c.Writer.Size()),
)
```

### Gzip（响应压缩）

```go
// core/middleware/gzip.go
import "github.com/gin-contrib/gzip"

func Gzip() gin.HandlerFunc {
    return gzip.Gzip(
        gzip.DefaultCompression,  // level 6
        gzip.WithExcludedPaths([]string{"/metrics", "/health", "/ready"}),
        gzip.WithExcludedExtensions([]string{".png", ".jpg", ".gif", ".ico"}),
    )
}
```

压缩策略：响应体 > 1KB 才压缩（gin-contrib/gzip 默认行为）。排除 `/metrics` `/health`——Prometheus 抓取不需要压缩，健康检查响应极短。排除图片扩展名——已压缩资源不二次压缩。

### SecurityHeaders（安全响应头）

```go
// core/middleware/security_headers.go
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Frame-Options", "DENY")                        // 防 clickjacking
        c.Header("X-Content-Type-Options", "nosniff")              // 防 MIME 嗅探
        c.Header("X-XSS-Protection", "1; mode=block")              // 防反射 XSS（旧浏览器）
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Cache-Control", "no-store, no-cache, must-revalidate")  // API 响应不缓存
        c.Header("Pragma", "no-cache")
        if !strings.Contains(c.FullPath(), "/swagger") {
            c.Header("Content-Security-Policy",
                "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'")
        }
        c.Next()
    }
}
```
CSP 排除 `/swagger`——Swagger UI 需要加载外部 CDN 资源。生产环境可通过 nginx `add_header` 统一加。

### Auth（认证——你是谁）

```go
// extends/rbac/transport/auth_middleware.go
func AuthMiddleware(jwtMgr *coreJWT.JWTManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        header := c.GetHeader("Authorization")
        if header == "" || !strings.HasPrefix(header, "Bearer ") {
            c.AbortWithStatusJSON(401, response.Err(errcode.ErrUnauthorized))
            return
        }
        claims, err := jwtMgr.ParseToken(c.Request.Context(), header[7:])
        if err != nil {
            c.AbortWithStatusJSON(401, response.Err(errcode.ErrUnauthorized))
            return
        }
        c.Set("user_id", claims.UserID)
        c.Set("role", claims.Role)
        c.Next()
    }
}
```

### RBAC（授权——你能做什么）

```go
// extends/rbac/transport/rbac_middleware.go
func RBACMiddleware(roleSvc port.RoleService) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("user_id")
        perms, err := roleSvc.GetUserPermissions(c.Request.Context(), userID)
        if err != nil {
            c.AbortWithStatusJSON(500, response.Err(errcode.ErrInternal))
            return
        }
        permSet := make(map[string]bool, len(perms))
        for _, p := range perms { permSet[p] = true }
        c.Set("permissions", permSet)  // Handler 按需读取
        c.Next()
    }
}
```

### RequirePerm（粒度校验——你能不能做这件事）

```go
// extends/rbac/transport/require_perm.go
func RequirePerm(code string) gin.HandlerFunc {
    return func(c *gin.Context) {
        permSet := c.GetStringMapBool("permissions")
        if !permSet[code] {
            c.AbortWithStatusJSON(403, response.Err(errcode.ErrForbidden))
            return
        }
        c.Next()
    }
}
```

### RateLimit（滑动窗口）

```
key := "ratelimit:" + c.ClientIP() + ":" + c.FullPath()
1. Redis ZREMRANGEBYSCORE key 0 (now - window)
2. count := Redis ZCARD key
3. if count >= limit {
       c.Header("X-RateLimit-Limit", limit)
       c.Header("X-RateLimit-Remaining", 0)
       c.Header("Retry-After", window.Seconds())
       c.AbortWithStatusJSON(429, response.Err(errcode.ErrTooManyRequest))
       return
   }
4. Redis ZADD key (now, now_uuid)
5. Redis EXPIRE key window
6. c.Next()
```

## 使用方式

```go
// core/router/engine.go
func NewEngine(cfg *config.Config, logger *zap.Logger) *gin.Engine {
    r := gin.New()

    // TrustedProxies：正确解析 X-Forwarded-For/X-Real-IP，防 IP 伪造绕过限流
    r.SetTrustedProxies(cfg.Server.TrustedProxies) // ["127.0.0.1", "10.0.0.0/8"]

    r.Use(middleware.RequestID())         // 1
    r.Use(middleware.Recovery())          // 2 ← 必须最外层
    r.Use(middleware.Trace())             // 3
    r.Use(middleware.Logger(logger))      // 4
    r.Use(middleware.Gzip())              // 5
    r.Use(middleware.SecurityHeaders())   // 6
    r.Use(middleware.CORS(cfg.CORS))      // 7
    return r
}

// extends/user/register.go — 登录接口加限流
func (m *Module) RegisterPublic(r *gin.RouterGroup) {
    auth := r.Group("/auth")
    // RateLimit 最后加，确保已经是认证用户的 IP 在精确计数
    auth.Use(middleware.RateLimit(m.cache, 5, time.Minute))
    {
        auth.POST("/login", m.handler.Login)
    }
}

// 需要认证的路由组
protected := r.Group("")
protected.Use(Auth(jwtMgr), RBAC(roleSvc))
{
    // 所有 protected 下的 handler 都能拿到 user_id + role + permissions
}

// 需要特定权限的路由
adminGroup := r.Group("")
adminGroup.Use(Auth(jwtMgr), RBAC(roleSvc), RequirePerm("user:delete"))
{
    adminGroup.DELETE("/users/:id", userHandler.Delete)
}
```
