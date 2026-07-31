# Extends/Menu — 菜单管理

## 职责

菜单的树形 CRUD，支持目录/菜单/按钮三级。菜单是前端路由和 RBAC 的基础数据。

## 目录结构

```
extends/menu/
├── domain/menu.go              # Menu 实体（含 Children 树形）
├── port/repository.go          # MenuRepository 接口
├── app/service.go              # MenuService（树形构建）
├── adapter/mysql/
│   ├── model.go                # menuPO
│   └── repo.go                 # 实现 port.MenuRepository
├── transport/
│   ├── handler.go
│   └── register.go
└── wire.go
```

## Domain

```go
type Menu struct {
    ID         uint      `json:"id"`
    ParentID   uint      `json:"parent_id"`     // 0=顶级
    Name       string    `json:"name"`
    Path       string    `json:"path"`           // 前端路由 /admin/users
    Component  string    `json:"component"`      // 前端组件路径
    Icon       string    `json:"icon"`           // AntD 图标
    Sort       int       `json:"sort"`
    Type       int       `json:"type"`           // 1=目录 2=菜单 3=按钮
    Permission string    `json:"permission"`      // user:list
    Status     int       `json:"status"`         // 1=显示 0=隐藏
    Children   []Menu    `json:"children"`       // 树形子节点
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```

## Port

```go
type MenuRepository interface {
    FindAll(ctx context.Context) ([]domain.Menu, error)
    FindByID(ctx context.Context, id uint) (*domain.Menu, error)
    FindByParentID(ctx context.Context, parentID uint) ([]domain.Menu, error)
    FindByRoleIDs(ctx context.Context, roleIDs []uint) ([]domain.Menu, error)
    Create(ctx context.Context, menu *domain.Menu) error
    Update(ctx context.Context, id uint, updates map[string]any) error
    Delete(ctx context.Context, id uint) error
    HasChildren(ctx context.Context, parentID uint) (bool, error)
}
```

## Service

```go
type MenuService struct {
    repo  port.MenuRepository
    cache coreCache.Cache
}

func (s *MenuService) GetTree(ctx context.Context) ([]domain.Menu, error)
func (s *MenuService) GetByID(ctx context.Context, id uint) (*domain.Menu, error)
func (s *MenuService) Create(ctx context.Context, menu *domain.Menu) error
func (s *MenuService) Update(ctx context.Context, id uint, updates map[string]any) error
func (s *MenuService) Delete(ctx context.Context, id uint) error
```

### GetTree 核心逻辑

```
1. 尝试缓存：cache.Get("menu:tree")
2. miss → repo.FindAll() 拿到所有菜单（扁平列表）
3. 构建树：用 map[parentID] → 挂 Children
       for m := range menus {
           tree[parentID] = append(tree[parentID], m)
       }
       递归 BuildTree(0)
4. 写入缓存：cache.Set("menu:tree", tree, 10*time.Minute)
5. 返回树形结构
```

### BuildTree

```
func BuildTree(menus []Menu, parentID uint) []Menu {
    nodes := filter(menus, parentID)  // 找出当前层
    for i := range nodes {
        nodes[i].Children = BuildTree(menus, nodes[i].ID)  // 递归子节点
    }
    return nodes
}
```

### Delete 校验

```
1. HasChildren(id) → true → ErrMenuHasChildren，拒绝删除
```

## Handler

```go
type MenuHandler struct { svc *MenuService }

func (h *MenuHandler) GetTree(c *gin.Context)    // GET  /api/v1/menus/tree
func (h *MenuHandler) GetByID(c *gin.Context)    // GET  /api/v1/menus/:id
func (h *MenuHandler) Create(c *gin.Context)      // POST /api/v1/menus
func (h *MenuHandler) Update(c *gin.Context)      // PUT  /api/v1/menus/:id
func (h *MenuHandler) Delete(c *gin.Context)      // DELETE /api/v1/menus/:id
```

## 路由注册

```go
func (m *Module) RegisterProtected(r *gin.RouterGroup) {
    menus := r.Group("/menus")
    menus.GET("/tree", m.handler.GetTree)   // 任何登录用户可取（前端渲染侧边栏）
    menus.GET("/:id", m.handler.GetByID)
    menus.POST("", m.handler.Create)
    menus.PUT("/:id", m.handler.Update)
    menus.DELETE("/:id", m.handler.Delete)
}
```

## 设计要点

- 菜单树用递归构建，时间复杂度 O(n)
- 树数据缓存 10min，写操作后失效缓存
- Type 字段支持三种：目录（折叠项）、菜单（路由）、按钮（权限点）
- 前端根据返回的 Children 递归渲染侧边栏
