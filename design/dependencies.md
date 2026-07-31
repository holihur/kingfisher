# Dependencies — 三方依赖清单

> 41 packages。一份文件，`go mod tidy && npm install` 即可就绪。
> 更新：2026-07-31

## Go 后端 (20 packages)

### Core — `go get`

| Package | 版本 | 用途 |
|------|------|------|
| `github.com/gin-gonic/gin` | v1.10+ | HTTP Router |
| `github.com/gin-contrib/gzip` | v1+ | Gzip 中间件 |
| `gorm.io/gorm` | v1.25+ | ORM |
| `gorm.io/driver/sqlite` | v1.6+ | SQLite 驱动（开发默认） |
| `gorm.io/driver/mysql` | v1.5+ | MySQL 驱动 |
| `gorm.io/driver/postgres` | v1.5+ | PostgreSQL 驱动 |
| `github.com/redis/go-redis/v9` | v9.5+ | Redis 客户端 |
| `github.com/golang-jwt/jwt/v5` | v5.2+ | JWT 生成/解析 |
| `github.com/spf13/viper` | v1.19+ | 配置管理 |
| `go.uber.org/zap` | v1.27+ | 结构化日志 |
| `gopkg.in/natefinch/lumberjack.v2` | v2.2+ | 日志滚动切割 |
| `golang.org/x/crypto` | latest | bcrypt 密码加密 |
| `github.com/google/uuid` | v1.6+ | UUID（JTI/RequestID） |
| `github.com/go-playground/validator/v10` | v10.20+ | 结构体校验（Gin 内置依赖） |

### Tools — `go install`

| Tool | 用途 | 安装命令 |
|------|------|----------|
| `github.com/google/wire/cmd/wire` | 编译期 DI 代码生成 | `go install github.com/google/wire/cmd/wire@latest` |
| `github.com/swaggo/swag/cmd/swag` | Swagger 文档生成 | `go install github.com/swaggo/swag/cmd/swag@latest` |
| `github.com/golang-migrate/migrate/v4/cmd/migrate` | SQL 迁移执行 | `go install -tags 'mysql,postgres,sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| `github.com/golangci/golangci-lint/cmd/golangci-lint` | 代码规范检查 | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |

### Test-only

| Package | 用途 |
|------|------|
| `github.com/stretchr/testify` | assert/mock |
| `github.com/testcontainers/testcontainers-go` | 集成测试容器 |

### OTel & Metrics（可选，生产启用）

| Package | 用途 |
|------|------|
| `go.opentelemetry.io/otel` | 链路追踪 |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` | OTLP gRPC exporter |
| `go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin` | Gin 自动埋点 |
| `github.com/prometheus/client_golang` | Prometheus metrics |

---

## 前端 (13 packages)

### Production — `npm install`

| Package | 版本 | 用途 |
|------|------|------|
| `react` | ^18 | UI 框架 |
| `react-dom` | ^18 | React DOM |
| `react-router-dom` | ^6 | 路由 |
| `antd` | ^5 | UI 组件库 |
| `@ant-design/icons` | ^5 | AntD 图标 |
| `@ant-design/pro-table` | ^3 | 表格（搜索+分页+工具栏） |
| `axios` | ^1.7 | HTTP 客户端 |
| `zustand` | ^4 | 状态管理 |

### Dev — `npm install -D`

| Package | 版本 | 用途 |
|------|------|------|
| `typescript` | ^5.5 | 类型检查 |
| `vite` | ^5 | 构建 |
| `@vitejs/plugin-react` | ^4 | Vite React 插件 |
| `eslint` | ^9 | 代码规范 |
| `@typescript-eslint/eslint-plugin` | ^8 | TypeScript lint |
| `openapi-typescript` | ^7 | Swagger → TS 类型生成 |
| `@playwright/test` | ^1.45 | E2E 测试 |
| `eslint-plugin-react` | ^7 | React lint |
| `eslint-plugin-import` | ^2 | Import lint |

---

## 基础设施 (Docker) — M7 才需要，M1–M6 无需

开发期（M1–M6）只需要 Go 1.23+ + Node 20+。SQLite 替代 MySQL，无需 Docker。

| Image | 版本 | 用途 | 何时需要 |
|------|------|------|----------|
| `mysql` | 8.0 | 生产数据库 | M7 部署 |
| `redis` | 7 | 缓存/限流/黑名单 | **M1 就需要**（限流/缓存/JWT 黑名单依赖 Redis） |
| `jaegertracing/all-in-one` | 1.58 | trace 可视化 | M7 |
| `prom/prometheus` | v2.52 | 指标采集 | M7 |
| `grafana/grafana` | 11.0 | 仪表盘 | M7 |

### 无 Docker 开发方案

| 依赖 | 无 Docker 替代 | 安装 |
|------|---------------|------|
| MySQL | SQLite（`database.driver: sqlite`） | 零安装，自动创建 `kingfisher.db` |
| Redis | 本地 `redis-server` 或 `memurai`（Windows） | `brew install redis && redis-server` |

```bash
# 无 Docker 开发——两条命令就绪
brew install redis && redis-server &    # Redis 后台启动
make run                                 # Go 启动（SQLite 自动创建）
```

Redis 是 M1 硬依赖（限流计数器、JWT 黑名单、缓存都需要）。如果本地也无法安装 Redis，可以在 M1 暂时将限流和黑名单改为内存实现，M2 再切 Redis。

---

## 按里程碑需要的依赖

| 里程碑 | 需要安装 |
|------|----------|
| M1 Core | gin, gorm, sqlite-driver, go-redis, golang-jwt, viper, zap, lumberjack, uuid, validator, wire, swag, golangci-lint |
| M2 用户 | 同上 + `x/crypto/bcrypt` |
| M3 RBAC | 同上（无新增） |
| M4 API | 同上 + golang-migrate |
| M5 前端 | react, react-dom, react-router-dom, antd, @ant-design/icons, @ant-design/pro-table, axios, zustand, typescript, vite, @vitejs/plugin-react |
| M6 CRUD | 同上 + openapi-typescript + @playwright/test + eslint* |
| M7 生产 | 同上 + OTel packages + prometheus client + Docker images |

---

## go.mod 最小示例

```
module kingfisher

go 1.23

require (
    github.com/gin-gonic/gin v1.10.0
    github.com/gin-contrib/gzip v1.0.1
    gorm.io/gorm v1.25.12
    gorm.io/driver/sqlite v1.6.0
    gorm.io/driver/mysql v1.5.7
    gorm.io/driver/postgres v1.5.9
    github.com/redis/go-redis/v9 v9.6.1
    github.com/golang-jwt/jwt/v5 v5.2.1
    github.com/spf13/viper v1.19.0
    go.uber.org/zap v1.27.0
    gopkg.in/natefinch/lumberjack.v2 v2.2.1
    golang.org/x/crypto v0.28.0
    github.com/google/uuid v1.6.0
    github.com/go-playground/validator/v10 (indirect with gin)
    github.com/stretchr/testify v1.9.0
)
```

## package.json 最小示例

```json
{
  "name": "kingfisher-web",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint src/ --max-warnings 0",
    "test:e2e": "playwright test",
    "gen-types": "openapi-typescript http://localhost:8080/swagger/doc.json -o src/types/api.generated.ts"
  },
  "dependencies": {
    "react": "^18.3.1",
    "react-dom": "^18.3.1",
    "react-router-dom": "^6.26.0",
    "antd": "^5.21.0",
    "@ant-design/icons": "^5.5.1",
    "@ant-design/pro-table": "^3.16.0",
    "axios": "^1.7.7",
    "zustand": "^4.5.4"
  },
  "devDependencies": {
    "typescript": "^5.5.4",
    "vite": "^5.4.2",
    "@vitejs/plugin-react": "^4.3.1",
    "eslint": "^9.10.0",
    "@typescript-eslint/eslint-plugin": "^8.6.0",
    "openapi-typescript": "^7.4.0",
    "@playwright/test": "^1.47.0"
  }
}
```
