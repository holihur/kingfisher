# Extends/RBAC — 角色权限管理

## 职责

角色 CRUD、权限定义、角色分配权限、角色分配菜单。同时提供 RBAC 中间件给 core.router 使用。

## 目录结构

```
extends/rbac/
├── domain/
│   ├── role.go                  # Role 实体
│   └── permission.go            # Permission 实体
├── port/
│   ├── role_repo.go             # RoleRepository 接口
│   └── perm_repo.go            # PermissionRepository 接口
├── app/
│   ├── role_service.go          # 角色管理
│   └── permission_service.go    # 权限管理
├── adapter/mysql/
│   ├── model.go
│   ├── role_repo.go
│   └── perm_repo.go
├── transport/
│   ├── role_handler.go
│   ├── permission_handler.go
│   ├── auth_middleware.go       # Auth 中间件（认证：你是谁）
│   ├── rbac_middleware.go       # RBAC 中间件（授权：你能做什么）
│   ├── require_perm.go          # RequirePerm 中间件（粒度：你能做这件事吗）
│   └── register.go
└── wire.go
```

## Domain

### Role

```go
type Role struct {
    ID          uint         `json:"id"`
    Name        string       `json:"name"`
    Code        string       `json:"code"`           // admin | editor | viewer
    Description string       `json:"description"`
    Status      int          `json:"status"`
    Permissions []Permission `json:"permissions,omitempty"`  // 关联权限
    Menus       []Menu       `json:"menus,omitempty"`         // 关联菜单
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
}
```

### Permission

```go
type Permission struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`              // 用户创建
    Code      string    `json:"code"`              // user:create
    Resource  string    `json:"resource"`           // user
    Action    string    `json:"action"`             // create|read|update|delete
    CreatedAt time.Time `json:"created_at"`
}
```

## 预设权限清单

| Code | Resource | Action | 说明 |
|------|----------|--------|------|
| user:list | user | read | 查看用户列表 |
| user:create | user | create | 创建用户 |
| user:update | user | update | 更新用户 |
| user:delete | user | delete | 删除用户 |
| menu:list | menu | read | 查看菜单 |
| menu:create | menu | create | 创建菜单 |
| menu:update | menu | update | 更新菜单 |
| menu:delete | menu | delete | 删除菜单 |
| role:list | role | read | 查看角色 |
| role:create | role | create | 创建角色 |
| role:update | role | update | 更新角色 |
| role:delete | role | delete | 删除角色 |
| config:list | config | read | 查看配置 |
| config:update | config | update | 更新配置 |

## Port

```go
type RoleRepository interface {
    FindAll(ctx context.Context) ([]domain.Role, error)
    FindByID(ctx context.Context, id uint) (*domain.Role, error)
    FindByCode(ctx context.Context, code string) (*domain.Role, error)
    Create(ctx context.Context, role *domain.Role) error
    Update(ctx context.Context, id uint, updates map[string]any) error
    Delete(ctx context.Context, id uint) error
    AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error
    AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error
    GetUserPermissions(ctx context.Context, userID uint) ([]string, error) // 返回 perm codes
}

type PermissionRepository interface {
    FindAll(ctx context.Context) ([]domain.Permission, error)
    Create(ctx context.Context, p *domain.Permission) error
    Delete(ctx context.Context, id uint) error
}
```

## Service

```go
type RoleService struct {
    roleRepo port.RoleRepository
    cache    coreCache.Cache
}

func (s *RoleService) List(ctx context.Context) ([]domain.Role, error)
func (s *RoleService) GetByID(ctx context.Context, id uint) (*domain.Role, error)
func (s *RoleService) Create(ctx context.Context, role *domain.Role) error
func (s *RoleService) Update(ctx context.Context, id uint, updates map[string]any) error
func (s *RoleService) Delete(ctx context.Context, id uint) error
// Update/Delete 含角色层级校验：不能修改/删除角色 code 为 "admin" 的角色；
// 低 role_level 的用户不能修改高 role_level 的用户的角色
func (s *RoleService) AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error
func (s *RoleService) AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error
func (s *RoleService) GetRolePermissions(ctx context.Context, roleID uint) ([]domain.Permission, error)
func (s *RoleService) GetRoleMenus(ctx context.Context, roleID uint) ([]domain.Menu, error)
func (s *RoleService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
```

## RBAC 中间件

三个独立文件，单一职责——详见 [middleware 设计文档](../middleware/design.md)。

```go
// transport/auth_middleware.go
func AuthMiddleware(jwtMgr *coreJWT.JWTManager) gin.HandlerFunc

// transport/rbac_middleware.go
func RBACMiddleware(roleSvc port.RoleService) gin.HandlerFunc

// transport/require_perm.go
func RequirePerm(code string) gin.HandlerFunc
```

## RBAC 中间件（旧代码已废弃，保留映射）

> 实现已迁移到 [middleware 设计文档](../middleware/design.md)，此处不再保留重复代码。
```

## Handler

```go
type RoleHandler struct { svc *RoleService }

func (h *RoleHandler) List(c *gin.Context)              // GET  /api/v1/roles
func (h *RoleHandler) GetByID(c *gin.Context)           // GET  /api/v1/roles/:id
func (h *RoleHandler) Create(c *gin.Context)            // POST /api/v1/roles
func (h *RoleHandler) Update(c *gin.Context)            // PUT  /api/v1/roles/:id
func (h *RoleHandler) Delete(c *gin.Context)            // DELETE /api/v1/roles/:id
func (h *RoleHandler) GetPermissions(c *gin.Context)    // GET  /api/v1/roles/:id/permissions
func (h *RoleHandler) AssignPerms(c *gin.Context)       // PUT  /api/v1/roles/:id/permissions
func (h *RoleHandler) GetMenus(c *gin.Context)          // GET  /api/v1/roles/:id/menus
func (h *RoleHandler) AssignMenus(c *gin.Context)       // PUT  /api/v1/roles/:id/menus

type PermissionHandler struct { svc *PermissionService }

func (h *PermissionHandler) List(c *gin.Context)        // GET  /api/v1/permissions
```

## 路由注册

```go
func (m *Module) RegisterProtected(r *gin.RouterGroup) {
    roles := r.Group("/roles", RequirePerm("role:list"))
    roles.GET("", m.roleHandler.List)
    roles.GET("/:id", m.roleHandler.GetByID)
    roles.POST("", RequirePerm("role:create"), m.roleHandler.Create)
    roles.PUT("/:id", RequirePerm("role:update"), m.roleHandler.Update)
    roles.DELETE("/:id", RequirePerm("role:delete"), m.roleHandler.Delete)
    roles.GET("/:id/permissions", RequirePerm("role:list"), m.roleHandler.GetPermissions)
    roles.PUT("/:id/permissions", RequirePerm("role:update"), m.roleHandler.AssignPerms)
    roles.GET("/:id/menus", RequirePerm("role:list"), m.roleHandler.GetMenus)
    roles.PUT("/:id/menus", RequirePerm("role:update"), m.roleHandler.AssignMenus)

    perms := r.Group("/permissions", RequirePerm("role:list"))
    perms.GET("", m.permHandler.List)
}
```

## 权限查询优化

```
GetUserPermissions(userID):
1. cache.Get("user:perms:{userID}")
2. miss → DB 查 user → role → role_permissions → permissions
3. cache.Set("user:perms:{userID}", codes, 30*time.Minute)
4. 角色/权限变更时，清空所有关联用户的权限缓存
```
