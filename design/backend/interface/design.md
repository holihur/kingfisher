# Port Interface — 依赖反转边界

## 职责

定义应用核心接口。**Service 只依赖这些接口，不依赖具体实现**。这是架构质量的【关键】分水岭。

## 接口清单

### UserRepository

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

### MenuRepository

```go
type MenuRepository interface {
    FindAll(ctx context.Context) ([]domain.Menu, error)
    FindByID(ctx context.Context, id uint) (*domain.Menu, error)
    FindByParentID(ctx context.Context, parentID uint) ([]domain.Menu, error)
    Create(ctx context.Context, menu *domain.Menu) error
    Update(ctx context.Context, id uint, updates map[string]any) error
    Delete(ctx context.Context, id uint) error
}
```

### RoleRepository

```go
type RoleRepository interface {
    FindAll(ctx context.Context) ([]domain.Role, error)
    FindByID(ctx context.Context, id uint) (*domain.Role, error)
    Create(ctx context.Context, role *domain.Role) error
    Update(ctx context.Context, id uint, updates map[string]any) error
    Delete(ctx context.Context, id uint) error
    AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error
    GetPermissions(ctx context.Context, roleID uint) ([]domain.Permission, error)
    AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error
    GetMenus(ctx context.Context, roleID uint) ([]domain.Menu, error)
}
```

### PermissionRepository

```go
type PermissionRepository interface {
    FindAll(ctx context.Context) ([]domain.Permission, error)
    Create(ctx context.Context, perm *domain.Permission) error
    Delete(ctx context.Context, id uint) error
}
```

### ConfigRepository

```go
type ConfigRepository interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string) error
    GetAll(ctx context.Context) (map[string]string, error)
    Delete(ctx context.Context, key string) error
}
```

### Cache Interface

```go
type Cache interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Delete(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) (bool, error)
    Incr(ctx context.Context, key string) (int64, error)
    Expire(ctx context.Context, key string, ttl time.Duration) error
}
```

## 设计要点

1. **所有方法第一个参数是 `context.Context`** — 支持超时取消和 trace 传播
2. **返回值不是 pointer 就是 value+error** — 零值有明确语义
3. **Create/Update 接收 `*domain.Xxx`，查询返回 `domain.Xxx`** — 写入用指针（GORM 需要），读取用值（不可变）
4. **FindAll 返回 `(items, total, error)`** — 分页是通用需求
5. **接口放在 `port` 目录** — 一看就知道系统的能力边界

## Service 如何使用（依赖注入）

```go
// app/auth_service.go
type AuthService struct {
    userRepo port.UserRepository  // ← 接口，不是具体实现
    cache    port.Cache           // ← 接口
}
// wire 编译时注入 adapter/mysql.UserRepo 和 adapter/redis.Cache
```

## 测试如何 Mock

```go
// test/testutil/mock_repo.go
type MockUserRepo struct {
    FindByIDFunc func(ctx context.Context, id uint) (*domain.User, error)
}
func (m *MockUserRepo) FindByID(ctx context.Context, id uint) (*domain.User, error) {
    return m.FindByIDFunc(ctx, id)
}

// 单测
mock := &MockUserRepo{
    FindByIDFunc: func(ctx context.Context, id uint) (*domain.User, error) {
        return &domain.User{ID: id, Username: "test"}, nil
    },
}
svc := NewUserService(mock, nil)
```
