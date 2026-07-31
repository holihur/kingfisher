# Read/Write Split — 数据库读写分离

## 现状

当前所有操作走同一个 `*gorm.DB` 连接。

## 目标

读操作走从库（Read Replica），写操作走主库（Master）。初期可读写同库（为未来拆分做准备），但架构上支持主从切换。

## 方案：GORM DBResolver Plugin

GORM 内置了 `DBResolver` 插件，基于 `*gorm.DB` 的方法自动路由。

```go
// core/database/gorm.go
func NewGormWithRW(cfg MySQLConfig, logger *zap.Logger) *gorm.DB {
    masterDSN := buildDSN(cfg.Master)
    db, _ := gorm.Open(mysql.Open(masterDSN), &gorm.Config{Logger: gormLogger})

    if cfg.Replica.Host != "" {
        // 有从库——启用读写分离
        replicaDSN := buildDSN(cfg.Replica)
        db.Use(dbresolver.Register(dbresolver.Config{
            Sources:  []gorm.Dialector{mysql.Open(masterDSN)},
            Replicas: []gorm.Dialector{mysql.Open(replicaDSN)},
            Policy:   dbresolver.RandomPolicy{},     // 多个从库时随机选
        }))
    }
    // 无从库——读写同库（不影响现有行为）

    return db
}
```

## 配置扩展

```yaml
mysql:
  # 单库模式（开发/小项目）——和现有 config 兼容
  host: 127.0.0.1
  port: 3306
  user: root
  password: ""
  database: kingfisher

  # 主从模式（生产）——新增字段
  master:                     # 主库（写）
    host: master.db.internal
    port: 3306
    user: root
    password: ${MYSQL_MASTER_PASSWORD}
    database: kingfisher
  replica:                    # 从库（读），可选
    host: replica.db.internal
    port: 3306
    user: readonly
    password: ${MYSQL_REPLICA_PASSWORD}
    database: kingfisher
```

**优先级**：如果 `master` 字段存在，优先用主从模式；否则回退到单库 `host/port`。

## DBResolver 自动路由规则

| 操作 | 路由到 | GORM 方法 |
|------|--------|-----------|
| SELECT | Replica | `Find`, `First`, `Take`, `Count`, `Row`, `Rows` |
| INSERT | Master | `Create`, `Save` |
| UPDATE | Master | `Update`, `Updates`, `Save` |
| DELETE | Master | `Delete` |
| 事务 | Master | `Transaction` |
| Raw SQL | 默认 Master | `Raw().Scan()` 需手动指定 |

## 强制走主库（特定读操作需要）

```go
// 场景：登录后立即查用户——需要最新数据，不能走从库（可能有复制延迟）
func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
    var po userPO
    err := r.db.Clauses(dbresolver.Write).  // ← 强制走主库
        WithContext(ctx).Where("username = ?", username).First(&po).Error
    return po.toDomain(), err
}
```

## Repository 适配

```go
type UserRepo struct {
    db *gorm.DB
}

// 普通读——自动走从库
func (r *UserRepo) FindAll(ctx context.Context, page, pageSize int) ([]domain.User, int64, error) {
    // DBResolver 自动识别 Find → 路由到 Replica
    return users, total, r.db.WithContext(ctx).Find(&pos).Error
}

// 强一致性读——走主库
func (r *UserRepo) FindByIDAfterWrite(ctx context.Context, id uint) (*domain.User, error) {
    return user, r.db.Clauses(dbresolver.Write).WithContext(ctx).First(&po, id).Error
}
```

## 设计要点

1. **开发环境不分主从**——不配置 replica 即可，单库模式自动兼容
2. **复制延迟 > 0 的业务场景**——用 `dbresolver.Write` 强制读主
3. 从库宕机——DBResolver 自动 fallback 到主库（配置 `Sources` 作为兜底）
4. 连接池独立——主库和从库各自维护连接池

## 何时启用

| 阶段 | 配置 | 说明 |
|------|------|------|
| 开发 | 单库 | `mysql.host/port` |
| 预发 | 单库 | 同上，检查日志确认无性能问题 |
| 生产（初期） | 单库 | 用云服务商托管 DB（自带主从，透明） |
| 生产（规模化） | 主从 | 配置 `master/replica`，启用 DBResolver |

## 可观测性

在 repository 层增加指标：

```go
var (
    dbReadCounter  = promauto.NewCounter(/* operation, table */)   // 读次数
    dbWriteCounter = promauto.NewCounter(/* operation, table */)   // 写次数
)
```

可在 Grafana 看到读写比例，判断是否需要增加从库或调整路由策略。
