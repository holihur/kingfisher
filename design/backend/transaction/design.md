# Transaction — 事务管理

## 职责

提供跨 Repository 的事务支持。遵循 **Service 层控制事务边界，Repository 层执行操作**。

## 核心设计：Unit of Work

```go
// core/database/unit_of_work.go
type UnitOfWork interface {
    Transaction(ctx context.Context, fn func(txCtx context.Context) error) error
}
```

## GORM 实现

```go
type GormUnitOfWork struct {
    db *gorm.DB
}

func (u *GormUnitOfWork) Transaction(ctx context.Context, fn func(txCtx context.Context) error) error {
    return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 将 *gorm.DB 注入 context，Repository 从 ctx 取
        txCtx := context.WithValue(ctx, txKey{}, tx)
        return fn(txCtx)
    })
}
```

## Repository 适配事务

```go
// adapter/mysql/user_repo.go
type UserRepo struct {
    db *gorm.DB
}

func (r *UserRepo) getDB(ctx context.Context) *gorm.DB {
    if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
        return tx   // 在事务中，用事务连接
    }
    return r.db.WithContext(ctx)  // 不在事务中，新建连接
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
    return r.getDB(ctx).Create(toPO(user)).Error
}
```

## Service 层使用

```go
// extends/user/app/service.go
type UserService struct {
    userRepo port.UserRepository
    uow      coreDB.UnitOfWork
}

// 同模块事务：两个操作（创建用户 + 创建用户档案）在同一个事务中
func (s *UserService) CreateUserWithProfile(ctx context.Context, username, email string) error {
    return s.uow.Transaction(ctx, func(txCtx context.Context) error {
        user := &domain.User{Username: username, Email: email}
        if err := s.userRepo.Create(txCtx, user); err != nil {
            return err  // 回滚
        }
        // 同模块第二个操作——共享同一事务
        if err := s.userRepo.Update(txCtx, user.ID, map[string]any{"status": 1}); err != nil {
            return err  // 回滚
        }
        return nil  // 提交
    })
}
```

## 事务传播

| 场景 | 行为 |
|------|------|
| Service A.Tx() 内调 Service B（无 Tx） | B 复用 A 的事务 |
| Service A.Tx() 内调 Service B.Tx() | GORM 不支持嵌套，实际复用同一个事务 |
| 独立调用 Service B | 无事务，每次操作独立提交 |

## 设计要点

- **事务边界在 Service 层**，不在 Repository 层
- **通过 Context 传递事务**：`txCtx := context.WithValue(ctx, txKey{}, tx)`
- Repository 不需要知道自己在事务中——`getDB(ctx)` 自动判断
- 只读操作（`FindByID`）不需要事务，但不能阻止它复用事务连接以实现一致性读
- **跨 extends 模块的事务**是可行的：Service 依赖多个 Repository 接口，底层同一个 `*gorm.DB`
