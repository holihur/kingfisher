# Config — 配置管理

## 职责

多环境配置加载、校验、热重载。配置不可写在代码里，按环境分层覆盖。

## 对外接口

```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server" validate:"required"`
    MySQL    MySQLConfig    `mapstructure:"mysql" validate:"required"`
    Redis    RedisConfig    `mapstructure:"redis" validate:"required"`
    JWT      JWTConfig      `mapstructure:"jwt" validate:"required"`
    Log      LogConfig      `mapstructure:"log"`
    RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

func Load(configPath string) (*Config, error)
func (c *Config) Validate() error
```

## 分层覆盖规则

```
config.yaml          ← 默认值（可提交 git）
config.dev.yaml      ← 开发覆盖（可提交）
config.prod.yaml     ← 生产覆盖（不提交，CI 注入）
环境变量             ← 最高优先级（MYSQL_HOST → mysql.host）
```

## 配置校验（validate tag）

```go
type MySQLConfig struct {
    Host     string `mapstructure:"host" validate:"required,hostname|ip"`
    Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
    User     string `mapstructure:"user" validate:"required"`
    Password string `mapstructure:"password" validate:"required"`
    Database string `mapstructure:"database" validate:"required"`
    MaxIdle  int    `mapstructure:"max_idle_conns" validate:"min=1,max=100"`
    MaxOpen  int    `mapstructure:"max_open_conns" validate:"min=1,max=500"`
}
```

## 敏感信息策略

| 字段 | 来源 | 原因 |
|------|------|------|
| mysql.password | 环境变量 | 不能进 git |
| redis.password | 环境变量 | 不能进 git |
| jwt.secret | 环境变量 | 不能进 git |
| server.port | config.yaml OK | 非敏感 |

## config.yaml 最小示例

```yaml
server:
  port: 8080
  mode: debug
  read_timeout: 10s
  write_timeout: 10s
  max_request_body: 10MB

mysql:
  host: 127.0.0.1
  port: 3306
  database: kingfisher
  charset: utf8mb4
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 1h
  # user, password 来自环境变量

redis:
  host: 127.0.0.1
  port: 6379
  db: 0
  pool_size: 20
  # password 来自环境变量

jwt:
  access_ttl: 2h
  refresh_ttl: 168h
  issuer: kingfisher
  # secret 来自环境变量

log:
  level: info           # debug|info|warn|error
  format: json          # json|console
  output: stdout        # stdout|file
  file_path: logs/app.log
  max_size: 100         # MB，自动滚动
  max_backups: 10
  max_age: 30           # days

rate_limit:
  enabled: true
  requests_per_minute: 60
  login_per_minute: 5   # 登录接口单独限制
```
