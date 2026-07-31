# Backend Overview — Core + Extends 架构

## 1. 架构理念

```
┌─────────────────────────────────────────┐
│                main.go                   │
│      加载 Core + 注册 Extends             │
├─────────────────────────────────────────┤
│  extends/   │  user  │  menu  │  rbac  │ config │  audit  │  ← 业务模块（可插拔）
├─────────────────────────────────────────┤
│  core/      │ config │ logger │ jwt  │ db │ cache │  ← 框架核心（可复用）
│             │ router │ middleware │ errcode │ telemetry │
└─────────────────────────────────────────┘
```

**依赖方向**：`main → extends → core`，绝对不可逆。Core 不知道 Extends 的存在。

## 2. 目录结构

```
kingfisher/
├── cmd/server/main.go
├── core/                              # 框架核心——零业务依赖
│   ├── config/config.go               # 配置加载
│   ├── logger/logger.go               # Zap 封装
│   ├── errcode/errcode.go             # 错误码框架
│   ├── response/response.go           # 统一响应
│   ├── jwt/jwt.go                     # JWT 引擎
│   ├── database/gorm.go               # GORM 初始化
│   ├── cache/redis.go                 # Redis 初始化
│   ├── telemetry/                     # OTel + Prometheus
│   │   ├── tracer.go
│   │   └── metrics.go
│   ├── middleware/                    # 通用中间件（不含业务）
│   │   ├── request_id.go
│   │   ├── trace.go
│   │   ├── recovery.go
│   │   ├── logger.go
│   │   ├── cors.go
│   │   └── ratelimit.go
│   └── router/
│       └── engine.go                  # Gin 引擎工厂 + 扩展注册点
└── extends/                           # 业务扩展——每个模块独立，依赖 core
    ├── user/                          # 用户模块
    │   ├── domain/user.go
    │   ├── port/repository.go
    │   ├── adapter/mysql/repo.go
    │   ├── app/service.go
    │   ├── transport/handler.go
    │   └── transport/register.go      # 向 core.router 注册路由
    ├── menu/                          # 菜单模块（同上结构）
    ├── rbac/                          # 角色权限模块（同上结构）
    ├── config/                        # 系统配置模块（同上结构）
    └── audit/                         # 审计日志模块（同上结构）
```

## 3. Core 暴露的注册接口

```go
// core/router/engine.go
func NewEngine(cfg *config.Config, logger *zap.Logger) *gin.Engine {
    r := gin.New()
    r.SetTrustedProxies(cfg.Server.TrustedProxies)
    // 挂载全局中间件
    r.Use(coreMiddleware.RequestID())
    r.Use(coreMiddleware.Recovery())     // 1 — 必须最外层，捕获所有下游 panic
    r.Use(coreMiddleware.Trace())        // 2
    r.Use(coreMiddleware.Logger(logger)) // 3
    r.Use(coreMiddleware.Gzip())         // 4
    r.Use(coreMiddleware.SecurityHeaders()) // 5
    r.Use(coreMiddleware.CORS(cfg.CORS)) // 6
    return r
}

func Register(r *gin.Engine, mod Module, jwtMw gin.HandlerFunc, rbacMw gin.HandlerFunc) {
    api := r.Group("/api/v1")
    mod.RegisterPublic(api)
    protected := api.Group("")
    protected.Use(jwtMw)
    {
        mod.RegisterProtected(protected)
    }
}
```

## 4. 扩展模块注册示例

```go
// extends/user/transport/register.go
type UserModule struct {
    handler *UserHandler
}

// 实现 core.Module
func (m *UserModule) RegisterPublic(r *gin.RouterGroup) {
    auth := r.Group("/auth")
    {
        auth.POST("/register", m.handler.Register)
        auth.POST("/login", m.handler.Login)
        auth.POST("/refresh", m.handler.Refresh)
    }
}

func (m *UserModule) RegisterProtected(r *gin.RouterGroup) {
    users := r.Group("/users")
    {
        users.GET("/:id", m.handler.GetByID)
        users.PUT("/:id", m.handler.Update)
    }
}

```

## 5. main.go 组装

```go
func main() {
    // 1. 初始化 Core
    cfg := coreConfig.Load("config/config.yaml")
    logger := coreLogger.New(cfg.Log)
    db := coreDB.New(cfg.MySQL, logger)
    rdb := coreCache.New(cfg.Redis)
    jwtMgr := coreJWT.New(cfg.JWT, rdb)
    ginEngine := coreRouter.NewEngine(cfg, logger, jwtMgr)

    // 2. 构建 Core 中间件
    jwtMw := coreMiddleware.Auth(jwtMgr)
    rbacMw := coreMiddleware.RBAC(rdb)  // 从缓存读权限

    // 3. 注册所有 Extends
    userModule := user.NewModule(db, rdb, jwtMgr)
    coreRouter.Register(ginEngine, userModule, jwtMw, rbacMw)

    menuModule := menu.NewModule(db, rdb)
    coreRouter.Register(ginEngine, menuModule, jwtMw, rbacMw)

    rbacModule := rbac.NewModule(db, rdb)
    coreRouter.Register(ginEngine, rbacModule, jwtMw, rbacMw)

    configModule := config.NewModule(db, rdb)
    coreRouter.Register(ginEngine, configModule, jwtMw, rbacMw)

    auditModule := audit.NewModule(db)
    coreRouter.Register(ginEngine, auditModule, jwtMw, rbacMw)

    // 4. 启动
    gracefullShutdown(ginEngine, cfg.Server.Port)
}
```

## 6. 模块内部结构（每个 extends 都一样）

```
extends/{module}/
├── domain/              # 领域实体（零依赖）
│   └── {entity}.go
├── port/                # 接口定义
│   └── repository.go     # 本模块需要的 Repository 接口
├── adapter/             # 接口实现
│   └── mysql/
│       ├── model.go     # GORM PO
│       └── repo.go      # 实现 port.Repository
├── app/                 # 用例 Service
│   └── service.go
└── transport/           # HTTP 层
    ├── handler.go        # Gin Handler
    └── register.go       # 实现 Module 接口
```

## 7. Core 与 Extends 的边界

| 属于 Core | 属于 Extends |
|-----------|-------------|
| config 加载 | User domain |
| Zap logger | Menu domain |
| JWT 生成/解析 | 用户注册登录逻辑 |
| GORM 连接池 | 菜单树形查询 |
| Redis 连接池 | 角色权限分配 |
| 通用中间件 | 各模块的 handler |
| 错误码框架 | 各模块的 service |
| 统一响应格式 | 各模块的 adapter |
| Module 接口 | 模块特有的缓存 key |
| Health/Metrics 端点 | API 接口的 Swagger 注解 |

## 8. 模块索引

| 分类 | 模块 | 文档 |
|------|------|------|
| Core | 配置 | [core/config](../core/design.md) |
| Core | 日志 | [core/logger](../core/../core/design.md) |
| Core | JWT | [core/jwt](../core/../core/design.md) |
| Core | 数据库 | [core/db](../core/../core/design.md) |
| Core | 缓存 | [core/cache](../core/../core/design.md) |
| Core | 错误码 | [core/errcode](../core/../core/design.md) |
| Core | 中间件 | [core/middleware](../core/../core/design.md) |
| Core | 路由引擎 | [core/router](../core/../core/design.md) |
| Extends | 用户 | [extends/user](../extends/user/design.md) |
| Extends | 菜单 | [extends/menu](../extends/menu/design.md) |
| Extends | RBAC | [extends/rbac](../extends/rbac/design.md) |
| Extends | 配置 | [extends/config](../extends/config/design.md) |
| Extends | 审计日志 | [extends/audit](../extends/audit/design.md) |
| Backend | API 契约 | [api-contract](../api-contract/design.md) |
| Backend | 读写分离 | [readwrite-split](../readwrite-split/design.md) |
| Backend | 迁移 SQL | [migration](../migration/design.md) |
| Backend | ADR | [adr](../adr/design.md) |
| Backend | Service 接口 | [service-interface](../service-interface/design.md) |
| Backend | 代码约束 | [guardrails](../guardrails/design.md) |
| Backend | 启动初始化 | [bootstrap](../bootstrap/design.md) |
| Backend | 缺失项追踪 | [gap-report](../audit/gap-report.md) |
| Backend | 请求校验规范 | [validation](../validation/design.md) |
| Frontend | 状态 UI 规范 | [state-ui](../../frontend/state-ui/design.md) |
| Backend | Swagger 注解 | [swagger-checklist](../swagger-checklist/design.md) |
| Backend | 性能基准 | [perf-bench](../perf-bench/design.md) |
| Frontend | 本地联调 | [local-dev](../../frontend/local-dev/design.md) |
| Shared | 类型共享 | [shared-types](../../shared-types/design.md) |
