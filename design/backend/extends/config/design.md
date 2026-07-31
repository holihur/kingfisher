# Extends/Config — 系统配置管理

## 职责

管理运行时系统配置（键值对），支持前端管理界面修改并实时生效。

## 目录结构

```
extends/config/
├── domain/config.go            # SystemConfig 实体
├── port/repository.go          # ConfigRepository 接口
├── app/service.go              # ConfigService（缓存策略）
├── adapter/mysql/
│   ├── model.go
│   └── repo.go
├── transport/
│   ├── handler.go
│   └── register.go
└── wire.go
```

## Domain

```go
type SystemConfig struct {
    ID        uint      `json:"id"`
    Key       string    `json:"key"`
    Value     string    `json:"value"`       // JSON string 或纯文本
    Remark    string    `json:"remark"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

## 预设配置项

| Key | Value 示例 | 说明 |
|-----|-----------|------|
| site_name | Kingfisher Admin | 系统名称 |
| site_logo | /logo.png | Logo 路径 |
| max_login_attempts | 5 | 最大登录失败次数 |
| lockout_duration | 15m | 锁定时间 |
| session_timeout | 30m | 会话超时 |

## Port

```go
type ConfigRepository interface {
    GetAll(ctx context.Context) ([]domain.SystemConfig, error)
    GetByKey(ctx context.Context, key string) (*domain.SystemConfig, error)
    Set(ctx context.Context, key, value string) error        // upsert
    Delete(ctx context.Context, key string) error
}
```

## Service（带缓存）

```go
type ConfigService struct {
    repo  port.ConfigRepository
    cache coreCache.Cache
}

func (s *ConfigService) GetAll(ctx context.Context) ([]domain.SystemConfig, error)
func (s *ConfigService) Get(ctx context.Context, key string) (*domain.SystemConfig, error)
func (s *ConfigService) Set(ctx context.Context, key, value string) error
func (s *ConfigService) Delete(ctx context.Context, key string) error
```

### GetAll 策略：Cache-Aside

```
1. cache.Get("config:all")
2. miss → repo.GetAll() → 转成 map[string]string
3. data,_:=json.Marshal(configs); cache.Set("config:all", string(data), 5*time.Minute)
4. 返回
```

### Set 策略：Write-Through

```
1. repo.Set(key, value)
2. cache.Delete("config:all")         // 失效全量缓存
3. cache.Delete(fmt.Sprintf("config:%s", key))  // 失效单键缓存
```

## Handler

```go
type ConfigHandler struct { svc *ConfigService }

func (h *ConfigHandler) GetAll(c *gin.Context)     // GET  /api/v1/configs
func (h *ConfigHandler) Get(c *gin.Context)        // GET  /api/v1/configs/:key
func (h *ConfigHandler) Set(c *gin.Context)        // PUT  /api/v1/configs/:key
func (h *ConfigHandler) Delete(c *gin.Context)     // DELETE /api/v1/configs/:key
```

## 路由注册

```go
func (m *Module) RegisterProtected(r *gin.RouterGroup) {
    configs := r.Group("/configs")
    configs.GET("", RequirePerm("config:list"), m.handler.GetAll)
    configs.GET("/:key", RequirePerm("config:list"), m.handler.Get)
    configs.PUT("/:key", RequirePerm("config:update"), m.handler.Set)
    configs.DELETE("/:key", RequirePerm("config:update"), m.handler.Delete)
}
```

## 与其他模块的集成

其他 Service 读取配置时，通过 ConfigService 读取（带缓存）：

```go
// user service 用配置控制登录限制
maxAttempts, _ := configSvc.Get(ctx, "max_login_attempts")
// strconv.Atoi 转换后使用
```

## 设计要点

- 系统配置是键值对，Value 可以是 JSON（前端决定解析方式）
- 写操作后失效缓存，但不阻塞请求（最终一致性可接受）
- 预设配置项在 `migrations/000004_seed_configs.up.sql` 中初始化
