# Extends/User — 用户模块

## 职责

用户注册、登录、个人信息管理。依赖 core 的 JWT、DB、Cache。

## 目录结构

```
extends/user/
├── domain/user.go              # User 实体
├── port/repository.go          # UserRepository 接口
├── app/service.go              # UserService（注册、登录、CRUD）
├── adapter/mysql/
│   ├── model.go                # userPO（GORM 模型）
│   └── repo.go                 # 实现 port.UserRepository
├── transport/
│   ├── handler.go              # Gin Handler
│   └── register.go             # 实现 core.Module 接口
└── wire.go                     # Wire Provider
```

## Domain

```go
type User struct {
    ID             uint      `json:"id"`
    Username       string    `json:"username"`
    Password       string    `json:"-"`           // bcrypt hash
    Email          string    `json:"email"`
    Avatar         string    `json:"avatar"`
    Status         int       `json:"status"`      // 1=启用 0=禁用
    SessionVersion int       `json:"-"`           // 递增以强制踢出所有 session
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
```

## Port（本模块需要的接口）

```go
type UserRepository interface {
    FindByID(ctx context.Context, id uint) (*domain.User, error)
    FindByUsername(ctx context.Context, username string) (*domain.User, error)
    FindAll(ctx context.Context, page, pageSize int, filters map[string]any) ([]domain.User, int64, error)
    Create(ctx context.Context, user *domain.User) error
    Update(ctx context.Context, id uint, updates map[string]any) error
    Delete(ctx context.Context, id uint) error
}
```

**注意**：UserRepository 放在 `extends/user/port/`，不是 `core/`。每个模块定义自己需要的接口，adapter 实现它。这是**垂直切分**。

## Service

```go
type UserService struct {
    repo    port.UserRepository
    cache   coreCache.Cache       // 来自 core/cache
    jwtMgr  *coreJWT.JWTManager   // 来自 core/jwt
}

func (s *UserService) Register(ctx context.Context, username, password, email string) (*domain.User, error)
func (s *UserService) Login(ctx context.Context, username, password string) (string, string, *domain.User, error)
func (s *UserService) GetByID(ctx context.Context, id uint) (*domain.User, error)
func (s *UserService) Update(ctx context.Context, id uint, updates map[string]any) error
func (s *UserService) Delete(ctx context.Context, id uint) error
func (s *UserService) List(ctx context.Context, page, pageSize int, keyword string) ([]domain.User, int64, error)
func (s *UserService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
// ChangePassword 内部校验通过后自动调用 RevokeSessions——修改密码即踢出所有其它设备
func (s *UserService) RevokeSessions(ctx context.Context, userID uint) error
```

### Login 流程

```
1. repo.FindByUsername(username)
2. 不论用户是否存在，都执行 bcrypt.CompareHashAndPassword（防用户枚举）
   → 不存在时用固定假 hash "$2a$12$..." 比对，结果必然不匹配
3. 比对结果 → 不匹配 → 统一返回 domain.ErrPasswordWrong (10103)
   → 不区分"用户不存在"和"密码错误"
4. 记录失败次数（按 username），超限 → domain.ErrLoginFailed
5. 只在比对成功时：jwtMgr.GenerateToken(user.ID, user.Role)
6. 清除失败计数
7. 返回 accessToken, refreshToken, user
```

防用户枚举：即使 username 不存在，也执行 bcrypt 比对。攻击者无法通过返回码判断账号是否存在。

### ChangePassword 流程

```
1. 查用户 → userRepo.FindByID(userID)
2. bcrypt.CompareHashAndPassword(user.Password, oldPassword)
3. 不匹配 → domain.ErrPasswordWrong
4. validatePassword(newPassword)  // 密码强度校验
5. bcrypt.GenerateFromPassword(newPassword, cost=12)
6. repo.Update(userID, {"password": hashedNew})
7. repo.IncrementSessionVersion(userID)  // 踢出所有旧 session——修改密码后旧 token 立即失效
8. cache.Delete("user:sv:" + userID)
```

### RevokeSessions 流程（强制踢人）

使用 session_version 方案，比逐个 JTI 加黑名单更高效：

```
1. repo.IncrementSessionVersion(userID)    // UPDATE users SET session_version = session_version + 1
2. cache.Delete("user:sv:" + userID)       // 失效缓存
```

Auth 中间件校验逻辑：

```
1. JWT payload 含 "sv" (session_version)
2. Redis GET "user:sv:{userID}" → miss 则回 DB 查 users.session_version
3. claims.SessionVersion != currentVersion → 401 (token 被踢)
```

## Handler

```go
type UserHandler struct { svc *UserService }

// 公开
func (h *UserHandler) Register(c *gin.Context)
func (h *UserHandler) Login(c *gin.Context)
func (h *UserHandler) Refresh(c *gin.Context)

// 需登录
func (h *UserHandler) GetByID(c *gin.Context)
func (h *UserHandler) Update(c *gin.Context)
func (h *UserHandler) ChangePassword(c *gin.Context)     // PUT /users/me/password
func (h *UserHandler) GetMe(c *gin.Context)               // GET /users/me
func (h *UserHandler) GetMyPermissions(c *gin.Context)    // GET /users/me/permissions

// 需管理员
func (h *UserHandler) List(c *gin.Context)
func (h *UserHandler) Delete(c *gin.Context)
func (h *UserHandler) RevokeSessions(c *gin.Context)     // DELETE /users/:id/sessions
```

## 路由注册（实现 core.Module）

```go
// transport/register.go
type Module struct {
    handler *UserHandler
    authMw  gin.HandlerFunc
}
func (m *Module) Name() string { return "user" }
func (m *Module) RegisterPublic(r *gin.RouterGroup) {
    auth := r.Group("/auth")
    auth.POST("/register", middleware.RateLimit(m.cache, 2, 5*time.Minute), m.handler.Register)  // 注册限流：2次/5min per IP，防垃圾注册
    auth.POST("/login", middleware.RateLimit(m.cache, 5, time.Minute), m.handler.Login)         // 登录限流：5次/min
    auth.POST("/refresh", m.handler.Refresh)
}
func (m *Module) RegisterProtected(r *gin.RouterGroup) {
    users := r.Group("/users")
    users.GET("/:id", m.handler.GetByID)       // 自己 + 管理员
    users.GET("/me", m.handler.GetMe)                             // 当前用户信息
    users.GET("/me/permissions", m.handler.GetMyPermissions)      // 当前用户权限列表（前端 fetchUserInfo 依赖）
    users.PUT("/:id", m.handler.Update)
    users.PUT("/me/password", m.handler.ChangePassword)       // 修改自己密码（任何登录用户）
    users.GET("", m.handler.List)                              // 管理员
    users.DELETE("/:id", m.handler.Delete)                     // 管理员
    users.DELETE("/:id/sessions", m.handler.RevokeSessions)    // 管理员踢出用户
}
```

## 工厂函数

```go
func NewModule(db *gorm.DB, cache coreCache.Cache, jwtMgr *coreJWT.JWTManager) core.Module {
    repo := adapter.NewUserRepo(db)
    svc := app.NewUserService(repo, cache, jwtMgr)
    handler := transport.NewUserHandler(svc)
    return &transport.Module{handler: handler}
}
```

## Swagger 注解示例

```go
// @Summary 用户登录
// @Tags User
// @Accept json
// @Produce json
// @Param body body LoginReq true "登录参数"
// @Success 200 {object} response.Response{data=LoginResp}
// @Router /api/v1/auth/login [post]
```
