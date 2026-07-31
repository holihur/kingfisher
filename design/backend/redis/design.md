# Redis — 缓存连接

## 职责

初始化 Redis 连接，实现 `port.Cache` 接口。

## 对外接口

```go
func NewRedis(cfg RedisConfig) (*redis.Client, error)
func NewRedisCache(client *redis.Client) port.Cache
```

## 核心逻辑

```
1. rdb := redis.NewClient(&redis.Options{
       Addr:         host:port,
       Password:     cfg.Password,
       DB:           cfg.DB,
       PoolSize:     cfg.PoolSize,
       DialTimeout:  5s,
       ReadTimeout:  3s,
       WriteTimeout: 3s,
   })
2. rdb.Ping(ctx) → 验证连接
```

## Cache 实现（adapter/redis/cache.go）

```go
type RedisCache struct { client *redis.Client }

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
    return c.client.Get(ctx, key).Result()
}
func (c *RedisCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
    return c.client.Set(ctx, key, val, ttl).Err()
}
// ... 其余方法
```

## 使用场景

| 场景 | Key 模板 | TTL | 使用者 |
|------|----------|-----|--------|
| Token 黑名单 | `blacklist:token:{jti}` | =access_ttl | AuthService |
| Session 版本 | `user:sv:{user_id}` | 30min | AuthMiddleware |
| 限流计数 | `ratelimit:{ip}:{path}` | 1 min | RateLimit 中间件 |
| 菜单缓存 | `menu:role:{role_id}` | 10 min | MenuService |
| 系统配置 | `config:{key}` | 5 min | ConfigService |
| 登录失败计数 | `login_fail:{username}` | 15 min | AuthService |

## 设计要点

- 所有操作带 `context.Context`
- `port.Cache` 接口定义在 port 层，Service 依赖接口
- 缓存 miss 时返回 `redis.Nil`，Service 层据此决定是否回退 DB
