# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Kingfisher 是一个全栈后台管理系统：Go（Gin + GORM）后端 + React/TypeScript（Vite + Ant Design + Zustand）前端。单进程同时服务 API 与前端 SPA（后端通过 `static_dir` 托管 `kingfisher-web/dist`）。所有代码注释、设计文档、提交信息均为中文——新增代码请沿用中文注释风格。

## 常用命令

使用 [task](https://github.com/go-task/task)（等价于 Makefile，`brew install go-task/tap/go-task`）。开发环境：后端 8080，前端 5173，Redis 必需（localhost:6379）。

| 命令 | 作用 |
|------|------|
| `task run` | 仅启动后端（`go run ./cmd/server`） |
| `task dev` | 前后端联调。**会重置 SQLite 库 + FLUSHDB Redis + kill 8080/5173 占用进程**，非首次启动慎用 |
| `task build` | 编译到 `bin/server`（注入版本信息） |
| `task test` | 全部 Go 测试（`go test -v -race -count=1 ./...`） |
| `task test-short` | 单元测试（跳过集成/慢测试） |
| `task lint` | golangci-lint |
| `task fmt` | goimports 格式化 |
| `task swagger` | 重新生成 `docs/` Swagger 文档 |
| `task lint-fe` / `task fmt-fe` | 前端 eslint/prettier 检查/格式化（在 `kingfisher-web/` 下） |
| `task test-e2e` | Playwright E2E（需本地 Redis） |
| `task cover` | 覆盖率 |
| `task deploy` | 部署到生产服务器（git pull + systemd） |

前端单独操作：`cd kingfisher-web && pnpm dev / build / lint`（lint 用 oxlint）。单测示例：`go test -v -race -run TestAuthService ./extends/user/app/`。

## 后端架构

### Core / Extends 分层

**依赖方向：`main → extends → core`，绝对不可逆。Core 不依赖任何业务代码。**

- `cmd/server/main.go` — 组装根（composition root）：手动组装所有模块、中间件、注入依赖。**不用 Wire**（`internal/wire` 只是带 `wireinject` 的空 stub）。
- `core/` — 框架核心，零业务依赖：`config` `logger` `errcode` `response` `jwt` `database` `cache` `middleware` `router` `telemetry` `query` `taskqueue` `mailer`。
- `extends/` — 业务模块，每个模块完全独立：`user` `rbac` `menu` `config` `dict` `message` `template` `task` `audit` `email` `system`。

### 扩展模块结构（hexagonal/clean architecture）

每个 `extends/{module}/` 遵循固定分层：

```
domain/        领域实体（零依赖）
port/          接口定义（本模块需要的 Repository/Service 接口）
adapter/       GORM PO + port 接口的 MySQL 实现（实际为 SQLite 兼容驱动）
app/           用例 Service（业务逻辑，含 _test.go）
transport/     Gin Handler + 模块注册（register.go）
```

每个模块的 `transport/register.go` 实现 `core/router.Module` 接口（`Name / Init / RegisterPublic / RegisterProtected / Shutdown`），在 `main.go` 中注册。注册路由时：公开接口放 `RegisterPublic`，受保护接口放 `RegisterProtected`（自动挂 auth + RBAC 中间件）。

**模块间协作靠接口注入，不直接 import 对方**：如 user 模块通过 `InjectConfigProvider` / `InjectEmailSender` / `InjectAuditLogger` 等 setter 注入回调。新增模块时参考 `extends/template/`（完整的示例模块）。

### 鉴权与权限模型

- JWT access/refresh 双 token，`core/jwt` 管理；Redis 存撤销名单/黑名单。access 过期返回 code `10104`（前端据此用 refresh 换新），未认证 `10003`。
- RBAC：用户→角色→权限，权限码形如 `user:create`。中间件 `rbacTransport.RequirePerm("user:list")` 做细粒度控制；`AuthMiddleware` + `RBACMiddleware` 全局挂在受保护路由上。
- 会话版本号（`users.session_version`）：改密码/吊销会话后使旧 token 失效。

### 列表查询 DSL（core/query）

所有列表接口统一用 `core/query` 包解析 URL 参数，**字段必须先在 `Defs` 白名单声明**（`TypeString/TypeBool/TypeInt/TypeUint/TypeTime`，标记 `Searchable`/`Filterable`）：

- `q=关键词` — 对 `Searchable` 字符串字段做 LIKE 模糊
- `filter={"is_public":true,"status":{"in":[1,2]}}` — 支持 `eq/ne/contains/in/gt/gte/lt/lte`
- `page` / `page_size`（默认 20，最大 100）、`sort=-created_at`（`-` 前缀倒序，白名单校验，默认 `id DESC`）

### 统一响应与错误码

- 响应体 `{code, message, data}`，`code=0` 成功；分页用 `response.Page`（`items/total/page/page_size`）。
- 业务错误码集中在 `core/errcode`（如 `ErrTokenExpired`、`ErrUnauthorized`），前端 `request.ts` 依赖特定 code 值（10104 等）做 token 刷新，改码需前后端同步。

### 任务队列（core/taskqueue，asynq）

模块实现可选接口 `taskqueue.WorkerProvider` 声明自己的 worker；`main.go` 通过类型断言自动收集，**无需手动维护 handler 清单**。周期任务由 `PeriodicProviderProvider` 声明，任务管理页可动态配置（`ScheduledTaskPO` 存 cron + payload，`cron.Spec`）。

### 审计日志（extends/audit）

`auditMod.Middleware()` 挂在所有写操作前，自动记录请求（方法→业务动作映射、耗时、结果、Diff）。跨模块记录用 `AuditRecorder` 回调注入。新增写接口默认自动被审计，无需额外代码。

### 数据库

- 开发（默认）：SQLite 纯 Go 驱动（`glebarez/sqlite`，无 cgo，WAL 模式）。**GORM `AutoMigrate` + `SeedData` 自动建表播种**，seed 幂等（已有用户则跳过）。种子账号：`admin/editor/viewer/multi`，密码均为 `Abcd1234`。
- 生产：MySQL / PostgreSQL，用 `migrations/` 下的版本化 SQL（`0000NN_*.up/down.sql`）。注意 `RunMigrations` 目前是空实现，迁移逻辑未完成。
- `core/database/models.go` 集中定义全部 GORM PO（各 adapter 的 model 会引用或映射这些表结构）。新增表需同时考虑 seed 与 menu/role 关联。

## 前端架构（kingfisher-web/）

- 技术栈：React 19 + Vite 8 + Ant Design 6 + Zustand + react-router 7 + axios。
- **动态路由**：菜单树由后端菜单表驱动（`stores/menu.ts` 拉取 → `router/index.tsx` 用 `resolveComponent` 把菜单 `component` 字符串映射到懒加载页面组件）。新增页面 = 后端加菜单记录（component 填 `pages/xxx/YYY`）+ 前端建对应组件。`pages/` 目录结构与菜单 component 一一对应。
- **请求层** `api/request.ts`：统一 token 注入、响应码分发、access 过期静默刷新（刷新队列）、幂等请求指数退避重试、错误提示去重（3 秒窗口，避免并发失败弹一堆）。写操作（POST/PATCH）不重试。
- 状态：`stores/auth.ts`（用户/权限/登录）、`stores/menu.ts`（菜单树）、`stores/app.ts`。
- 列表页统一用 `components/DataTable`（与后端 `core/query` 参数契约对应：sort/filter 通过 table onChange 转为 URL 参数）。
- E2E：Playwright（`e2e/specs`），`task test-e2e`，后端用独立 `config/e2e.yaml`（端口 18080，Redis db=1，独立临时库）。

## 工程约束（CI 强制执行）

- `scripts/check-guardrails.sh` 在 CI 强制：业务代码禁止 `panic()`、`log.Fatal`、`fmt.Println/Print`；`golangci-lint`（v2.8）含 gosec/staticcheck/errcheck 等；测试排除文件放宽部分规则。
- Dockerfile 为三阶段统一镜像（前端 dist + 后端二进制 + alpine 运行），单镜像同时服务前后端。
- Swagger 注解写在 handler 上（`docs/` 由 `task swagger` 生成，需提交）。
