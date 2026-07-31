# Bootstrap — 项目初始化 & 启动流程

## 职责

定义从零到可运行项目的完整初始化步骤。新成员/新项目照着走，5 分钟跑起来。

---

## 1. 项目初始化

### 1.1 创建项目

```bash
# 方式 A：clone 脚手架
git clone https://github.com/example/kingfisher.git my-project
cd my-project
rm -rf .git && git init && git add -A && git commit -m "init: from kingfisher scaffold"

# 方式 B：go mod 新建（不用脚手架）
mkdir my-project && cd my-project
go mod init my-project
```

### 1.2 安装依赖

```bash
# 安装 Go 工具链
make setup

# 等价于：
go mod tidy
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/google/wire/cmd/wire@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 1.3 配置环境

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env，填入实际值
# MYSQL_PASSWORD=xxx
# REDIS_PASSWORD=xxx
# JWT_SECRET=xxx
```

---

## 2. 启动顺序

### 2.1 启动依赖服务

```bash
# 方式 A：Docker（推荐）
docker-compose -f deploy/docker-compose.dev.yaml up -d mysql redis
# 等待 MySQL 就绪
until docker-compose -f deploy/docker-compose.dev.yaml exec mysql mysqladmin ping -h localhost; do sleep 2; done

# 方式 B：本地安装
# MySQL 8.0 + Redis 7，手动启动
```

### 2.2 执行数据库迁移

```bash
make migrate-up
# 等价于：go run ./cmd/migrate up
# 输出：applied 8 migrations
```

### 2.3 验证迁移结果

```bash
mysql -u root -p kingfisher -e "SHOW TABLES;"
# users, roles, permissions, role_permissions, menus, role_menus, system_configs

mysql -u root -p kingfisher -e "SELECT id, username, email FROM users;"
# 1 | admin | admin@example.com
```

### 2.4 启动应用

```bash
# 开发模式
make run
# 输出：
# [INFO] config loaded from config/config.yaml
# [INFO] mysql connected (127.0.0.1:3306/kingfisher)
# [INFO] redis connected (127.0.0.1:6379)
# [INFO] server starting on :8080
# [INFO] visit http://localhost:8080/swagger/index.html
```

### 2.5 验证启动

```bash
# 健康检查
curl http://localhost:8080/health
# {"status":"ok"}

# 就绪检查
curl http://localhost:8080/ready
# {"status":"ready","mysql":"ok","redis":"ok"}

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Abcd1234"}'
# {"code":0,"message":"success","data":{"access_token":"eyJ...","refresh_token":"eyJ...","user":{...}}}
```

---

## 3. 初始化流程图

```
┌─────────────────────────────────────────────────┐
│                  main.go 启动                    │
├─────────────────────────────────────────────────┤
│                                                  │
│  1. config.Load("config/config.yaml")           │
│     ├── 读取 config.yaml                         │
│     ├── 自动映射环境变量 (MYSQL_HOST → mysql.host) │
│     ├── Validate() 校验必填项                     │
│     └── 若 JWT_SECRET 为空 → Fatal              │
│                                                  │
│  2. logger.New(cfg.Log)                         │
│     ├── Debug 模式 → console 编码 + 颜色          │
│     └── Release 模式 → JSON + 文件 + 滚动         │
│                                                  │
│  3. database.NewGorm(cfg.MySQL, logger)          │
│     ├── 拼接 DSN                                 │
│     ├── gorm.Open + 设置连接池                     │
│     ├── db.Ping() 验证连接                        │
│     └── 失败 → 重试 3 次（间隔 2s），仍失败 → Fatal │
│                                                  │
│  4. 开发环境：database.RunMigrations(db, "migrations") │
│     └── 生产环境：跳过，由运维手动执行              │
│                                                  │
│  5. cache.NewRedis(cfg.Redis)                    │
│     ├── redis.NewClient                          │
│     ├── rdb.Ping(ctx)                            │
│     └── 失败 → Fatal                             │
│                                                  │
│  6. jwt.NewJWTManager(cfg.JWT, cache)            │
│     └── sync.Once 初始化 secret + TTL            │
│                                                  │
│  7. telemetry.InitTracer(cfg.Telemetry)          │
│     └── (可选，配置了 endpoint 才启用)             │
│                                                  │
│  8. 组装 App                                     │
│     ├── Core 中间件链                             │
│     ├── extends 模块注册                          │
│     └── 路由挂载                                 │
│                                                  │
│  9. http.Server.ListenAndServe()                 │
│     └── goroutine 中启动                         │
│                                                  │
│ 10. signal.Notify(quit, SIGINT, SIGTERM)         │
│     └── 等待退出信号                              │
│                                                  │
│ 11. srv.Shutdown(ctx)                            │
│     ├── 30s 超时                                 │
│     ├── 停止接收新连接                            │
│     └── 等待进行中的请求完成                      │
│                                                  │
│ 12. 资源关闭                                     │
│     ├── telemetry.Shutdown()                     │
│     ├── rdb.Close()                              │
│     ├── sqlDB.Close()                            │
│     └── logger.Sync()                            │
│                                                  │
└─────────────────────────────────────────────────┘
```

---

## 4. 配置文件加载优先级

```
环境变量 (最高)      JWT_SECRET=prod-secret       ← K8s Secret / .env
     ↓ 覆盖
config.prod.yaml     生产环境特定覆盖                ← 不在 git 中
     ↓ 覆盖
config.dev.yaml      开发环境覆盖（本地端口等）       ← 可提交
     ↓ 覆盖
config/config.yaml   默认值（基础配置）               ← 提交 git
```

**合并规则**：
1. Viper 先读 `config.yaml`
2. 若存在 `config.{env}.yaml`（`{env}` 来自环境变量 `APP_ENV`），merge 覆盖
3. 环境变量覆盖同名字段（`.` 变 `_`，如 `MYSQL_HOST` → `mysql.host`）

---

## 5. 启动失败处理

| 阶段 | 失败 | 行为 |
|------|------|------|
| Config Load | 文件不存在 / 格式错误 | `log.Fatal` 立即退出 |
| Config Validate | 必填项为空 | `log.Fatal` 立即退出 |
| MySQL 连接 | 第一次失败 | 重试 3 次（间隔 2s） |
| MySQL 连接 | 3 次后仍失败 | `log.Fatal` 退出 |
| Redis 连接 | 失败 | `log.Fatal` 退出（Redis 是核心依赖） |
| Migration | 失败 | `log.Fatal` 退出 |
| HTTP Server | 端口被占用 | `log.Fatal` 退出 |

> Fatal 退出 + Docker `restart: unless-stopped` = 自动重试启动。

---

## 6. 开发环境快速启动（Makefile 入口）

```bash
# 第一次
make setup          # 装依赖 + 生成 swagger + wire
make docker-dev     # 启动 MySQL + Redis
make migrate-up     # 建表 + 种子数据
make run            # 启动 Go

# 之后每天
make run            # 改了代码 → air 自动重启
```

---

## 7. 新增 extends 模块的步骤

```bash
# 1. 创建模块目录
mkdir -p extends/payment/{domain,port,adapter/mysql,app,transport}

# 2. 定义 domain
# extends/payment/domain/order.go

# 3. 定义 port 接口
# extends/payment/port/repository.go

# 4. 实现 adapter
# extends/payment/adapter/mysql/model.go
# extends/payment/adapter/mysql/repo.go

# 5. 实现 service
# extends/payment/app/service.go

# 6. 实现 handler + 注册路由
# extends/payment/transport/handler.go
# extends/payment/transport/register.go   ← 实现 core.Module 接口

# 7. 在 main.go 中注册
# coreRouter.RegisterModule(r, payment.NewModule(db, cache, jwtMgr), authMw, rbacMw)

# 8. 更新 Wire
# extends/payment/wire.go  ← 新增 Provider Set

# 9. 生成 Wire + Swagger
make wire && make swagger
```

---

## 8. 依赖图

```
cmd/server/main.go
 ├── core/config          ← config.yaml
 ├── core/logger          ← config.Log
 ├── core/database        ← config.MySQL
 │   └── migrations/      ← SQL 文件
 ├── core/cache           ← config.Redis
 ├── core/jwt             ← config.JWT + core/cache
 ├── core/telemetry       ← config.Telemetry (可选)
 ├── core/middleware      ← core/logger + core/cache + core/jwt
 ├── core/router          ← core/middleware
 └── extends/*            ← core/database + core/cache + core/jwt
     ├── user
     ├── menu
     ├── rbac
     └── config
```

---

## 9. 验证清单（启动后必查）

| 检查项 | 命令 | 预期 |
|--------|------|------|
| 进程存活 | `curl localhost:8080/health` | `{"status":"ok"}` |
| DB 连通 | `curl localhost:8080/ready` | `{"status":"ready","mysql":"ok","redis":"ok"}` |
| Swagger | 浏览器 `/swagger/index.html` | 可见所有接口 |
| 登录 | `curl POST /api/v1/auth/login ...` | 返回 token |
| Metric | `curl localhost:8080/metrics` | 返回 Prometheus 指标 |
| Trace | Jaeger `:16686` | 有 trace 数据 |
