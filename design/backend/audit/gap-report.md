# Gap Report — 深度排查缺失项

> 52 个设计文档 vs 管理后台生产必备能力，逐项排查
> 排查日期：2026-07-31 | 文档版本 v1

## 统计

| 级别 | 数量 | 定义 | 阻塞里程碑 |
|------|------|------|-----------|
| P0 | 7 | v1 必须补齐，否则到不了生产 | M1/M2/M4/M7 |
| P1 | 7 | 首版发布后第一个迭代补 | 不阻塞 v1 |
| P2 | 9 | 已知限制，明确不做或日后处理 | — |

---

## P0 · v1 必须补齐

### 1. `PUT /api/v1/users/me/password` — 修改自己密码

| 属性 | 值 |
|------|-----|
| 阻塞 | **M2** — 用户模块不完整 |
| 影响 | `api-contract` 列出了此接口，但 `extends/user` 无对应实现 |
| 工作量 | 小 |

**修改文件**：

```
extends/user/port/service.go         + ChangePassword(ctx, userID uint, oldPwd, newPwd string) error
extends/user/app/service.go          + 实现：查用户→校验 oldPwd→bcrypt 新密码→repo.Update
extends/user/transport/handler.go    + ChangePassword handler
extends/user/transport/register.go   + registered protected: PUT /users/me/password
```

**核心逻辑**：
```
1. c.Get("user_id") → 当前登录用户
2. c.ShouldBindJSON(&{OldPassword, NewPassword})
3. svc.ChangePassword(userID, oldPwd, newPwd)
4. 成功 → 200, message: "密码修改成功，请重新登录"
5. 旧密码错误 → 10103
```

**安全考量**：修改成功后不做即时登录态变更；前端提示"请重新登录"。

---

### 2. 操作审计日志 — 零覆盖

| 属性 | 值 |
|------|-----|
| 阻塞 | **M4** — 后端 API 合规 |
| 影响 | 无法回答"谁在什么时候删了哪个用户" |
| 工作量 | 中 |

**新增 `extends/audit/` 模块**：

```
extends/audit/
├── domain/audit.go              AuditLog 实体
├── port/repository.go           AuditRepository 接口
├── adapter/mysql/
│   ├── model.go                 GORM PO
│   └── repo.go                  实现
├── app/service.go               AuditService
├── transport/
│   ├── handler.go               GET /audit-logs (查询)
│   ├── middleware.go            Audit 中间件
│   └── register.go              Module 注册
└── wire.go
```

**`domain/audit.go`**：
```go
type AuditLog struct {
    ID         uint      `json:"id"`
    UserID     uint      `json:"user_id"`
    Username   string    `json:"username"`
    Action     string    `json:"action"`     // CREATE / UPDATE / DELETE / LOGIN
    Resource   string    `json:"resource"`   // user / menu / role / config
    ResourceID uint      `json:"resource_id"`
    Detail     string    `json:"detail"`     // JSON: {"old":..., "new":..., "changed_fields":["email"]}
    IP         string    `json:"ip"`
    UserAgent  string    `json:"user_agent"`
    CreatedAt  time.Time `json:"created_at"`
}
```

**迁移 SQL** (`migrations/000009_create_audit_logs.up.sql`):
```sql
CREATE TABLE audit_logs (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id     BIGINT UNSIGNED NOT NULL,
    username    VARCHAR(32)  NOT NULL,
    action      VARCHAR(16)  NOT NULL,
    resource    VARCHAR(32)  NOT NULL,
    resource_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    detail      JSON,
    ip          VARCHAR(45)  NOT NULL DEFAULT '',
    user_agent  VARCHAR(512) NOT NULL DEFAULT '',
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_user_id (user_id),
    INDEX idx_resource (resource, resource_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Audit 中间件** (`extends/audit/transport/middleware.go`)：
```go
func AuditMiddleware(auditSvc port.AuditService) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()  // 先执行业务

        // 只记录写操作
        if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
            return
        }
        if c.Writer.Status() >= 400 {
            return  // 只记成功的写操作（失败的由 error log 覆盖）
        }

        userID := c.GetUint("user_id")
        resource := extractResource(c.FullPath())     // "/api/v1/users/:id" → "user"
        resourceID := extractResourceID(c)            // 从 path param 或 response body 提取

        auditSvc.Log(c.Request.Context(), &domain.AuditLog{
            UserID: userID, Action: c.Request.Method, Resource: resource,
            ResourceID: resourceID, IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
        })
    }
}
```

**Handler** (`extends/audit/transport/handler.go`)：
```
GET /api/v1/audit-logs?page=1&page_size=20&user_id=1&resource=user&action=DELETE
```
仅管理员可查看。不支持删除审计日志（合规要求）。

**注册** (`extends/audit/transport/register.go`)：
```go
func (m *AuditModule) RegisterProtected(r *gin.RouterGroup) {
    r.GET("/audit-logs", RequirePerm("audit:list"), m.handler.List)
}
// Init 中挂载 Audit 中间件到所有 protected 路由
```

**性能考量**：审计日志用异步写入（buffered channel + batch insert）。单条写入 P50 < 1ms，不阻塞业务请求。

---

### 3. 强制踢人 / Session 管理

| 属性 | 值 |
|------|-----|
| 阻塞 | **M4** — 安全合规 |
| 影响 | Admin 无法撤销被盗账号的活跃 session |
| 工作量 | 小 |

**修改文件**：

```
extends/rbac/transport/auth_middleware.go   + ParseToken 时额外检查：该 user 的 session_version
core/jwt/jwt.go                             + GenerateToken 时注入 session_version claim
extends/user/port/service.go                + RevokeUserSessions
extends/user/app/service.go                 + 实现
extends/user/transport/handler.go           + DELETE /users/:id/sessions
```

**Session 版本号方案**（优于黑名单所有 JTI）：

1. `users` 表加 `session_version INT DEFAULT 1`
2. JWT payload 加 `sv INT`（session_version）
3. Revoke 时 `UPDATE users SET session_version = session_version + 1 WHERE id = ?`
4. Auth 中间件校验：`claims.SessionVersion == user.SessionVersion`（从 Redis 缓存读，miss 回 DB）
5. 不等 → 401

**优点**：不需要查 Redis 黑名单，一次 Redis `GET user:sv:{userID}` 即可。

**修改文件清单**：
```
migrations/000009_alter_users_add_session_version.up.sql  + ALTER TABLE users ADD COLUMN session_version INT DEFAULT 1
extends/user/domain/user.go                              + SessionVersion int
extends/user/adapter/mysql/model.go                      + SessionVersion
extends/user/transport/handler.go                        + RevokeSessions handler
extends/user/transport/register.go                       + 注册路由
```

---

### 4. Module 生命周期钩子

| 属性 | 值 |
|------|-----|
| 阻塞 | **M1** — core 接口不完整 |
| 影响 | 模块无 Init/Shutdown，影响审计中间件注册、缓存预热、后台 goroutine 管理 |
| 工作量 | 小 |

**修改文件**：

```
core/router/engine.go   Module 接口加 Init/Shutdown
cmd/server/main.go      启动/关闭时遍历 modules 调用 Init/Shutdown
```

**新接口**：
```go
type Module interface {
    Name() string
    Init(ctx context.Context) error
    RegisterPublic(r *gin.RouterGroup)
    RegisterProtected(r *gin.RouterGroup)
    Shutdown(ctx context.Context) error
}
```

**main.go 调用**：
```go
// 初始化
for _, m := range app.Modules {
    if err := m.Init(ctx); err != nil {
        logger.Fatal("module init failed", zap.String("module", m.Name()), zap.Error(err))
    }
}
// 关闭（倒序）
for i := len(app.Modules) - 1; i >= 0; i-- {
    if err := app.Modules[i].Shutdown(ctx); err != nil {
        logger.Error("module shutdown error", zap.String("module", app.Modules[i].Name()), zap.Error(err))
    }
}
```

**各模块 Init 用途**：

| 模块 | Init | Shutdown |
|------|------|----------|
| audit | 启动异步写入 goroutine | flush buffer → 关闭 channel |
| rbac | 预热权限缓存到 Redis | — |
| menu | 预热菜单树缓存 | — |

---

### 5. `/version` 端点

| 属性 | 值 |
|------|-----|
| 阻塞 | **M7** — 生产排查必须 |
| 影响 | 出问题时无法确认部署了哪个版本 |
| 工作量 | 极小 |

**实现**：

```go
// core/router/engine.go — 在 NewEngine 中注册
r.GET("/version", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "version":    version,     // -ldflags -X main.version=$(git describe --tags)
        "commit":     commit,      // -ldflags -X main.commit=$(git rev-parse --short HEAD)
        "build_time": buildTime,   // -ldflags -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)
        "go_version": runtime.Version(),
    })
})
```

```
// Makefile
LDFLAGS = -s -w \
  -X main.version=$(shell git describe --tags --always 2>/dev/null || echo "dev") \
  -X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
  -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	go build -ldflags="$(LDFLAGS)" -o bin/server ./cmd/server
```

**响应示例**：
```json
{"version":"v1.0.3","commit":"a3f2b1c","build_time":"2026-08-01T10:30:00Z","go_version":"go1.23.4"}
```

---

### 6. Graceful Degradation 策略集中文档

| 属性 | 值 |
|------|-----|
| 阻塞 | **M7** — 运维决策 |
| 影响 | 依赖不可用时的系统行为散落在 acceptance Chaos 验收中，无集中决策 |
| 工作量 | 极小（纯文档，无代码） |

**依赖不可用行为矩阵**：

| 依赖 | 状态 | `/health` | `/ready` | 公开接口 | 需要 RBAC 的接口 | 缓存命中读 | JWT 黑名单检查 |
|------|------|-----------|----------|----------|-----------------|-----------|---------------|
| MySQL | down | 200 | 503 | 登录 500 | 500 | — | — |
| MySQL | 慢 (>3s) | 200 | 200 | 等待/超时 500 | 等待/超时 500 | — | — |
| Redis | down | 200 | 503 | 登录 200(限流失效) | 缓存 miss→回源 DB, RBAC 拒绝 503 | miss→回源 DB | 失败→拒绝 503 |
| Redis | 慢 (>1s) | 200 | 200 | 200(略慢) | RBAC 拒绝 503 | miss→回源 DB | 超时→拒绝 503 |
| Jaeger | down | 200 | 200 | 200(otel 静默) | 200 | 200 | 200 |
| 所有依赖 | 正常 | 200 | 200 | 200 | 200 | 200 | 200 |

**降级决策原则**：

1. **安全优先于可用性** — Redis 不可用时 RBAC 拒绝全部请求（不降级放行）
2. **公开接口不依赖 Redis** — 登录不受 Redis 影响（限流失效是 acceptable degradation）
3. **可观测性不阻断业务** — Jaeger 写失败不影响请求，SDK 内部静默处理
4. **缓存 miss 可回源 DB** — 性能下降但不 500
5. **JWT 黑名单 Redis 不可达 → 拒绝** — 不冒已注销 token 放行的风险

---

### 7. Gzip 压缩中间件

| 属性 | 值 |
|------|-----|
| 阻塞 | **M1** — core middleware 不完整 |
| 影响 | JSON 响应无压缩，带宽浪费 70%+ |
| 工作量 | 极小 |

**实现**：

```go
// core/middleware/gzip.go
import "github.com/gin-contrib/gzip"

func Gzip() gin.HandlerFunc {
    return gzip.Gzip(
        gzip.DefaultCompression,  // level 6
        gzip.WithExcludedPaths([]string{"/metrics", "/health"}),
        gzip.WithExcludedExtensions([]string{".png", ".jpg", ".gif"}),
    )
}
```

**注册位置** (`core/router/engine.go`)：
```go
r.Use(middleware.Gzip())  // 排在 Logger 之后、Handler 之前
```

**中间件链更新**：
```
RequestID → Recovery → Trace → Logger → Gzip → CORS → Auth/RBAC/RequirePerm → RateLimit → Handler
```

---

## P1 · 首版发布后第一个迭代补

### 8. `POST /api/v1/users/:id/unlock` — 管理员手动解锁

**修复位置**：`extends/user/transport/handler.go` `extends/user/app/service.go` `extends/user/port/service.go`

**Service**：
```go
func (s *UserService) Unlock(ctx context.Context, userID uint) error {
    // Redis key: login_fail:{username}
    user, err := s.repo.FindByID(ctx, userID)
    if err != nil { return err }
    return s.cache.Delete(ctx, "login_fail:"+user.Username)
}
```

**Handler**：`POST /api/v1/users/:id/unlock`，仅管理员 + `RequirePerm("user:update")`

### 9. 登录失败计数 TTL 到期后清零

**确认行为**（不改代码，仅文档化）：
- Redis key `login_fail:{username}` TTL = `lockout_duration`（配置项，默认 15min）
- TTL 到期 → Redis 自动删除 key → 下次登录失败时 `INCR` 从 1 开始
- 不需要主动清零逻辑，Redis TTL 即清零机制

### 10. Test fixture 工厂

**新建** `test/testutil/fixture.go`：

```go
package testutil

import (
    "testing"
    "kingfisher/extends/user/domain"
    "kingfisher/extends/rbac/domain"
)

func NewTestUser(t *testing.T, db *gorm.DB, username string) *domain.User {
    t.Helper()
    hashed, _ := bcrypt.GenerateFromPassword([]byte("Abcd1234"), bcrypt.MinCost)
    u := &domain.User{Username: username, Password: string(hashed), Email: username + "@test.com", Status: 1}
    require.NoError(t, db.Create(u).Error)
    return u
}

func NewTestRole(t *testing.T, db *gorm.DB, name, code string) *domain.Role {
    t.Helper()
    r := &domain.Role{Name: name, Code: code, Status: 1}
    require.NoError(t, db.Create(r).Error)
    return r
}

func NewTestMenu(t *testing.T, db *gorm.DB, parentID uint, name, path string, menuType int) *domain.Menu {
    t.Helper()
    m := &domain.Menu{ParentID: parentID, Name: name, Path: path, Type: menuType, Status: 1}
    require.NoError(t, db.Create(m).Error)
    return m
}
```

### 11. Mock 自动生成

**port 文件** 加 `//go:generate` 注释：

```go
// extends/user/port/repository.go
//go:generate mockgen -destination=../../../test/testutil/mock_user_repo.go -package=testutil kingfisher/extends/user/port UserRepository

// extends/user/port/service.go
//go:generate mockgen -destination=../../../test/testutil/mock_user_service.go -package=testutil kingfisher/extends/user/port AuthService,UserService
```

**Makefile**：
```makefile
gen-mocks: ## 生成 mock 文件
	go generate ./extends/...
```

### 12. API 响应 JSON Schema 校验 ⏸️ v2

**明确不做**。v1 只做 HTTP status + 关键字段断言（`assert.Equal(t, 0, resp.Code)`）。结构化 JSON Schema 校验需要额外的 schema 定义文件，量大于实际收益（管理后台接口数量 < 25 个）。

### 13. 幂等性 `Idempotency-Key` ⏸️ v2

**明确不做**。管理后台的 POST/PUT 操作不涉及金额或库存，重复提交（网络超时后用户手动重试）的后果是重复创建一条记录或重复更新，业务上可接受。

### 14. 请求超时按路由分级 ⏸️ v2

**当前策略**：全局 `ReadTimeout: 10s` + `WriteTimeout: 10s`，所有路由相同。v1 不按路由分级——路由分级超时需要 middleware per-route timeout，而 Gin 的 `http.Server` 超时是全局的，需要额外实现 `context.WithTimeout` wrapper。量大于收益。

---

## P2 · 已知限制

| # | 缺失项 | 明确原因 | 替代方案 |
|---|--------|----------|----------|
| 15 | 忘记密码/重置密码 | v1 无邮件/SMS 服务 | 管理员手动改密码 |
| 16 | 批量删除/导入/导出 | 管理后台用户量 < 1000 | 逐条操作，2 分钟可完成 |
| 17 | 恢复已删除用户（undelete） | 软删除保留数据 | DBA 手动 `UPDATE users SET deleted_at=NULL` |
| 18 | 用户级权限覆盖（不通过角色） | 增加 RBAC 模型复杂度 | 为用户单独建角色 |
| 19 | 权限继承（父子权限） | 管理后台权限粒度粗 | 14 个平级权限等于 4 个资源 × CRUD |
| 20 | 密码历史（防重复使用） | 需要密码变更记录表 | 密码强度策略兜底 |
| 21 | 配置热重载通知 | Docker 重启 < 5s | 改配置 → 重启 → 即生效 |
| 22 | 暗黑模式 | UI 工程量 > 收益 | AntD 5 支持 ConfigProvider theme，后续版本可加 |
| 23 | i18n 国际化 | v1 面向中文团队 | errcode msg 在 map 中集中管理，后续翻译只需改 map |

---

## 对里程碑的影响

| 里程碑 | P0 阻塞 | 说明 |
|--------|---------|------|
| M1 Core | ~~#4 Module~~ ~~#7 Gzip~~ | ✅ 已修复：Module 接口已加 Init/Shutdown，middleware 已加 Gzip |
| M2 用户 | ~~#1 修改密码~~ | ✅ 已修复：extends/user 已补 ChangePassword + RevokeSessions |
| M3 RBAC | 无 | — |
| M4 API | ~~#2 审计日志~~ ~~#3 强制踢人~~ | ✅ 已修复：新增 extends/audit；extends/user 已加 session_version 踢人 |
| M5 前端 | 无 | 审计日志查看页面在 M6 前端页面中实现 |
| M6 CRUD | 无 | — |
| M7 生产 | ~~#5 /version~~ | ✅ 已修复：startup 已加 /version 端点 + ldflags；#6 降级矩阵在 gap report 中 |

---

## Gap 关闭追踪

| # | P | 项 | 状态 | 关闭日期 |
|---|----|----|------|----------|
| 1 | P0 | 修改自己密码 | ✅ 已修复 | 2026-07-31 |
| 2 | P0 | 操作审计日志 | ✅ 已修复 | 2026-07-31 |
| 3 | P0 | 强制踢人 | ✅ 已修复 | 2026-07-31 |
| 4 | P0 | Module 生命周期 | ✅ 已修复 | 2026-07-31 |
| 5 | P0 | /version 端点 | ✅ 已修复 | 2026-07-31 |
| 6 | P0 | Graceful Degradation 矩阵 | ✅ 已修复 | 2026-07-31 |
| 7 | P0 | Gzip 中间件 | ✅ 已修复 | 2026-07-31 |
| 8 | P1 | 账号解锁 | ✅ 已设计 | 2026-07-31 |
| 9 | P1 | 登录失败计数清零确认 | ✅ 已确认 | 2026-07-31 |
| 10 | P1 | Test fixture 工厂 | ✅ 已设计 | 2026-07-31 |
| 11 | P1 | Mock 自动生成 | ✅ 已设计 | 2026-07-31 |
| 12 | P1 | JSON Schema 校验 | ⏸️ v2 | — |
| 13 | P1 | 幂等性 | ⏸️ v2 | — |
| 14 | P1 | 路由分级超时 | ⏸️ v2 | — |
| 15-23 | P2 | 已知限制 | ✅ 已记录 | 2026-07-31 |

---

## Security Fixes · 安全漏洞专项修复

> 排查日期：2026-07-31 | 11 项安全问题全部关闭

| # | 级别 | 问题 | 修复 |
|---|------|------|------|
| 1 | 🔴 | 安全响应头缺失 | ✅ 新增 SecurityHeaders 中间件：X-Frame-Options、CSP、X-Content-Type-Options、X-XSS-Protection、Referrer-Policy |
| 2 | 🔴 | RateLimit IP 伪造 | ✅ Gin `SetTrustedProxies` + config `server.trusted_proxies` |
| 3 | 🔴 | 注册无限流 | ✅ extends/user RegisterPublic 加 `RateLimit(2, 5min)` |
| 4 | 🔴 | 角色无层级——低权限可自提权 | ✅ roles 表 `level` 字段；Update 校验调用者 level ≥ 目标 level |
| 5 | 🔴 | 安全事件无审计 | ✅ Audit 中间件记录 LOGIN/LOGIN_FAIL；审计日志不可删除 |
| 6 | 🟡 | Refresh Token 无 rotation | ✅ ADR 文档化决策——v1 不做，改密码/踢人替代 |
| 7 | 🟡 | 用户枚举（10102 vs 10103） | ✅ Login 统一返回 10103，用 dummy hash 防时间侧信道 |
| 8 | 🟡 | Config value 无类型校验 | ✅ `typedConfigs` map + `validateType` |
| 9 | 🟢 | 改密码后旧 token 有效 | ✅ ChangePassword 末尾自动 `IncrementSessionVersion` |
| 10 | 🟢 | 无双因素认证 | ✅ 已知限制，P2 |
| 11 | 🟢 | 种子默认密码 | ✅ Migration 文档已警告"生产部署后立即修改密码" |
