# Service Interface — Service 接口化

## 问题

当前 Handler 直接依赖具体 `*app.AuthService` struct：

```go
type AuthHandler struct {
    svc *app.AuthService  // ❌ 具体实现，单测无法 mock
}
```

Handler 的单元测试被迫实例化整个 Service 链（Service → Repo → DB），失去了单元测试的隔离性。

## 方案：为每个 Service 定义接口

放在 `extends/{module}/port/` 目录，与 Repository 接口同级。

### extends/user/port/service.go

```go
type AuthService interface {
    Register(ctx context.Context, username, password, email string) (*domain.User, error)
    Login(ctx context.Context, username, password string) (accessToken, refreshToken string, user *domain.User, err error)
    RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, err error)
}

type UserService interface {
    GetByID(ctx context.Context, id uint) (*domain.User, error)
    Update(ctx context.Context, id uint, updates map[string]any) error
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, page, pageSize int, keyword string) ([]domain.User, int64, error)
}
```

### extends/menu/port/service.go

```go
type MenuService interface {
    GetTree(ctx context.Context) ([]domain.Menu, error)
    GetByID(ctx context.Context, id uint) (*domain.Menu, error)
    Create(ctx context.Context, menu *domain.Menu) error
    Update(ctx context.Context, id uint, updates map[string]any) error
    Delete(ctx context.Context, id uint) error
}
```

### extends/rbac/port/service.go

```go
type RoleService interface {
    List(ctx context.Context) ([]domain.Role, error)
    GetByID(ctx context.Context, id uint) (*domain.Role, error)
    Create(ctx context.Context, role *domain.Role) error
    Update(ctx context.Context, id uint, updates map[string]any) error
    Delete(ctx context.Context, id uint) error
    AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error
    AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error
    GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
}

type PermissionService interface {
    List(ctx context.Context) ([]domain.Permission, error)
}
```

### extends/config/port/service.go

```go
type ConfigService interface {
    GetAll(ctx context.Context) ([]domain.SystemConfig, error)
    Get(ctx context.Context, key string) (*domain.SystemConfig, error)
    Set(ctx context.Context, key, value string) error
    Delete(ctx context.Context, key string) error
}
```

## Handler 改为依赖接口

```go
type AuthHandler struct {
    svc port.AuthService  // ✅ 接口，单测轻松 mock
}

func NewAuthHandler(svc port.AuthService) *AuthHandler {
    return &AuthHandler{svc: svc}
}
```

## Wire 绑定

```go
var UserSet = wire.NewSet(
    adapter.NewUserRepo,
    wire.Bind(new(userPort.UserRepository), new(*adapter.UserRepo)),

    app.NewAuthService,
    wire.Bind(new(userPort.AuthService), new(*app.AuthService)),   // 接口绑定

    app.NewUserService,
    wire.Bind(new(userPort.UserService), new(*app.UserService)),   // 接口绑定

    transport.NewAuthHandler,
    transport.NewUserHandler,
    transport.NewUserModule,
    wire.Bind(new(coreRouter.Module), new(*transport.UserModule)),
)
```

## Handler 单测对比

```go
// ❌ 之前——需要注入完整 Service → 连 DB
func TestHandler_Old(t *testing.T) {
    db := setupTestDB(t)
    repo := adapter.NewUserRepo(db)
    svc := app.NewUserService(repo, nil, nil)  // 三个真实依赖
    handler := transport.NewUserHandler(svc)    // 扯了一整条链
}

// ✅ 现在——只 mock Service 接口
func TestHandler_New(t *testing.T) {
    mockSvc := &MockUserService{
        GetByIDFunc: func(ctx context.Context, id uint) (*domain.User, error) {
            return &domain.User{ID: 1, Username: "test"}, nil
        },
    }
    handler := transport.NewUserHandler(mockSvc)  // 只 mock 一层
}
```

## 接口粒度原则

| 原则 | 说明 |
|------|------|
| 接口放在 port/ | 和 Repository 接口同级，依赖反转一致 |
| 接口不要太细 | 一个模块一个 Service 接口，不拆成 N 个单方法接口 |
| 接口不要太粗 | AuthService 和 UserService 分开，职责清晰 |
| Mock 自动生成 | 用 `mockgen` 或手写 `test/testutil/` |

## 设计要点

- 这是**锦上添花**：核心架构（Core+Extends+port 接口）已就绪，Service 接口化让测试隔离性更彻底
- 不强制所有模块都要 Service 接口——只对 Handler 层需要单测隔离的模块做
- 接口定义在 `port/` 还是 `app/`？答：`port/`，因为 port 是依赖反转边界
