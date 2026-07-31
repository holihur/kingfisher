# Core — 框架核心

## 职责

Core 提供**所有业务模块共用的基础设施**。Core 不知道任何业务模块的存在，只暴露接口让 extends 注册。

## Core 组成

```
core/
├── config/         # 配置加载 + 校验       → 详见 [../config/design.md](../config/design.md)
├── logger/         # Zap 封装 + 脱敏       → 详见 [../logger/design.md](../logger/design.md)
├── errcode/        # 错误码体系             → 详见 [../errcode/design.md](../errcode/design.md)
├── response/       # 统一响应格式
├── jwt/            # JWT 双 token + 黑名单  → 详见 [../jwt/design.md](../jwt/design.md)
├── database/       # GORM + 迁移            → 详见 [../mysql/design.md](../mysql/design.md)
├── cache/          # Redis 初始化 + Cache 接口 → 详见 [../redis/design.md](../redis/design.md)
├── telemetry/      # OTel + Prometheus     → 详见 [../observability/design.md](../observability/design.md)
├── middleware/      # 通用中间件             → 详见 [../middleware/design.md](../middleware/design.md)
└── router/          # Gin 引擎 + 扩展注册接口
```

## Core 对外暴露的关键接口

### 1. Module（扩展注册口 + 生命周期）

```go
// core/router/module.go
type Module interface {
    Name() string                                    // 模块名
    Init(ctx context.Context) error                  // 启动时：预热缓存、启动后台goroutine
    RegisterPublic(r *gin.RouterGroup)              // 公开路由（无需认证）
    RegisterProtected(r *gin.RouterGroup)           // 需登录
    Shutdown(ctx context.Context) error              // 关闭时：停止goroutine、flush buffer
}
```

所有 extends 模块实现此接口，`main.go` 按顺序调用：
```go
for _, m := range app.Modules {
    m.Init(ctx)   // 按注册顺序
}
// …路由注册…
for i := len(app.Modules) - 1; i >= 0; i-- {
    app.Modules[i].Shutdown(ctx)  // 倒序关闭
}
```

各模块 Init/Shutdown 用途：

| 模块 | Init | Shutdown |
|------|------|----------|
| audit | 启动异步写入 goroutine | flush buffer → 关闭 channel |
| rbac | 预热权限缓存到 Redis | — |
| menu | 预热菜单树缓存 | — |

### 2. DB 连接

```go
// core/database/gorm.go
func NewDatabase(cfg DatabaseConfig, logger *zap.Logger) *gorm.DB
```

返回 `*gorm.DB`，各 extends 的 adapter 接收它。

### 3. Cache 接口

```go
// core/cache/cache.go
type Cache interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, val any, ttl time.Duration) error
    Delete(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

`adapter/redis` 实现此接口，extends 的 service 依赖此接口。

### 4. JWT Manager

```go
// core/jwt/jwt.go
type JWTManager struct { ... }
func (m *JWTManager) GenerateToken(ctx context.Context, userID uint, role string) (access, refresh string, err error)
func (m *JWTManager) ParseToken(ctx context.Context, tokenStr string) (*Claims, error)
func (m *JWTManager) RefreshToken(ctx context.Context, refreshToken string) (string, error)
func (m *JWTManager) RevokeToken(ctx context.Context, tokenStr string) error
```

### 5. 统一响应

```go
// core/response/response.go
func OK(data any) *Response
func Err(code int) *Response
func ErrWithMsg(code int, msg string) *Response
func Page(items any, total int64, page, pageSize int) *Response
```

### 6. 通用中间件

```go
func RequestID() gin.HandlerFunc       // X-Request-ID 透传或生成
func Trace() gin.HandlerFunc           // OTel span 创建
func Recovery() gin.HandlerFunc        // panic → 500
func Logger(log *zap.Logger) gin.HandlerFunc
func CORS(cfg CORSConfig) gin.HandlerFunc
func RateLimit(cache port.Cache, limit int, window time.Duration) gin.HandlerFunc
```

**注意：Auth 和 RBAC 中间件不在 core 中！** 它们属于 extends/rbac 模块，因为认证逻辑（JWT payload 包含什么角色）和权限判断（哪些角色有权限）是业务相关的。Core 只提供 JWT 解析能力，不定义角色体系。

## Engine 工厂

```go
// core/router/engine.go
func NewEngine(cfg *config.Config, logger *zap.Logger) *gin.Engine {
    r := gin.New()
    r.Use(RequestID())
    r.Use(Trace())
    r.Use(Recovery())
    r.Use(Logger(logger))
    r.Use(CORS(cfg.CORS))
    return r
}

// 注册一个扩展模块
func Register(r *gin.Engine, mod Module, middlewares ...gin.HandlerFunc) {
    api := r.Group("/api/v1")
    mod.RegisterPublic(api)
    protected := api.Group("")
    for _, mw := range middlewares {
        protected.Use(mw)
    }
    mod.RegisterProtected(protected)
}
```

## Core 的依赖（外部库）

```
gin, gorm, go-redis, golang-jwt, viper, zap, otel, prometheus
```

**绝对不依赖任何 extends 包**。

## 设计要点

- Core 中的每个组件都是可替换的：换一个 logger 实现不影响 extends
- RouteRegistrar 接口让 extends 自行管理路由分组
- Auth/RBAC 中间件放在 extends 而非 core，因为角色体系是业务
- Core 的 config、logger 等可通过 `core.SetLogger()` 之类的全局函数获取，避免 extends 传参链路太长
