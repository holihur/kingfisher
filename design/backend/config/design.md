# Config — 配置管理

## 职责

多环境配置加载、校验、热重载。配置不可写在代码里，按环境分层覆盖。

## 对外接口

```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server" validate:"required"`
    Database DatabaseConfig `mapstructure:"database" validate:"required"`
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
type DatabaseConfig struct {
    Driver   string          `mapstructure:"driver" validate:"required,oneof=sqlite mysql postgres"`
    SQLite   SQLiteConfig    `mapstructure:"sqlite"`
    MySQL    MySQLConfig     `mapstructure:"mysql"`     // driver=mysql 时必填
    Postgres PostgresConfig  `mapstructure:"postgres"`   // driver=postgres 时必填
}

type SQLiteConfig struct {
    Path string `mapstructure:"path"` // kingfisher.db — 默认为项目根目录
}

type MySQLConfig struct {
    Host     string `mapstructure:"host" validate:"required,hostname|ip"`
    Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
    User     string `mapstructure:"user" validate:"required"`
    Password string `mapstructure:"password" validate:"required"`
    Database string `mapstructure:"database" validate:"required"`
    MaxIdle  int    `mapstructure:"max_idle_conns" validate:"min=1,max=100"`
    MaxOpen  int    `mapstructure:"max_open_conns" validate:"min=1,max=500"`
    MaxLifetime string `mapstructure:"conn_max_lifetime"` // 1h
}

type PostgresConfig struct {
    Host     string `mapstructure:"host" validate:"required,hostname|ip"`
    Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
    User     string `mapstructure:"user" validate:"required"`
    Password string `mapstructure:"password" validate:"required"`
    Database string `mapstructure:"database" validate:"required"`
    SSLMode  string `mapstructure:"sslmode"` // disable|require|verify-full
    MaxIdle  int    `mapstructure:"max_idle_conns"`
    MaxOpen  int    `mapstructure:"max_open_conns"`
}
```

## 敏感信息策略

| 字段 | 来源 | 原因 |
|------|------|------|
| redis.password | 环境变量 | 不能进 git |
| jwt.secret | 环境变量 | 不能进 git |
| database.mysql.password | 环境变量 | 不能进 git |
| database.postgres.password | 环境变量 | 不能进 git |
| database.sqlite.path | config.yaml OK | sqlite 文件路径非敏感 |

## config.yaml 最小示例

```yaml
server:
  port: 8080
  mode: debug
  read_timeout: 10s
  write_timeout: 10s
  max_request_body: 10MB

database:
  driver: sqlite                # sqlite | mysql | postgres
  sqlite:
    path: kingfisher.db          # 开发环境零依赖启动

  # driver=mysql 时使用：
  # mysql:
  #   host: 127.0.0.1
  #   port: 3306
  #   database: kingfisher
  #   charset: utf8mb4
  #   max_idle_conns: 10
  #   max_open_conns: 100
  #   conn_max_lifetime: 1h
  #   # user, password 来自环境变量

  # driver=postgres 时使用：
  # postgres:
  #   host: 127.0.0.1
  #   port: 5432
  #   database: kingfisher
  #   sslmode: disable
  #   # user, password 来自环境变量

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
