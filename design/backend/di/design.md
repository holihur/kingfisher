# DI — 依赖注入（Wire）

## 职责

使用 Google Wire 编译期生成依赖注入代码，避免手写组装和维护 `main.go` 中的大量构造函数调用。

## 为什么不用运行时 DI（dig/inject）

| | Wire | dig / fx |
|------|------|----------|
| 时机 | 编译期 | 运行时 |
| 错误发现 | `wire` 命令报错 | 启动时 panic |
| 性能 | 零反射 | 反射开销 |
| 依赖图 | 显式 Provider Set | 隐式 |

## 项目 Wire 结构

```
internal/wire/
├── core.go        # Core 组件的 Provider Set
├── user.go        # User 模块的 Provider Set
├── menu.go        # Menu 模块的 Provider Set
├── rbac.go        # RBAC 模块的 Provider Set
├── config.go      # Config 模块的 Provider Set
├── wire.go        # wire.Build(...) 主入口
└── wire_gen.go    # 自动生成——禁止手动编辑
```

## Core Provider Set

```go
// internal/wire/core.go
var CoreSet = wire.NewSet(
    // Config
    coreConfig.Load,

    // Logger
    coreLogger.New,

    // Database
    coreDB.NewDatabase,
    coreDB.NewUnitOfWork,

    // Cache (接口绑定实现)
    coreCache.NewRedisClient,
    wire.Bind(new(coreCache.Cache), new(*coreCache.RedisCache)),
    coreCache.NewRedisCache,

    // JWT
    coreJWT.NewJWTManager,

    // Middleware
    middleware.NewAuth,        // 来自 extends/rbac
    middleware.NewRBAC,        // 来自 extends/rbac
)
```

## Extends Provider Set

```go
// internal/wire/user.go
var UserSet = wire.NewSet(
    // Adapter
    adapter.NewUserRepo,
    wire.Bind(new(userPort.UserRepository), new(*adapter.UserRepo)),  // 接口绑定实现

    // Service
    app.NewUserService,

    // Handler
    transport.NewUserHandler,

    // Module (实现 core.Module 接口)
    transport.NewUserModule,
    wire.Bind(new(coreRouter.Module), new(*transport.UserModule)),
)
```

## 主 Wire 文件

```go
// internal/wire/wire.go
//go:build wireinject

func InitializeApp(configPath string) (*App, error)  // startup 调用: wire.InitializeApp("config/config.yaml") {
    wire.Build(
        CoreSet,
        UserSet,
        MenuSet,
        RBACSet,
        ConfigSet,
        NewApp,    // 聚合所有 Module
    )
    return nil, nil
}
```

## main.go 最终形态

```go
func main() {
    app, err := wire.InitializeApp("config/config.yaml")
    if err != nil { log.Fatal(err) }
    app.Run()   // 启动 HTTP Server
}
```

## App 聚合

```go
// cmd/server/app.go
type App struct {
    Engine   *gin.Engine
    Modules  []coreRouter.Module
    Logger   *zap.Logger
    DB       *gorm.DB
    Cache    coreCache.Cache
    Server   *http.Server
}

func NewApp(
    engine *gin.Engine,
    userMod coreRouter.Module,
    menuMod coreRouter.Module,
    rbacMod coreRouter.Module,
    configMod coreRouter.Module,
    jwtMw gin.HandlerFunc,
    rbacMw gin.HandlerFunc,
    logger *zap.Logger,
    cfg *coreConfig.Config,
) *App {
    // 注册所有模块
    coreRouter.RegisterModule(engine, userMod, jwtMw, rbacMw)
    coreRouter.RegisterModule(engine, menuMod, jwtMw, rbacMw)
    coreRouter.RegisterModule(engine, rbacMod, jwtMw, rbacMw)
    coreRouter.RegisterModule(engine, configMod, jwtMw, rbacMw)
    return &App{Engine: engine, Logger: logger}
}

func (a *App) Run() {
    // 优雅启动 + 关闭
}
```

## Makefile 中的 wire 命令

```makefile
wire:
	cd internal/wire && wire
```

## 设计要点

- `wire_gen.go` 是自动生成的，不要手动编辑，提交 git
- 每次新增/修改 Provider 后执行 `make wire`
- `wire.Bind` 将接口绑定到实现（编译期检查实现是否完整）
- Module 被注册到 `[]coreRouter.Module` 切片，Wire 自动填充
