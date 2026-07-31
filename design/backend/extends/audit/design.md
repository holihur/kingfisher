# Extends/Audit — 操作审计日志

## 职责

记录所有写操作（POST/PUT/DELETE），提供查询接口。谁在什么时候做了什么——合规必备。

## 目录结构

```
extends/audit/
├── domain/audit.go              # AuditLog 实体
├── port/repository.go           # AuditRepository 接口
├── app/service.go               # AuditService（异步写入）
├── adapter/mysql/
│   ├── model.go                # GORM PO
│   └── repo.go                 # 实现 port.AuditRepository
├── transport/
│   ├── handler.go              # GET /audit-logs
│   ├── register.go             # Module 注册
│   └── middleware.go            # Audit 中间件（记录写操作）
└── wire.go
```

## Domain

```go
type AuditLog struct {
    ID         uint      `json:"id"`
    UserID     uint      `json:"user_id"`
    Username   string    `json:"username"`
    Action     string    `json:"action"`     // CREATE / UPDATE / DELETE / LOGIN
    Resource   string    `json:"resource"`   // user / menu / role / config
    ResourceID uint      `json:"resource_id"`
    Detail     string    `json:"detail"`     // JSON: {"changed_fields":["email","status"]}
    IP         string    `json:"ip"`
    UserAgent  string    `json:"user_agent"`
    CreatedAt  time.Time `json:"created_at"`
}
```

## Port

```go
type AuditRepository interface {
    InsertBatch(ctx context.Context, logs []domain.AuditLog) error
    FindAll(ctx context.Context, page, pageSize int, filters map[string]any) ([]domain.AuditLog, int64, error)
}
```

## Service（异步写入）

```go
type AuditService struct {
    repo   port.AuditRepository
    buffer chan *domain.AuditLog    // 缓冲通道，容量 1000
}

func NewAuditService(repo port.AuditRepository) *AuditService {
    s := &AuditService{repo: repo, buffer: make(chan *domain.AuditLog, 1000)}
    go s.worker()  // 后台 goroutine 批量写入
    return s
}

func (s *AuditService) Log(ctx context.Context, log *domain.AuditLog) {
    select {
    case s.buffer <- log:       // 非阻塞写入 channel
    default:                     // buffer 满 → 丢弃（不阻塞业务）
        logger.Warn("audit buffer full, dropping log")
    }
}

func (s *AuditService) worker() {
    batch := make([]domain.AuditLog, 0, 50)
    ticker := time.NewTicker(2 * time.Second)
    for {
        select {
        case log := <-s.buffer:
            batch = append(batch, *log)
            if len(batch) >= 50 {
                s.repo.InsertBatch(context.Background(), batch)
                batch = batch[:0]
            }
        case <-ticker.C:
            if len(batch) > 0 {
                s.repo.InsertBatch(context.Background(), batch)
                batch = batch[:0]
            }
        }
    }
}
```

- 异步写入：buffer channel + 批量 insert，业务请求不等待审计日志写入
- 每 2s 或满 50 条刷一次 DB
- Buffer 满时丢弃（审计日志允许少量丢失，不能阻塞主业务）

## Handler

```go
type AuditHandler struct { svc *AuditService }

// GET /api/v1/audit-logs?page=1&page_size=20&user_id=1&resource=user&action=DELETE
func (h *AuditHandler) List(c *gin.Context)
```
仅管理员可查看。不支持修改和删除审计日志（合规要求）。

## Audit 中间件

```go
// transport/middleware.go
func AuditMiddleware(auditSvc *AuditService) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()  // 先执行业务，再审计

        // 只记录写操作 + 成功响应 + 登录事件（安全审计必须记录登录失败）
        if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
            return
        }
        if c.Writer.Status() < 200 || c.Writer.Status() >= 300 {
            return
        }

        resource := extractResource(c.FullPath())      // "/api/v1/users/:id" → "user"
        resourceID := extractResourceID(c)             // 从 path param 或 url 提取

        auditSvc.Log(c.Request.Context(), &domain.AuditLog{
            UserID:     c.GetUint("user_id"),
            Username:   c.GetString("username"),
            Action:     c.Request.Method,
            Resource:   resource,
            ResourceID: resourceID,
            IP:         c.ClientIP(),
            UserAgent:  c.Request.UserAgent(),
        })
    }
}

// path → resource 映射
func extractResource(path string) string {
    // "/api/v1/users/:id" → "user"
    // "/api/v1/menus/tree" → "menu"
    // "/api/v1/roles/:id/permissions" → "role"
    parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
    if len(parts) > 0 { return parts[0] }
    return "unknown"
}
```

## 路由注册

```go
// transport/register.go
func (m *AuditModule) Name() string { return "audit" }
func (m *AuditModule) Init(ctx context.Context) error {
    // worker 已在 NewAuditService 中启动，此处无额外操作
    return nil
}
func (m *AuditModule) RegisterPublic(r *gin.RouterGroup)  {}
func (m *AuditModule) RegisterProtected(r *gin.RouterGroup) {
    r.GET("/audit-logs", RequirePerm("audit:list"), m.handler.List)
}
func (m *AuditModule) Shutdown(ctx context.Context) error {
    // flush 剩余 buffer
    m.svc.Flush()
    return nil
}
```

## 中间件注册

Audit 中间件通过 Module.Init() 添加到所有写操作路由中，而非全局 Use（避免审计 /health /metrics）：

```go
// 在 router.go 中为需要审计的路由组挂载
// 方式 A: 在 RegisterProtected 中每个写路由组手动加
protected := r.Group("")
protected.Use(Auth(jwtMgr), RBAC(roleSvc), AuditMiddleware(auditSvc))
```

## 工厂函数

```go
func NewModule(db *gorm.DB) core.Module {
    repo := adapter.NewAuditRepo(db)
    svc := app.NewAuditService(repo)
    handler := transport.NewAuditHandler(svc)
    return &transport.AuditModule{handler: handler, svc: svc}
}
```

## 迁移 SQL

```sql
-- migrations/000010_create_audit_logs.up.sql
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

```sql
-- migrations/000010_create_audit_logs.down.sql
DROP TABLE IF EXISTS audit_logs;
```

## 预设权限

追加到 `migrations/000008_seed_data.up.sql`：

```sql
INSERT INTO permissions (id, name, code, resource, action) VALUES
(15, '查看审计日志', 'audit:list', 'audit', 'read');
INSERT INTO role_permissions (role_id, permission_id) VALUES (1, 15); -- admin
```

## 性能考量

- 异步写入：业务请求 P50 < 1ms 增加（纯 channel write）
- 批量 insert：50 条一次 `INSERT INTO ... VALUES (...),(...),...` 
- Buffer 满丢弃：日志 warn，不被 DDoS 用的审计日志打满内存
- 审计日志表按月份分区（v1 不做，数据量大后再加）
