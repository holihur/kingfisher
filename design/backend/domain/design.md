# Domain — 领域模型

## 职责

纯业务实体，**零外部依赖**（不 import GORM、Gin、任何框架）。所有字段用 Go 原生类型，JSON tag 用于序列化。

## 模型定义

### User

```go
type User struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Password  string    `json:"-"`              // bcrypt hash，禁止序列化
    Email     string    `json:"email"`
    Avatar    string    `json:"avatar"`
    Status    int       `json:"status"`         // 1=启用 0=禁用
    RoleID    uint      `json:"role_id"`        // 关联角色
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### Menu（树形结构，支持无限层级）

```go
type Menu struct {
    ID        uint      `json:"id"`
    ParentID  uint      `json:"parent_id"`      // 0=顶级菜单
    Name      string    `json:"name"`            // 菜单名称
    Path      string    `json:"path"`            // 前端路由 /admin/users
    Component string    `json:"component"`       // 前端组件路径 Users/index.tsx
    Icon      string    `json:"icon"`            // Ant Design 图标名
    Sort      int       `json:"sort"`            // 排序
    Type      int       `json:"type"`            // 1=目录 2=菜单 3=按钮
    Permission string   `json:"permission"`      // 权限标识 user:list
    Status    int       `json:"status"`          // 1=显示 0=隐藏
    Children  []Menu    `json:"children,omitempty" gorm:"-"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### Role

```go
type Role struct {
    ID          uint         `json:"id"`
    Name        string       `json:"name"`
    Code        string       `json:"code"`       // 角色编码 admin/editor/viewer
    Description string       `json:"description"`
    Status      int          `json:"status"`
    Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
    Menus       []Menu       `json:"menus,omitempty" gorm:"many2many:role_menus;"`
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
}
```

### Permission

```go
type Permission struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`          // 权限名称
    Code        string    `json:"code"`          // 权限标识 user:create
    Resource    string    `json:"resource"`      // 资源 user
    Action      string    `json:"action"`        // 动作 create/read/update/delete
    CreatedAt   time.Time `json:"created_at"`
}
```

### SystemConfig（键值对）

```go
type SystemConfig struct {
    ID        uint      `json:"id"`
    Key       string    `json:"key"`             // site_name, max_login_attempts
    Value     string    `json:"value"`           // JSON 或纯文本
    Remark    string    `json:"remark"`           // 说明
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

## 领域错误

```go
// domain/errors.go
var (
    ErrUserNotFound     = errors.New("user not found")
    ErrUserExists       = errors.New("user already exists")
    ErrPasswordWrong    = errors.New("wrong password")
    ErrTokenExpired     = errors.New("token expired")
    ErrTokenInvalid     = errors.New("token invalid")
    ErrForbidden        = errors.New("forbidden")
    ErrMenuNotFound     = errors.New("menu not found")
    ErrRoleNotFound     = errors.New("role not found")
    ErrConfigNotFound   = errors.New("config not found")
)
```

## 设计要点

- `domain` 包不 import 任何框架（GORM/gorm 也不行）
- GORM tag 在 adapter 层通过嵌入解决，不在 domain 定义
- 领域错误用 `errors.New`，Service 层返回这些错误，Handler 层映射到 HTTP 状态码
- `Menu.Children` tag `gorm:"-"` 表示非数据库字段，adapter 层手动填充树形结构
