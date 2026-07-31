# ADR — Architecture Decision Records

## ADR-001: 选择 Core + Extends 架构而非 DDD 分层

**状态**: 已采纳
**日期**: 2026-07-31

### 背景

需要为多个 Go 后台项目提供脚手架。项目类型相似（管理后台、CRUD 占比 80%），但各有特殊需求（支付、IM、CMS）。

### 决策

选择 **Core + Extends 插件化**架构。

### 考虑的替代方案

| 方案 | 优点 | 缺点 | 为什么不选 |
|------|------|------|-----------|
| 标准分层 (handler→service→repo) | 简单，新人上手快 | 无法复用核心能力；每个项目 fork 后越走越远 | 不适合多项目复用场景 |
| DDD (聚合根/值对象/领域事件) | 复杂业务建模强 | 学习曲线高，80% CRUD 项目属于过度设计 | 管理后台的复杂度不在此 |
| CQRS + Event Sourcing | 审计、回溯、性能 | 运维复杂度剧增，需要 Kafka + 事件存储 | 管理后台不需要事件溯源 |
| Clean Architecture (UseCase/Entity) | 与 Core+Extends 本质类似 | 目录层级更深 | 语义上 Core+Extends 更直观（core 就一个包，extends 可独立仓库） |

### 后果

- ✅ 新项目只需写 extends 模块，core 零改动
- ✅ extends 可以独立发布为 Go module（`kingfisher-contrib/user`）
- ⚠️ Core 变更需要考虑向后兼容
- ⚠️ extends 之间跨模块调用需走 port 接口

---

## ADR-002: JWT 而非 Session

**状态**: 已采纳
**日期**: 2026-07-31

### 背景

需要认证方案。选项：Session（Cookie+Redis）vs JWT（无状态 token）vs OAuth2（第三方登录）。

### 决策

使用 **JWT 双 token**（access + refresh），access token 2h，refresh token 7d。

### 考虑的替代方案

| 方案 | 为什么不选 |
|------|-----------|
| Session + Redis | 每次请求查 Redis，多一次网络 IO；水平扩展需共享 Redis，但本就有 Redis |
| 单 JWT（超长 TTL） | 无法主动注销，token 泄露风险大 |
| OAuth2/OIDC | 管理后台不需要社交登录，徒增复杂度 |

### 关于 "JWT 无法主动注销"

通过 Redis 黑名单解决。注销时将 token 的 JTI 写入 Redis，TTL 设为 token 剩余有效期。权衡：每次请求多一次 Redis EXISTS 查询（< 1ms），换来主动注销能力。

### 为什么不用 Refresh Token Rotation

标准 rotation 是每次 refresh 换新 refresh token。此处选择不 rotation（refresh 只在快过期时更新），简化前端逻辑。

### 后果

- ✅ 无状态，水平扩展友好
- ✅ 可主动注销
- ⚠️ 每次请求查 Redis 黑名单（< 1ms，可接受）

---

## ADR-003: golang-migrate 而非 GORM AutoMigrate

**状态**: 已采纳
**日期**: 2026-07-31

### 背景

ORM 的 AutoMigrate 很方便，但生产环境有风险。

### 决策

开发环境可用 AutoMigrate 加速迭代，**生产环境必须用 golang-migrate + 纯 SQL**。

### 原因

| GORM AutoMigrate | golang-migrate |
|------------------|----------------|
| 不知道它到底改了什么 | 每次变更是明确的 SQL |
| 不支持列重命名（会先删再加→数据丢失） | 完全掌控 |
| 不支持回滚 | 支持 down 回滚 |
| 团队 review 不了 | SQL 文件可 review |

### 后果

- ✅ 生产变更有审计、可回滚
- ⚠️ 开发体验稍差（GORM model 改完需要手写 SQL），用 make 命令缓解

---

## ADR-004: PO/Domain 分离

**状态**: 已采纳
**日期**: 2026-07-31

### 背景

是否让 domain model 直接带 GORM tag？

### 决策

**分离 PO（Persistent Object）和 Domain**。PO 在 `adapter/mysql/model.go` 中定义，带 GORM tag。Domain 在 `extends/{module}/domain/` 中定义，零框架依赖。

### 原因

```
domain 带 GORM tag → import gorm → 整个领域层和框架耦合
生产换 ORM → 改 domain → 改 service → 改 handler → 改测试
```

分离后换 ORM 只需改 adapter 层，domain/service/handler 不受影响。

### 后果

- ✅ domain 纯净，可独立单测
- ✅ 换 ORM 成本局限在 adapter 层
- ⚠️ 多了 PO ↔ Domain 转换代码（`toDomain()` / `toPO()`），但这是值得的

---

## ADR-005: Zustand 而非 Redux Toolkit

**状态**: 已采纳
**日期**: 2026-07-31

### 背景

React 状态管理方案选择。

### 决策

**Zustand v4**。

### 原因

| | Redux Toolkit | Zustand |
|------|--------------|---------|
| Boilerplate | slice + reducer + action + selector | 一个函数 |
| TypeScript | 需要额外类型推导 | 原生支持 |
| Bundle 大小 | ~12KB | ~2KB |
| 学习曲线 | 中等 | 低 |
| 生态 | 最大 | 快速增长 |

管理后台状态复杂度不高（token、userInfo、menuTree、permissions），Zustand 完全够用。

### 后果

- ✅ 代码量少，store 定义干净
- ⚠️ DevTools 不如 Redux 成熟（可忽略）
