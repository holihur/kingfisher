# Performance Benchmark — 性能基准

## 目标指标

以下为脚手架目标（不含业务逻辑），新项目在此基础上加上自身业务开销。

| 指标 | 目标值 | 测试条件 |
|------|--------|----------|
| P50 延迟 | < 5ms | 不含 DB/Redis 的纯框架开销 |
| P99 延迟 | < 50ms | 含 DB 查询的典型接口 |
| 单机 QPS | ≥ 2000 | 4C8G 实例，`/api/v1/users?page=1&page_size=20` |
| 启动时间 | < 3s | 含 DB 连接 + 迁移检查 |
| 内存（空闲）| < 50MB | 无请求 |
| 内存（满载）| < 200MB | 2000 QPS 持续 1min |
| Docker 镜像 | < 15MB | 编译后 Alpine |
| GC 暂停 | < 1ms | P99 |

## 测试环境基准

| 参数 | 值 |
|------|-----|
| 实例 | AWS c5.large (2C4G) 或等价 |
| Go 版本 | 1.23 |
| GOMAXPROCS | 2（默认 = CPU 核数） |
| MySQL | 同可用区 RDS db.t3.medium，连接池 100 |
| Redis | 同可用区 ElastiCache cache.t3.micro |

## 各接口的性能预算

| 接口 | P50 | P99 | QPS 预算 |
|------|-----|-----|----------|
| GET /health | < 1ms | < 2ms | — |
| GET /api/v1/users?page=1&page_size=20 | < 20ms | < 50ms | 500 |
| GET /api/v1/users/:id | < 5ms | < 15ms | 200 |
| POST /api/v1/auth/login | < 30ms | < 80ms | 50 |
| GET /api/v1/menus/tree | < 10ms | < 30ms | 100 |
| PUT /api/v1/roles/:id/permissions | < 15ms | < 40ms | 20 |

> 注：超过预算的接口需要排查 N+1、缺索引、未命中缓存等问题。

## 性能测试工具

### Go Benchmark（单函数）

```go
// extends/user/adapter/mysql/repo_test.go
func BenchmarkUserRepo_FindAll(b *testing.B) {
    db := setupBenchDB(b)
    repo := adapter.NewUserRepo(db)
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _, _ = repo.FindAll(ctx, 1, 20, nil)
    }
}
```

```
go test -bench=. -benchmem -benchtime=3s ./extends/user/adapter/mysql/
```

### Vegeta（HTTP 压测）

```bash
# 登录拿 token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Abcd1234"}' | jq -r '.data.access_token')

# 生成负载
echo "GET http://localhost:8080/api/v1/users?page=1&page_size=20" | \
  vegeta attack -duration=30s -rate=100 -header "Authorization: Bearer $TOKEN" | \
  vegeta report -type=text

# 预期输出
# Requests      [total, rate]   3000, 100.00
# Duration      [total, attack] 30s, 30s
# Latencies     [min, mean, 50, 90, 95, 99, max]  2ms, 8ms, 5ms, 12ms, 18ms, 35ms, 80ms
# Success       [ratio]         100.00%
# Status Codes  [code:count]    200:3000
```

### wrk（快速摸底）

```bash
wrk -t4 -c100 -d30s --latency \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/users?page=1\&page_size=20
```

## 性能优化清单（按优先级）

| # | 优化项 | 影响范围 | 预期提升 |
|------|--------|----------|----------|
| 1 | 菜单树缓存（Redis） | GET /menus/tree | P99 50ms → 5ms |
| 2 | 权限列表缓存 | 每次请求的 RBAC 中间件 | P50 减少 3-5ms |
| 3 | DB 连接池预创建 | 冷启动请求 | 首个请求延迟减半 |
| 4 | GORM 预编译语句 | 高频查询 | P50 减少 1-2ms |
| 5 | JSON 序列化用 `jsoniter` | 所有 JSON 响应 | P50 减少 0.5-1ms |
| 6 | 分页查询用覆盖索引 | 大表翻页 | 末页 P99 200ms → 20ms |
| 7 | 静态文件 CDN（Logo 等） | 前端加载 | 首屏 2s → 500ms |
| 8 | Gin Release Mode | 所有请求 | P50 减少 0.2ms |

## 可观测验证

通过 Grafana Dashboard 持续验证：

```
Panel 1: HTTP Request Rate (by endpoint)       ← 各接口 QPS
Panel 2: P50/P95/P99 Latency (by endpoint)     ← 延迟分布
Panel 3: DB Query Duration (P50/P95/P99)       ← DB 耗时
Panel 4: Redis Command Duration (P50/P95/P99)  ← Redis 耗时
Panel 5: GC Pause (P99)                        ← GC 暂停
Panel 6: Goroutine Count                       ← 协程泄漏检测
```

## 设计要点

- 性能基准应该在 CI 中跑（至少 commit 级别），检测回归——Vegeta 输出结果对比上一次
- 缓存命中率目标 > 90%（菜单树/权限列表/配置）
- 如果某个接口的 P99 超过预算的 2 倍，CI 告警但不阻断（防止环境抖动误报）
- 压测前预热：先发 100 个请求让连接池、缓存、JIT 就绪再测
