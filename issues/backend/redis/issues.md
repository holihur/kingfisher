# Redis — 设计与实现差异

> 来源：`design/backend/redis/design.md` 对照 `core/cache/redis.go`
> 排查日期：2026-07-31

## P2

### RDS-1 ✅ Cache 接口位置与设计不符
  **Status: ✅ Architecture documented in ADR**
- 设计：`port.Cache` 接口（端口层），Redis 实现位于 `adapter/redis/cache.go`
- 实现：接口 `Cache` 定义在 `core/cache/`（`core/cache/redis.go`），Redis 实现同包 `RedisCache`
- 影响：core 依赖具体 cache 包而非 port 层；设计意图是 core 零业务依赖、接口归 port（轻微架构偏差）

### RDS-2 ✅ 降级策略与设计冲突
  **Status: ✅ Redis Fatal — matches design**
- 设计：Redis 初始化失败应 `logger.Fatal`（启动即失败）
- 实现：`main.go` 降级为 Warn + `redisCache = nil`（限流/黑名单/缓存全部禁用但服务继续）
- 影响：与设计「Redis 是强依赖」冲突（见 acceptance A-13）；实际是更宽容的生产行为，需显式决策

## 一致项 ✅
- 连接参数（PoolSize/DialTimeout/ReadTimeout/WriteTimeout）、Ping 验证与设计一致
- `Get/Set/Delete/Exists/Incr/Expire` 接口齐备
- 抽象接口已存在（`cache.Cache`），`extends` 模块通过接口注入
