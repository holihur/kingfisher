# Acceptance — 验收标准

> 每个里程碑的完成定义（Definition of Done）。完成 = 所有检查项通过，不通过 = 不能进入下一个里程碑。

## 自动化策略

**100% 自动化。不接受人工验收。**

所有验收项由以下层覆盖，CI 中强制执行。任何一层不通过，整个 pipeline 阻断。

自动化覆盖情况：

| 测试层 | 覆盖内容 | 对应验收章节 |
|--------|----------|-------------|
| `go test -short` | Config 校验、中间件、JWT 生成/解析、Service 逻辑、校验器 | 各 M 的"工程验收""测试验收" |
| `go test` + testcontainers | Repository CRUD、事务、缓存、迁移 | 各 M 的"数据验收""迁移验收" |
| `httptest` / curl 脚本 | 全部 API Happy Path + Sad Path | 各 M 的"功能验收""容错验收""鉴权验收" |
| Playwright | 登录流程、CRUD 页面交互、权限按钮显隐、表单验证 | M5/M6 的"功能验收""UI 验收""表单验收""权限按钮验收" |
| `go test -bench` + Vegeta | 性能指标 | M7 的"性能验收" |

## 覆盖率

| 里程碑 | Happy Path | Sad Path | Edge Case | Chaos | 合计 |
|--------|-----------|----------|-----------|-------|------|
| M1 Core | 25 | 9 | 7 | 0 | 41 |
| M2 用户 | 10 | 21 | 4 | 5 | 40 |
| M3 RBAC | 15 | 13 | 0 | 5 | 33 |
| M4 API | 17 | 16 | 5 | 0 | 38 |
| M5 登录 | 43 | 14 | 3 | 0 | 60 |
| M6 CRUD | 59 | 16 | 22 | 0 | 97 |
| M7 生产 | 58 | 7 | 0 | 10 | 75 |
| 跨里程碑 | 31 | — | — | — | 31 |

| 维度 | 条目数 | 占比 |
|------|--------|------|
| Happy Path（预期行为） | ~150 | 48% |
| Sad Path（输入异常/权限不足/资源不存在） | ~60 | 19% |
| Edge Case（边界值/并发/特殊字符/浏览器行为） | ~41 | 13% |
| Chaos（服务中断/恢复） | ~20 | 7% |
| 工程质量（编译/lint/测试/覆盖率/依赖安全） | ~40 | 13% |

> 无未决项。每条验收均可直接写成测试用例——执行者零决策。

---

## M1 — Core 框架启动

### 🤖API 功能验收

- [ ] `curl localhost:8080/health` 返回 `{"status":"ok"}`，HTTP 200
- [ ] `curl localhost:8080/ready` 返回 `{"status":"ready","mysql":"ok","redis":"ok"}`，HTTP 200
- [ ] `curl localhost:8080/metrics` 返回 Prometheus 格式指标，至少含 `http_requests_total`

- [ ] MySQL 不可用时启动：重试 3 次后 Fatal 退出，退出码非零
- [ ] Redis 不可用时启动：Fatal 退出，日志含 `redis init failed`
- [ ] `JWT_SECRET` 为空时启动：Fatal 退出，日志含 `JWT secret must be set`
- [ ] `curl localhost:8080/api/v1/notexist` 返回 404，非 panic 500

### Sad Path & Edge Case

- [ ] 配置文件路径不存在 → `Load()` 返回明确错误，非 panic
- [ ] `config.yaml` 格式错误（缺少冒号、缩进错误） → 启动失败，日志指出具体行号
- [ ] MySQL DSN 中密码含特殊字符（`@` `:` `%` `#`） → 连接正常（DSN 已 URL encode）
- [ ] MySQL 连接超时（firewall 阻断） → 3 次重试后 Fatal，日志含每次重试的间隔
- [ ] Redis 连接超时 → Fatal，不跳过（Redis 是核心依赖）
- [ ] `JWT_SECRET` 值 = `change-me-in-production`（默认占位符） → 启动拒绝，日志提示必须修改
- [ ] 端口被占用 → Fatal，日志含 `address already in use`
- [ ] 配置中 `server.port: 0` → Validate 拒绝
- [ ] 配置中 `server.port: 99999` → Validate 拒绝
- [ ] `server.mode` 值非 debug/release/test → Validate 拒绝
- [ ] 同时发 100 个请求到 `/health` → 全部 200，无连接拒绝
- [ ] SIGTERM 时正在处理一个需要 10s 的慢请求 → Shutdown 等够 30s，慢请求完成，正常 200；Shutdown 超时后才强制断开
- [ ] 启动后立即 SIGTERM（< 1s） → 不 panic，正常退出
- [ ] 连续两次 SIGTERM → 第二次强制退出（不等 Shutdown）
- [ ] 日志文件所在目录不存在 → Zap 创建目录或使用 lumberjack 自动创建
- [ ] 日志文件权限不足（chmod 000） → 启动失败，日志输出到 stderr 而非静默

### 中间件验收

- [ ] 任意请求返回头含 `X-Request-ID`
- [ ] 请求头传入 `X-Request-ID: my-id` 时，响应头回传 `X-Request-ID: my-id`
- [ ] CORS：OPTIONS 预检请求返回 204，含 `Access-Control-Allow-Origin`
- [ ] 请求已注册但无 handler 的路由：返回 405 Method Not Allowed

### 安全验收

- [ ] 请求体 > 10MB 返回 413（非 OOM）
- [ ] 配置文件中无明文密码（password/secrets 均为空或占位符）

### 工程验收

- [ ] `go build ./cmd/server` 编译无错误
- [ ] `golangci-lint run` 无 error
- [ ] `go vet ./...` 无输出
- [ ] 目录结构与设计文档一致（`core/` 下 7 个子包，不 import 任何 `extends/` 包）
- [ ] `core/router/engine.go` 中 `Module` 接口被 `extends/` 实现，不在 `core/` 实现
- [ ] `go.mod` 无 `replace` 指令指向本地路径（可独立构建）
- [ ] 启动日志中无 `[WARNING]` 或 `[ERROR]`（INFO 级别以上）

### 测试验收

- [ ] `go test -short ./core/...` 全部通过，耗时 < 5s
- [ ] `core/middleware` 每个中间件有单测（覆盖正常 + 异常路径）
- [ ] `core/jwt` 有单测：生成/解析/过期/黑名单

---


### SQLite 开发模式验收

- [ ] `config.yaml` 中 `database.driver: sqlite` → `make run` 零依赖启动（无需 Docker、MySQL）
- [ ] 启动后自动创建 `kingfisher.db`（项目根目录）
- [ ] GORM AutoMigrate 自动建表（8 张表）
- [ ] 种子数据自动写入（admin 用户 / 3 角色 / 15 权限 / 15 菜单 / 5 配置）
- [ ] `sqlite3 kingfisher.db "SELECT * FROM users;"` 可读到 admin 记录
- [ ] 切换 `database.driver: mysql` → 连接 MySQL → golang-migrate 执行 migrations/*.sql
- [ ] 切换 `database.driver: postgres` → 连接 PG → golang-migrate 执行 migrations/*.sql
- [ ] 三驱动切换时 extends 模块代码零改动
- [ ] `kingfisher.db` 在 `.gitignore` 中，不会提交

## M2 — 用户注册登录

### 功能验收

- [ ] `POST /api/v1/auth/register` 创建用户成功，HTTP 200，返回用户信息（不含 password）
- [ ] 重复注册相同 username：返回 `code:10101`，HTTP 400
- [ ] `POST /api/v1/auth/login` 传入正确密码：返回 `access_token` + `refresh_token` + `user`，HTTP 200
- [ ] `POST /api/v1/auth/login` 传入错误密码：返回 `code:10103`，HTTP 400
- [ ] `POST /api/v1/auth/login` 不存在的用户名：统一返回 `code:10103`（密码错误），HTTP 400。不区分"不存在"和"密码错"防用户枚举
- [ ] `POST /api/v1/auth/refresh` 传入有效 refresh_token：返回新 `access_token`
- [ ] `POST /api/v1/auth/refresh` 传入 access_token：拒绝，返回 `code:10105`
- [ ] `POST /api/v1/auth/refresh` 传入过期 refresh_token：拒绝，返回 `code:10104`
- [ ] `GET /api/v1/users/:id` 不带 token：返回 `code:10003`，HTTP 401
- [ ] `GET /api/v1/users/:id` 带有效 token：返回用户信息，HTTP 200

### 容错验收

- [ ] 密码 < 8 位注册：返回参数错误
- [ ] 密码纯小写注册：返回密码强度不足
- [ ] username 含 `<script>` 注册：返回参数错误（XSS 防护）
- [ ] 登录失败第 6 次（1 分钟内）：返回 `code:10107`，HTTP 429

### Token 验收

- [ ] access_token 在 2h 后过期，过期后访问受保护接口返回 `code:10104`
- [ ] refresh_token 在 7d 后过期，过期后刷新返回 `code:10104`
- [ ] token payload 含 `user_id`、`role`、`jti`、`exp`、`type`
- [ ] 注销后的 token 再次使用返回 `code:10105`（token 已失效）

### 数据验收

- [ ] 用户密码在数据库中存储为 bcrypt hash（以 `$2a$` 开头）
- [ ] API 响应中不出现 password 字段
- [ ] 软删除用户后，`GET /api/v1/users/:id` 返回 404

### Sad Path & Edge Case

- [ ] 注册时 username 为空字符串 → 返回参数错误
- [ ] 注册时 username 长度 = 33（max 32） → 返回参数错误
- [ ] 注册时 username 纯空格 → 返回参数错误（trim 后为空）
- [ ] 注册时 username 含 emoji `🤡admin` → 返回参数错误（username 只允许字母数字下划线连字符）
- [ ] 注册时 username 含 Unicode 零宽字符 → 返回参数错误或正常化处理
- [ ] 注册时 password 长度 = 7（min 8） → 返回参数错误
- [ ] 注册时 password 长度 = 65（max 64） → 返回参数错误
- [ ] 注册时 password 全大写无数字 → 返回密码强度不足
- [ ] 注册时 password 全数字无字母 → 返回密码强度不足
- [ ] 注册时 email 格式 `notanemail` → 返回邮箱格式错误
- [ ] 注册时 email 超长（> 128） → 返回参数错误
- [ ] 登录时 username 大小写不敏感（MySQL utf8mb4_unicode_ci 默认行为，`admin` 和 `Admin` 都能登录）
- [ ] 登录时 `Content-Type: text/plain` 而非 `application/json` → Gin binding 失败，返回 400
- [ ] 登录时 body 为空 → 返回参数错误
- [ ] 登录时 body 为非 JSON（`this is not json`） → 返回 400，非 panic
- [ ] 登录时 body JSON 有多余未知字段 → 忽略，正常处理（非严格模式）。安全审计时再决定是否改为拒绝
- [ ] 登录时 username 含 SQL 注入 `admin' OR '1'='1` → GORM 参数化查询防注入，返回用户不存在
- [ ] Refresh 时 token 被篡改（改中间一个字符） → 签名验证失败，返回 `code:10105`
- [ ] Refresh 时 token 算法改为 `none`（alg=none 攻击） → golang-jwt 默认拒绝
- [ ] Refresh 时传入空字符串 → 返回 `code:10001`
- [ ] Refresh 时传入的不是 JWT（随便一个字符串） → 返回 `code:10105`
- [ ] 并发注册两个相同 username → 一个成功一个返回 `code:10101`（unique index 保证）
- [ ] 用户被软删除后用相同 username 注册 → 拒绝，返回 `code:10101`。从 deleted_at 不为空的记录中恢复 username 唯一索引失效，需手动清理旧记录或清空 deleted_at 才能重用
- [ ] 登录时 bcrypt.CompareHashAndPassword 耗时 ~250ms（cost=12） → P99 在预算内
- [ ] 并发登录同一用户 100 次 → 登录限流 5次/min 生效，返回 429

### Chaos

- [ ] 登录时 Redis 不可用 → 限流中间件获取计数失败 → 不阻断请求（限流失效但不阻止登录），日志 `warn: ratelimit redis error`
- [ ] 登录时 Redis 不可用 → JWT 生成不受影响（JWT 签名不依赖 Redis）→ 登录仍然成功
- [ ] Refresh 时 Redis 不可用 → JWT 黑名单 `IsRevoked` 返回 error → 拒绝刷新，返回 503
- [ ] 注册时 MySQL 连接超时 → repository.Create 超时返回 error → HTTP 500，不 panic
- [ ] Token 黑名单 Redis key 被手动 `FLUSHALL` 清掉 → 已注销 token 变为可用 → 接受此风险（token 本身有时效性，且 `FLUSHALL` 是运维事件非正常操作）

### 工程验收

- [ ] `extends/user/` 下 5 个子目录结构完整（domain/port/adapter/app/transport）
- [ ] `extends/user/app/service.go` 不 import `gin` — Service 层与 HTTP 框架解耦
- [ ] `extends/user/adapter/mysql/repo.go` 实现 `port.UserRepository` 接口 — 编译期检查
- [ ] `extends/user/transport/register.go` 实现 `core.Module` 接口 — 编译期检查
- [ ] 无循环依赖：`extends/user` 不 import `extends/rbac` 或 `extends/menu`

### 测试验收

- [ ] `extends/user/app/service_test.go`：Register/Login/RefreshToken 各 ≥ 2 个用例
- [ ] `extends/user/adapter/mysql/repo_test.go`：CRUD + FindByUsername + 软删除
- [ ] Login 单测覆盖：正确密码 / 错误密码 / 用户不存在 / 用户已禁用
- [ ] Service 单测用 mock Repository，不连真实 DB

---

## M3 — RBAC 鉴权

### 功能验收

- [ ] 预设 3 个角色：admin（全部权限）、editor（查看编辑）、viewer（只读）
- [ ] 预设 15 个权限（14 业务 + 1 审计 audit:list），按 resource 分 5 组
- [ ] `GET /api/v1/roles` 返回 3 个角色
- [ ] `GET /api/v1/roles/:id` 返回角色 + 关联权限 + 关联菜单
- [ ] `POST /api/v1/roles` 创建角色成功
- [ ] `PUT /api/v1/roles/:id/permissions` 分配权限成功
- [ ] `PUT /api/v1/roles/:id/menus` 分配菜单成功
- [ ] `GET /api/v1/permissions` 返回全部 15 个权限

### 鉴权验收（关键）

- [ ] admin 登录 → `GET /api/v1/users` → 200，可以看到所有用户
- [ ] viewer 登录 → `GET /api/v1/users` → 403
- [ ] viewer 登录 → `GET /api/v1/configs` → 200（只有查看权限）
- [ ] viewer 登录 → `PUT /api/v1/configs/site_name` → 403
- [ ] 无角色用户登录 → 只能访问公开接口
- [ ] admin 登录 → `GET /api/v1/menus/tree` → 返回全部菜单
- [ ] editor 登录 → `GET /api/v1/menus/tree` → 只返回有权限的菜单

### 中间件验证

- [ ] Auth 中间件：无 token → 401；过期 token → 401；有效 token → ctx 注入 user_id/role
- [ ] RBAC 中间件：admin → ctx 注入 15 个 perm（含 audit:list）；editor → ctx 注入 8 个 perm；viewer → ctx 注入 4 个 perm
- [ ] RequirePerm 中间件：有权限 → c.Next()；无权限 → 403

### 性能验收

- [ ] 权限列表被缓存，首次查 DB 后后续从 Redis 读取
- [ ] 用户变更角色后，权限缓存在 30min 内失效或手动清除
- [ ] RBAC 中间件单次耗时 < 5ms（含 Redis 读取）

### Sad Path & Edge Case

- [ ] Token payload 中 role 字段被篡改为 `admin`（签名为原始 viewer） → ParseToken 签名校验可发现（HMAC）或 payload 不被信任，重新查 DB/Redis 获取角色
- [ ] 角色被删除后，已登录的该角色用户 → RBAC 中间件返回空权限列表（不 crash、不 500）
- [ ] 用户属于一个已删除的角色 → 权限列表为空，等同于无角色
- [ ] 给角色分配一个不存在的 permission ID → 返回参数错误
- [ ] 给角色分配空数组 `[]` → 清空该角色所有权限（confirmed：确认不是 bug 是有意行为）
- [ ] 给角色分配菜单时 menu_id 对应一个 type=3（按钮）→ 允许（菜单分配不区分 type）
- [ ] 创建角色的 code 重复 → 返回 `code:10301`
- [ ] 删除角色时该角色仍有用户 → 返回 `code:10303`（或允许但用户失去角色）
- [ ] RBAC 中间件中 Redis 不可用 → 拒绝所有需要权限检查的请求，返回 503 `{"code":10009,"message":"service unavailable"}`（安全优先于可用性）。缓存 miss（Redis GET 成功但 key 不存在）→ 正常回源 DB，不是降级。Redis 连接失败（不可达/超时）→ 拒绝。登录接口和公开接口不受影响。
- [ ] RBAC 中间件中 Redis 超时（>3s） → 有 timeout 保护，不阻塞请求
- [ ] 权限缓存键被意外删除 → miss 回源 DB，透明恢复
- [ ] 用户 ID 不存在但 token 有效 → RBAC 查不到权限，返回空列表
- [ ] 并发 Admin 修改用户角色 + 用户正在请求 → 最终一致性，最迟 30min 内生效

### Chaos

- [ ] Redis 不可用时 Admin 调用 `PUT /roles/:id/permissions` → 分配成功（DB 写成功），缓存失效 `cache.Delete` 失败 → 日志 warn → 权限缓存过期（30min）后自动刷新，此窗口内旧权限仍生效
- [ ] Redis 不可用时用户请求受保护接口 → RBAC 中间件 `GetUserPermissions` 的缓存连接失败 → 拒绝请求返回 503。缓存 key 不存在（miss）→ 正常回源 DB 查询权限（非降级，是 Cache-Aside 标准行为）
- [ ] Redis 和 DB 同时不可用 → RBAC 中间件返回 503 → 登录接口仍可用（登录只依赖 DB + JWT）
- [ ] 权限缓存 key 被手动删除后用户请求 → cache miss → 回源 DB 重建缓存 → 延迟 +5ms 但功能正常
- [ ] 分配权限时传入空数组 `[]` → 清空该角色所有权限（有意行为，非 bug）→ 该角色用户下次请求返回空权限列表

### 工程验收

- [ ] `extends/rbac/` 中间件已拆分为 3 个独立文件（auth/rbac/require_perm）— 单一职责
- [ ] `extends/rbac/transport/auth_middleware.go` 不依赖 `port.RoleService` — Auth 只管认证
- [ ] `extends/rbac/app/role_service.go` 调用 `GetUserPermissions` 时有缓存逻辑
- [ ] `RequirePerm` 不依赖 DB/Redis — 只读 ctx 中已注入的 permissions
- [ ] `extends/rbac/` 不 import `extends/user` — 无循环依赖

### 测试验收

- [ ] `AuthMiddleware`：无 token / 过期 token / 有效 token / 已注销 token
- [ ] `RBACMiddleware`：有角色 / 无角色 / 权限列表为空
- [ ] `RequirePerm`：有权限 / 无权限
- [ ] `RoleService.AssignPermissions` 分配后缓存即时失效

---

## M4 — 后端 API 全量

### 功能验收

- [ ] `GET /api/v1/menus/tree` 返回树形结构，层级 ≥ 2
- [ ] `POST /api/v1/menus` 创建菜单成功
- [ ] `DELETE /api/v1/menus/:id` 有子节点时返回 `code:10203`
- [ ] `GET /api/v1/configs` 返回 4 个预设配置
- [ ] `PUT /api/v1/configs/:key` 更新配置后，缓存即时失效
- [ ] 列表接口支持分页：`?page=1&page_size=20`
- [ ] 列表接口支持搜索：`?keyword=admin`

### Sad Path & Edge Case

- [ ] 创建菜单 parent_id 指向不存在的菜单 → 返回参数错误
- [ ] 创建菜单 path 重复（如两条 `/system/users`） → 拒绝，返回 `code:10201`，path 在 type=2（菜单）时必须唯一
- [ ] 删除菜单时 Redis 缓存不可用 → 删除成功（DB 已删），缓存失效静默失败，下次读取时 miss 回源
- [ ] 菜单树节点 > 1000 时递归构建 → 不栈溢出，时间 < 100ms
- [ ] 菜单树缓存 Redis 中数据损坏（非 JSON）→ 回源 DB 重建缓存
- [ ] 分页 page=0 → 自动修正为 page=1（静默修正，不返回错误）
- [ ] 分页 page=-1 → 返回参数错误 `code:10001`
- [ ] 分页 page_size=0 → 自动修正为默认值 20
- [ ] 分页 page_size=1000（超 max 100） → 截断为 max 100（静默修正）
- [ ] 排序字段 sort=injection;DROP TABLE → 校验失败，返回参数错误
- [ ] `GET /api/v1/users/:id` id 为非数字 `abc` → Gin 路由不匹配，返回 404
- [ ] `GET /api/v1/users/:id` id=0 → FindByID 返回 not found
- [ ] `GET /api/v1/users/:id` id=2^32-1（uint 最大值） → not found，不 overflow
- [ ] 事务中第一个操作成功第二个失败 → 两个都回滚
- [ ] 事务中 DB 连接断开 → 事务回滚，返回 500
- [ ] 并发更新同一用户 → 后提交的覆盖前提交的（last write wins），无乐观锁冲突检测（如需要则后续版本加 version 字段）
- [ ] 迁移 `make migrate-up` 时数据库已有同名表 → 跳过或报错？golang-migrate 默认报错（dirty state）
- [ ] 迁移 SQL 文件 checksum 被修改 → golang-migrate 拒绝执行，报 dirty
- [ ] Wire `make wire` 时有循环依赖 → wire 编译报错，输出清晰错误信息
- [ ] Wire 生成的代码中缺少某个 Provider → 编译错误，非运行时 panic

### Swagger 验收

- [ ] `make swagger` 执行无报错
- [ ] 浏览器 `/swagger/index.html` 可见全部 4 组接口
- [ ] 每个接口点开展示 Request/Response Schema
- [ ] Swagger UI 上 Try it out → Execute 返回正确结果
- [ ] 所有 `@Success` 注解都指定了 `{object}` 类型

### 迁移验收

- [ ] `migrations/` 目录含 10 个 SQL 文件（8 建表 + 1 种子 + 1 session_version ALTER）
- [ ] 每个 `.up.sql` 有对应的 `.down.sql`
- [ ] `make migrate-up` 执行成功，8 张表创建（users/roles/permissions/role_permissions/menus/role_menus/system_configs/audit_logs）
- [ ] `make migrate-down` 回滚成功，表删除
- [ ] 种子数据：admin 用户可登录、3 角色存在、14 权限存在、15 菜单存在、4 配置存在

### 事务验收

- [ ] 创建用户同时分配角色：其中一个失败，另一个回滚
- [ ] 事务中 panic：回滚
- [ ] 事务外独立操作：互不影响

### Wire 验收

- [ ] `make wire` 执行无报错
- [ ] `wire_gen.go` 中所有依赖编译通过
- [ ] 新增 extends 模块后 `make wire` 自动生成新 Provider 代码

### 工程验收

- [ ] `internal/wire/wire_gen.go` 由 `make wire` 生成，禁止手改
- [ ] 4 个 extends 模块全部注册到 `main.go` 且编译通过
- [ ] `make swagger` 生成 `docs/` 后，`api.generated.ts` 可刷新（CI 不报 diff）
- [ ] `migrations/` 目录中 SQL 文件名序号连续、无跳号
- [ ] `make test` 全部通过（含集成测试），耗时 < 2min

---

## M5 — 前端登录闭环

### 功能验收

- [ ] 浏览器 `localhost:5173` → 自动跳转 `/login`
- [ ] 不输入任何内容点登录 → 表单校验提示"请输入用户名"
- [ ] 输入错误密码 → 提示"密码错误"
- [ ] 输入 admin/Abcd1234 → 登录成功，跳转首页 `/dashboard`
- [ ] 登录后左侧出现侧边栏菜单树（系统管理 > 用户/菜单/角色/配置 共 4 个子菜单）
- [ ] 顶栏显示当前用户名 "admin"，右侧有退出按钮
- [ ] 点击退出 → 返回登录页
- [ ] 退出后直接访问 `/dashboard` → 跳转登录页
- [ ] 登录后刷新页面 → 保持登录状态，不跳回登录页

### UI 验收

- [ ] 登录页居中显示，Logo + 系统名称
- [ ] 侧边栏可折叠/展开，折叠时只显示图标
- [ ] 菜单根据后端返回的 `icon` 渲染对应 AntD 图标
- [ ] 面包屑与当前页面路径一致，点击可跳转
- [ ] 页面切换时，侧边栏高亮当前菜单项

### Token 验收

- [ ] 登录后 access_token 存储在 localStorage
- [ ] 请求自动带 `Authorization: Bearer <token>`
- [ ] 接口返回 401 时自动触发 refresh → 重放请求
- [ ] 并发 3 个请求同时遇 401：只发起 1 次 refresh，其余排队
- [ ] refresh 也失败 → 清除 token → 跳转登录页

### 状态覆盖验收

- [ ] 首次进入页面：显示骨架屏（非白屏）
- [ ] 网络断开时操作：提示"网络异常，请检查网络连接"
- [ ] 后端返回 500：提示"服务器内部错误"
- [ ] 后端返回 429：提示"请求过于频繁，请稍后再试"

### Sad Path & Edge Case

- [ ] 登录时后端完全不可达（关掉 Go） → 不白屏，提示"网络异常，请检查网络连接"
- [ ] 登录时后端返回非 JSON → Axios 不 crash，提示解析错误
- [ ] 登录时后端返回 HTTP 502（网关错误） → 提示"服务器错误"
- [ ] 登录成功后 access_token 在 localStorage 被手动删除 → 下次请求 401 → 自动跳转登录页
- [ ] 登录成功后 refresh_token 过期 → 刷新失败 → 清除 token → 跳转登录页
- [ ] 登录成功后浏览器关闭再打开 → token 从 localStorage 恢复，自动登录
- [ ] 登录页面在已登录状态直接访问 → 自动跳转首页，不闪登录页
- [ ] 登录页面 URL 带 `?redirect=/system/users` → 登录后跳转到 `/system/users` 而非首页
- [ ] 用户连续输入错误密码 5 次 → 第 6 次返回 429 → 前端提示"请求过于频繁"
- [ ] 菜单树 API 返回空数组 `[]` → 侧边栏不渲染菜单项（不白屏），内容区正常
- [ ] 菜单树 API 返回 500 → 侧边栏显示"菜单加载失败"，重试按钮可用
- [ ] 用户信息 API 返回 500 → 顶栏显示"--"，不崩溃
- [ ] Token 即将过期（exp 时间距现在 < 30s）时发起请求 → 前端主动 refresh 后再发请求，不等 401。实现：Axios 请求拦截器中检查 `jwtDecode(token).exp`
- [ ] 3 个 Tab 页同时打开，Tab A 退出登录 → Tab B 和 C 下次操作时检测到 token 无效 → 跳转登录
- [ ] localStorage 满了（5MB） → 写入 token 失败，给用户提示
- [ ] 浏览器禁用 localStorage（隐私模式某些设置） → 登录时 `setToken` 抛出 `QuotaExceededError` → 捕获后提示"您的浏览器禁用了本地存储，请开启后重试" → 不跳转，留在登录页
- [ ] 用户复制 URL 在新标签页打开 → Auth Guard 检查 token → 有则放行无则跳转登录
- [ ] HMR 热更新时修改 Auth Store → 不触发全量刷新，状态不丢失
- [ ] `npm run build` 时发现 TS 类型不匹配（如 API 返回字段改名了） → 编译报错（如果用了 shared types）

### 工程验收

- [ ] `npm run build` 无 TS 错误，无 warning
- [ ] `npm run lint` 无 error（ESLint）
- [ ] `npm run preview` 构建产物可正常访问
- [ ] 构建产物 `dist/` 大小 < 2MB（gzip 后 < 500KB）
- [ ] 路由懒加载：各页面独立 chunk，首屏只加载当前页面 JS
- [ ] HMR 生效：修改 `AdminLayout.tsx` 中的文字，浏览器即时刷新不丢失状态

### 组件库验收

- [ ] `<PermissionBtn code="xxx">` 存在，当用户无该权限时不渲染子元素
- [ ] `<Skeleton type="table|card|detail">` 三种变体可用，API 一致
- [ ] `<EmptyState>` 三种场景：default（暂无数据）/ search（搜索无结果）/ nodata（从未有过数据）
- [ ] `<ErrorResult>` 四种状态：error（通用）/ 404 / 403 / 500
- [ ] 以上组件在 `components/` 目录，被至少一个页面引用且通过 lint

### 路由验收

- [ ] 访问未定义路由 `/abc` → 显示 404 页面（非白屏）
- [ ] 无权限用户直接访问 `/system/users` URL → 跳转 403 页面或首页（非白屏）
- [ ] 路由切换时浏览器标题随之变化（`document.title`）
- [ ] 登录页 `/login` 在已登录状态访问 → 自动跳转首页

### 样式验收

- [ ] 侧边栏：深色背景 `#001529`，选中项高亮 `#1677ff`
- [ ] 顶栏：白色背景，底部阴影 `box-shadow: 0 1px 4px rgba(0,0,0,0.08)`
- [ ] 内容区：`margin: 24px`，白色背景，`border-radius: 8px`
- [ ] 侧边栏折叠时：菜单文字隐藏，仅展示图标，宽度 ~80px
- [ ] 浏览器窗口缩窄至 < 768px：侧边栏自动隐藏或转为抽屉式
- [ ] 页面滚动时顶栏 fixed，内容区正常滚动

### Store 验收

- [ ] 退出登录后：`localStorage` 中 token 清空，Zustand 中 `userInfo`/`permissions`/`menuTree` 重置为初始值
- [ ] 刷新页面后：token 从 localStorage 恢复，自动调用 `fetchUserInfo` 和 `fetchMenus`
- [ ] 两个浏览器 Tab 同时登录不同用户：各自 Store 独立，互不污染

### 可访问性验收

- [ ] 登录表单 Tab 键可在输入框间切换
- [ ] 登录按钮 Enter 键可提交
- [ ] 侧边栏菜单键盘方向键可导航
- [ ] Modal 打开时焦点自动进入，关闭后焦点回到触发按钮
- [ ] `<img>` 标签有 `alt` 属性

---

- [ ] `GET /api/v1/audit-logs?page=1&page_size=20` → 返回审计日志分页列表
- [ ] `GET /api/v1/audit-logs?resource=user&action=DELETE` → 筛选生效
- [ ] 审计日志不可修改、不可删除（合规）：PUT/DELETE /audit-logs/:id → 405

## M6 — 全栈 CRUD 闭环

### 用户管理验收

- [ ] 用户列表页显示分页表格，列：ID/用户名/邮箱/角色/状态/创建时间/操作
- [ ] 搜索框输入关键词 → 过滤列表
- [ ] 点击"新增用户" → 弹窗 → 填写 → 提交 → 列表刷新
- [ ] 编辑用户：修改邮箱 → 提交 → 数据更新，用户名 disable 不可改
- [ ] 删除用户：点击 → Popconfirm 确认 → 删除 → 列表刷新
- [ ] 无 `user:create` 权限时，"新增用户"按钮不渲染
- [ ] 无 `user:delete` 权限时，"删除"按钮不渲染

### 菜单管理验收

- [ ] 展开树形表格，层级缩进正确
- [ ] 类型标签颜色：目录=蓝/菜单=绿/按钮=橙
- [ ] 点击"新增根菜单" → 弹窗 → TreeSelect 可选父级
- [ ] 创建子菜单：选择父级 → 填入 → 提交 → 树刷新
- [ ] 有子节点的菜单删除按钮禁用或后端拒绝

### 角色管理验收

- [ ] 角色列表显示 3 个角色
- [ ] 点击"权限" → 弹窗 Tab 页 4 组 → 勾选 → 保存
- [ ] 点击"菜单" → Tree 勾选 → 保存
- [ ] 修改角色权限后，该角色用户重新登录 → 菜单变化

### 系统配置验收

- [ ] 配置列表显示 4 个配置项
- [ ] 编辑 site_name → 保存 → 登录页/页头标题变化
- [ ] 编辑 max_login_attempts → 保存 → 登录失败上限变化

### URL 状态同步验收

- [ ] 用户列表搜索"admin"→ 分页第 2 页 → URL 出现 `?keyword=admin&page=2`
- [ ] 刷新浏览器 → 列表保持搜索词"admin"、第 2 页（URL 参数复原）
- [ ] 浏览器后退 → 回到上一页——不是回到上一组搜索参数（`replaceState` 而非 `pushState`）
- [ ] 复制 URL 粘贴到新 Tab → 打开同一页面、同一搜索词、同一页码
- [ ] 清除搜索词 → URL 中 `keyword` 参数消失
- [ ] 切换 page_size 20→50 → URL `page_size=50`，且回到 `page=1`
- [ ] 菜单管理展开/折叠节点 → URL `expanded` 参数同步
- [ ] 弹窗打开/关闭 → URL 不变（弹窗状态不进 URL）
- [ ] 表单输入未保存 → URL 不变（表单草稿不进 URL）

### 全链路验收

- [ ] 浏览器创建用户 → `mysql -e "SELECT * FROM users"` 看到新记录
- [ ] 浏览器修改角色权限 → API 返回新权限列表
- [ ] 浏览器删除菜单 → 侧边栏菜单项消失

### Sad Path & Edge Case

- [ ] 用户列表页 page=99999（远超实际页数） → 返回空数组 `[]`，total 正确，非 500
- [ ] 用户列表搜索关键词含 `%` 或 `_`（SQL LIKE 通配符） → 正确转义，返回匹配结果
- [ ]  用户列表页连续快速点击搜索（防抖） → 只发最后一个请求，不堆积
- [ ] 编辑用户弹窗打开中，网络断开 → 保存按钮点击后提示网络错误，弹窗不关数据不丢
- [ ] 编辑用户时，另一个管理员已删除该用户 → 保存时后端返回 404，前端提示"用户不存在或已被删除"
- [ ] 新增用户时填写超长 username（> 32） → 前端 maxLength 限制 + 后端校验双层防护
- [ ] 新增用户时粘贴一段富文本到 email 字段 → 后端校验格式 + 防 XSS（nohtml）
- [ ] 删除用户时该用户正在被另一个管理员编辑 → Popconfirm 确认后删除成功
- [ ] 删除自己（删除当前登录的用户） → 拒绝，返回 `code:10001` `"message":"不能删除当前登录用户"`。前端在确认删除时如果目标 ID === 当前 user_id，按钮 disabled 且有 Tooltip 提示
- [ ] 菜单管理页面菜单树 > 200 条 → 渲染不卡顿（ProTable 虚拟滚动或分页）
- [ ] 菜单管理新建子菜单后父级 Type=菜单 → 应拒绝？或前端提示改为目录
- [ ] 角色管理"分配权限"弹窗中取消勾选全部权限 → 保存后该角色权限列表为空
- [ ] 角色管理"分配菜单"取消勾选顶级菜单 → 其子菜单也取消勾选（AntD Tree 默认行为）
- [ ] 系统配置编辑一个不存在的 key（URL 手动输入） → 后端返回 `code:10401`
- [ ] 系统配置的 value 为空字符串 → 允许（部分配置可置空，如 `site_logo=""` 表示不使用 Logo）
- [ ] 浏览器点 X 关掉弹窗（ESC 键） → Form 数据重置，不触发保存
- [ ] 弹窗打开时按 Tab 键多次 → 焦点在弹窗内循环，不跳到背景页面
- [ ] 页面滚动一定深度后打开弹窗 → 弹窗居中（不滚到页面顶部）
- [ ] 分页器显示"共 0 条"时 → 不显示分页控件
- [ ] 切换每页条数 20→50 后 → 回到第 1 页，URL query 参数同步
- [ ] 连续切换页面时快速点击 → 请求按发出顺序返回，不存在后发先到覆盖

### 浏览器 Edge Case

- [ ] 浏览器后退按钮从编辑弹窗返回到列表页 → 弹窗关闭，列表页状态保持（搜索条件、页码不变）
- [ ] 浏览器前进按钮 → 回到弹窗打开状态（history.pushState 在弹窗打开时记录了 URL hash 或 query）
- [ ] 浏览器后退到登录页 → 如已登录则自动跳转首页
- [ ] 深链接访问 `/system/users` 未登录 → 跳转 `/login?redirect=/system/users` → 登录后回到 `/system/users`
- [ ] 浏览器页面被系统回收内存（mobile Safari 或 Chromium tab discarding）→ 重新打开 Tab 时 React 重新渲染，Auth Guard 从 localStorage 恢复 token → 自动登录
- [ ] `window.history.replaceState` 和 `pushState` 混用 → URL 始终与页面状态一致，不会出现 `/system/users` 的 URL 显示首页内容
- [ ] 用户手动修改 URL hash → 不触发页面跳转（React Router 控制的 hash 路由）
- [ ] 多 Tab 协同：Tab A 新增用户，Tab B 刷新列表 → Tab B 能看到新用户
- [ ] 多 Tab 冲突：Tab A 和 Tab B 同时编辑同一用户 → 先提交成功，后提交覆盖（last-write-wins，无乐观锁）
- [ ] `window.onbeforeunload`：表单有未保存内容时关闭 Tab → 弹窗提示"你有未保存的内容，确定离开吗？"（非 Modal 内，是浏览器原生弹窗）
- [ ] 浏览器 `Ctrl+Shift+N` 隐私模式 → 登录功能正常，但 localStorage 在窗口关闭后被清除 → 下次打开需重新登录（符合预期）
- [ ] `Ctrl+F5` 强制刷新 → 绕过缓存，重新加载 JS/CSS → 页面正常渲染（路由与 URL 匹配）
- [ ] SVG icon 缺失时（如后端返回不存在的 AntD 图标名） → 侧边栏显示默认 icon（`QuestionOutlined`），不报错不白屏
- [ ] 浏览器设置字号放大 200% → 布局不错位，侧边栏和内容区不重叠
- [ ] 移动端访问（响应式）→ 侧边栏自动折叠或改为抽屉式，表单控件不溢出

### UI 一致性验收

- [ ] 4 个管理页面的表格列均含 `created_at` 且格式一致（YYYY-MM-DD HH:mm:ss）
- [ ] 新增/编辑弹窗的按钮文案一致：新增="确定"、编辑="保存"
- [ ] 删除操作的 Popconfirm 文案一致："确认删除？"
- [ ] 操作成功后的 message.success 提示一致（2 秒自动消失）
- [ ] 操作失败后的 message.error 提示由 Axios 拦截器统一处理（页面不重复 catch 弹窗）
- [ ] 分页器位置一致：表格右下角，`showSizeChanger: true`
- [ ] 搜索栏 `labelWidth: 'auto'`，搜索/重置按钮对齐

### 表单验收

- [ ] 创建表单：必填项标红星，提交时为空则提示
- [ ] 编辑表单：回显已有数据，主键字段 disabled
- [ ] 提交中按钮 loading，禁止重复点击
- [ ] 提交失败弹窗不关，保留用户输入（不死数据）
- [ ] Modal `destroyOnClose`，关闭后 Form 数据重置
- [ ] 密码输入框 `type="password"`，带眼睛切换明/密文

### 权限按钮验收（逐页）

| 页面 | 权限码 | 预期 |
|------|--------|------|
| 用户列表 | `user:create` | viewer 看不到"新增用户"按钮 |
| 用户列表 | `user:update` | viewer 看不到"编辑"链接 |
| 用户列表 | `user:delete` | viewer 看不到"删除"链接 |
| 菜单管理 | `menu:create` | viewer 看不到"新增根菜单"按钮 |
| 菜单管理 | `menu:update` | viewer 看不到"编辑"/"添加子项"链接 |
| 菜单管理 | `menu:delete` | viewer 看不到"删除"链接 |
| 角色管理 | `role:create` | editor 看不到"新增角色"按钮 |
| 角色管理 | `role:update` | editor 看不到"权限"/"菜单"/"编辑"链接 |
| 角色管理 | `role:delete` | editor 看不到"删除"链接 |
| 系统配置 | `config:update` | viewer 看不到"编辑"链接 |

### 加载顺序验收

- [ ] 登录后首页加载顺序：骨架屏 → 菜单树返回 → 侧边栏渲染 → Dashboard 内容出现
- [ ] 菜单数据未返回时侧边栏不渲染（不闪现空菜单）
- [ ] 用户信息未返回时顶栏显示"加载中…"（非空白）

### 浏览器兼容验收

- [ ] Chrome 最新版：全部功能正常
- [ ] Firefox 最新版：全部功能正常
- [ ] Edge 最新版：全部功能正常
- [ ] Safari 最新版：全部功能正常

---

## M7 — 生产就绪

### Docker 验收

- [ ] `docker-compose up -d` 一键启动 5 个服务（MySQL/Redis/Backend/Frontend/Jaeger）
- [ ] Backend 镜像 < 15MB
- [ ] Frontend 容器 nginx 反代正确，`/api/*` → backend
- [ ] `docker-compose down` 数据不丢失（volume 持久化）
- [ ] `docker-compose.dev.yaml` 启动后修改代码 → 热重载生效

### CI 验收

- [ ] `git push` 触发 GitHub Actions
- [ ] Lint PASS
- [ ] Test PASS（含集成测试）
- [ ] Build PASS → 镜像可用

### 可观测验收

- [ ] `curl localhost:9090/metrics` Prometheus 可抓取
- [ ] Grafana Dashboard 显示 6 个面板（QPS/P99/DB/Redis/GC/Goroutine）
- [ ] Jaeger 可见 trace，每个请求的 span 树完整
- [ ] 健康检查：`/health` 200、`/ready` 200

### 性能验收

- [ ] `/api/v1/users?page=1&page_size=20` P99 < 50ms（c5.large 上压测 30s/100 QPS）
- [ ] `/api/v1/auth/login` P99 < 80ms
- [ ] `/api/v1/menus/tree` P99 < 30ms（缓存命中）
- [ ] 空闲内存 < 50MB

### 文档验收

- [ ] `design/` 目录 50 个文档完整
- [ ] ADR 5 个决策可追溯
- [ ] API 契约每个接口有 curl 示例
- [ ] PROGRESS.md 里程碑全部打勾
- [ ] README.md 含快速启动命令（一行 clone + 一行 docker-compose up）

### 工程验收

- [ ] `docker-compose up -d` 首次启动到可访问 < 60s
- [ ] `docker stats` 中 Backend 容器内存 < 200MB
- [ ] Docker 镜像不含 `.git`、`design/`、测试文件
- [ ] `.dockerignore` 排除不需要的文件
- [ ] GitHub Actions CI 三阶段并行：lint / test / build
- [ ] `make docker-build` 生成镜像后 `docker scout` 无 critical CVE（基础镜像已打补丁）

### 安全最终验收

- [ ] `gitleaks detect` 扫描 git 历史无泄露的 secret
- [ ] `govulncheck ./...` 无已知漏洞
- [ ] 生产 `config.prod.yaml` 中 `server.mode: release`（非 debug）
- [ ] 生产环境 Gin Recovery 中间件不输出 stack trace 到 HTTP 响应（只写日志）
- [ ] 生产环境日志级别为 `info`（非 `debug`）
- [ ] 生产环境 Swagger UI 走 basic auth 或按环境变量开启/关闭

### 备份 & 恢复验收

- [ ] MySQL volume 持久化：`docker-compose down && docker-compose up -d` 数据不丢
- [ ] Redis RDB/AOF 持久化已配置
- [ ] 迁移有 down 脚本：`make migrate-down` 可回滚到上一版本

### 前端生产验收

- [ ] Frontend nginx 配置 gzip 压缩（JS/CSS/HTML）
- [ ] Frontend nginx 配置静态资源缓存（`Cache-Control: max-age=31536000` for hashed assets）
- [ ] Frontend nginx 配置 `try_files $uri /index.html`（SPA fallback）
- [ ] `docker build -f deploy/Dockerfile.frontend .` 镜像 < 50MB
- [ ] Frontend 不直接暴露到公网（仅 Backend nginx 反代对外）

### Chaos & Recovery

- [ ] MySQL 容器被 `docker kill` 杀掉 → Backend Health 检查 /ready 返回 `mysql: down` → K8s readiness probe 失败，流量被摘掉
- [ ] MySQL 容器被 kill 后 10s 恢复 → Backend 重连成功，/ready 恢复 200
- [ ] MySQL 容器被 kill 时正在执行的事务 → 回滚，HTTP 返回 500
- [ ] Redis 被 kill 时 JWT 黑名单检查 → `IsRevoked` 返回 `true, error`（Redis 不可达） → Auth 中间件拒绝请求返回 503（安全优先：宁可全拒，不放过已注销的 token）
- [ ] Redis 恢复后 → 无人工干预自动重连（go-redis 自带重连）
- [ ] 整个 `docker-compose down && docker-compose up -d` → Backend 先于 MySQL 启动 → 重试 3 次后 Fatal → Docker `restart: unless-stopped` 自动重启 → MySQL 就绪后 Backend 启动成功
- [ ] 硬盘满（无法写日志） → Zap 不 block 进程，丢弃日志（或输出到 stderr）
- [ ] 高并发下 GORM 连接池耗尽（100 个连接全部占用） → 新请求等待 conn_max_lifetime 或超时返回 500，非永久 hang
- [ ] 内存泄漏模拟：1000 次登录不退出 → Goroutine 数稳定（无泄漏），内存稳定
- [ ] 跑 10 分钟 Vegeta 压测后停压 → CPU/内存回到基线（无内存泄漏）

---

## 跨里程碑验收项

以下贯穿全部里程碑，每个 M 完成后都要检查。

### 代码质量

- [ ] `golangci-lint run` 无 error（warning 可豁免但需记录原因）
- [ ] 无 `panic` 在业务逻辑中（仅 main.go 启动阶段可用 Fatal）
- [ ] 所有导出函数有文档注释
- [ ] 无 hardcode 的密码/secret/内网 IP
- [ ] 圈复杂度 ≤ 15（单函数），文件行数 ≤ 500

### 测试

- [ ] 单元测试覆盖率 ≥ 80%（`go test -cover ./...`）
- [ ] 单元测试覆盖 Service 层核心方法（M2 起）
- [ ] 集成测试覆盖 Repository CRUD（M2 起）
- [ ] 鉴权相关接口有测试覆盖（M3 起）
- [ ] 前端关键用户流程有 E2E 测试（M6 起）
- [ ] CI 中测试不依赖本地 DB/Redis（用 testcontainers 或 mock）

### 依赖安全

- [ ] `go mod tidy` 无未使用依赖
- [ ] `go mod verify` 校验通过
- [ ] `npm audit` 无 high/critical 漏洞
- [ ] 直接依赖数：Go < 30、前端 < 20（用 `go mod graph | wc -l` 和 `npm ls --depth=0` 检查）

### 日志

- [ ] 关键操作有日志（登录/注册/数据变更/权限变更）
- [ ] 日志不含 password/token/secret 明文
- [ ] 每条日志含 `request_id`（来自中间件）
- [ ] 每条日志含 `trace_id`（来自 OTel）

---

## 自动化策略

所有验收项默认自动化。不接受"人工点一下"的验收方式。

| 测试层 | 覆盖内容 | 文件位置 | CI |
|--------|----------|----------|----|
| `go test -short` | Config 校验、中间件行为、JWT 生成/解析、Service 逻辑、validator、Mock 场景 | `core/**/*_test.go` `extends/**/*_test.go` | 每次 push |
| `go test` + testcontainers | Repository CRUD、事务回滚、缓存读写、迁移执行/回滚 | `extends/**/adapter/**/*_test.go` `test/integration/*_test.go` | 每次 push |
| httptest + curl 脚本 | 所有 API 的 Happy Path + Sad Path + 鉴权 + 限流 | `test/api/*_test.go` `scripts/smoke.sh` | 每次 push |
| Playwright | M5/M6 全部 UI 交互、权限按钮显隐、表单验证、路由守卫、浏览器 Edge Case | `e2e/*.spec.ts` | 每次 PR |
| Bash 自动化脚本 | 启动检查、混沌恢复、镜像构建、安全检查 | `scripts/chaos.sh` `scripts/deploy-check.sh` | 每日定时 / 发布前 |
| `go test -bench` + Vegeta | 性能基准回归 | `scripts/bench.sh` | 每日定时 |

**人工验收项：0 项。**

以下 5 项是曾经被认为"无法自动化"的场景及对应的自动化方案。每项都有对应的脚本或 CI step，不允许人工验证。

| 原"无法自动化"场景 | 自动化方案 | 文件/位置 |
|------|-----------|------|
| Jaeger trace 链路完整性 | Jaeger API `GET /api/traces/{traceID}` → 断言 `spans[]` 中每个 span 的 `references[].refType=CHILD_OF` 链无断 | `scripts/check-traces.sh` |
| Grafana Dashboard 数据连续 | Prometheus API 查询 6 个指标 → 断言序列无超过 60s 的 null 段 | `scripts/check-metrics.sh` |
| CVE 扫描 | `docker scout quickview` CLI 退出码非零 = CI fail | `.github/workflows/ci.yaml` |
| 浏览器 150% 缩放布局 | Playwright `page.evaluate(() => document.documentElement.style.zoom='150%')` + 截图 → 断言 `document.body.scrollWidth <= window.innerWidth` | `e2e/visual.spec.ts` |
| secret 泄露 | `gitleaks detect` → pre-commit hook + CI step | `.github/workflows/ci.yaml`

---

## 自动化测试 Specs

### A. Playwright E2E Specs

> 以下每个 spec 文件对应一个功能模块，覆盖原 Manual Runbook 的所有步骤。CI 中 `npx playwright test` 执行。

#### `e2e/login.spec.ts`

```
场景 1: 未登录自动跳转
  page.goto('/') → expect(url).toContain('/login')

场景 2: 表单校验
  click('登 录') → expect('.ant-form-item-explain').toContain('请输入用户名')
  fill username='a' → expect 校验 '至少 3 个字符'

场景 3: 错误密码
  fill admin/wrong → click 登录 → expect toast '密码错误'

场景 4: 不存在用户
  fill nonexist/anything → click 登录 → expect toast '用户不存在'

场景 5: 登录成功
  fill admin/Abcd1234 → click 登录 → expect url('/dashboard')

场景 6: 限流
  for 6 次: fill admin/wrong → click 登录
  expect toast '请求过于频繁'

场景 7: Tab 键顺序
  press Tab×3 → expect activeElement 依次为 username/password/button

场景 8: Enter 提交
  fill admin/pwd → press Enter → 触发登录

场景 9: 响应式 (mobile viewport)
  page.setViewportSize(400, 800) → expect card 不溢出
```

#### `e2e/layout.spec.ts`

```
场景 1: 侧边栏菜单
  login as admin
  expect sidebar text 'Kingfisher'
  expect menu items: Dashboard, 系统管理
  click 系统管理 → expect 子菜单: 用户管理/菜单管理/角色管理/系统配置

场景 2: 折叠
  click collapse btn → expect sidebar width ~80px, logo 'K', 只有图标
  click expand → expect 恢复文字

场景 3: 顶栏
  expect breadcrumb '首页 / Dashboard'
  expect avatar + 'admin'

场景 4: 面包屑
  navigate to /system/users → expect breadcrumb '首页 / 系统管理 / 用户管理'

场景 5: 刷新保持登录
  page.reload() → expect url 不变, sidebar menu 展开到当前页

场景 6: 退出
  click avatar → click '退出登录' → expect url('/login')
```

#### `e2e/logout-guard.spec.ts`

```
场景 1: Auth Guard
  page.goto('/dashboard') (未登录) → expect 跳转 /login

场景 2: redirect 参数
  page.goto('/system/users') → expect url 含 redirect=%2Fsystem%2Fusers
  登录 → expect url('/system/users')

场景 3: 多 Tab
  browser.newContext() → 两个 context 各自登录不同用户
  context1 退出 → expect context1 跳转登录页, context2 不受影响

场景 4: 退出后直接访问
  退出 → page.goto('/dashboard') → expect 跳转 /login
```

#### `e2e/user-crud.spec.ts`

```
场景 1: 列表展示
  login as admin → click 用户管理
  expect table columns: ID/用户名/邮箱/角色/状态/创建时间/操作
  expect status 启用=绿色 Badge, 禁用=红色 Badge

场景 2: 搜索
  fill 搜索框 'admin' → click 搜索 → expect rows=1, 含 'admin'
  click 重置 → expect rows > 1

场景 3: 新增
  click '新增用户' → expect modal title '新增用户'
  不填提交 → expect 必填项红字提示
  填写 username/Password1/email → 提交 → expect toast '创建成功', table 刷新含新行

场景 4: 编辑
  click row '编辑' → expect modal 回显数据, username disabled
  修改 email → 提交 → expect toast, table email 列更新

场景 5: 删除
  click row '删除' → expect Popconfirm → 确认 → expect toast, 该行消失

场景 6: 分页
  click page 2 → expect url 含 ?page=2, 数据切换
  切换 pageSize 50 → expect rows≤50, url 含 page=1

场景 7: 搜索无结果
  fill keyword 'zzzzzxyz' → expect empty state '没有找到匹配的结果'
```

#### `e2e/permission-btn.spec.ts`

```
场景 1: viewer 无权
  login as viewer → expect sidebar 仅 Dashboard
  page.goto('/system/users') → expect 无 '新增用户' btn, 无 '编辑'/'删除' link

场景 2: editor 部分权
  login as editor → 用户管理: 有 '新增''编辑', 无 '删除'
  菜单管理: 有 '新增根菜单''编辑''添加子项', 无 '删除'
  角色管理: 无 '新增角色''权限''菜单''编辑''删除'

场景 3: admin 全权
  login as admin → 所有按钮可见

场景 4: 按钮不渲染（不是 disabled）
  expect(PermissionBtn 无权限时 child 不出现在 DOM)
```

#### `e2e/menu-manage.spec.ts`

```
场景 1: 树形表格
  login as admin → click 菜单管理
  expect 每行有缩进（level 通过 paddingLeft 验证）
  expect Tag 颜色: 目录=blue, 菜单=green, 按钮=orange

场景 2: 新增根菜单
  click '新增根菜单' → expect TreeSelect 为空
  fill 名称/选目录 type/填 icon → 提交 → expect 表格更新

场景 3: 添加子项
  click 某目录行 '添加子项' → expect TreeSelect 已选中父级
  填写 → 提交 → expect 新行缩进更深

场景 4: 删除有子节点
  click 有子节点的行 '删除' → expect toast '有子节点'
```

#### `e2e/role-permission.spec.ts`

```
场景 1: 权限弹窗
  login as admin → 角色管理 → click admin 行 '权限'
  expect modal + 4 Tabs (用户/菜单/角色/配置)
  admin 各 Tab Checkbox 均已勾选

场景 2: 分配权限生效
  取消 viewer 的 user:list → 保存
  login as viewer → page.goto('/system/users') → expect 403 或内容不可见

场景 3: 菜单分配
  click admin 行 '菜单' → expect Tree 全部勾选
  取消勾选 '系统管理' → expect 子节点一并取消

场景 4: 菜单分配生效
  login as admin → expect 侧边栏仅 Dashboard(无系统管理)
```

#### `e2e/system-config.spec.ts`

```
场景 1: 配置列表
  login as admin → 系统配置 → expect 4 行

场景 2: 编辑生效
  编辑 site_name='My Sys' → 保存
  page.goto('/login') → expect title 含 'My Sys'

场景 3: 登录限流动态生效
  编辑 max_login_attempts=3 → 保存
  登录失败 4 次 → expect 429
  恢复 max_login_attempts=5
```

#### `e2e/browser-edge.spec.ts`

```
场景 1: 后退保留状态
  用户列表第 2 页 → 点击编辑 → 弹窗打开 → page.goBack()
  → expect 弹窗关闭, 列表仍在第 2 页, 搜索框值保留

场景 2: ESC 关闭弹窗
  弹窗打开 → press Escape → expect 弹窗关闭, 表单重置

场景 3: 弹窗外点击关闭
  弹窗打开 → click modal mask → expect 关闭

场景 4: Tab 焦点循环
  弹窗内 → press Tab×10 → expect activeElement 始终在 modal 内 (非 body 元素)

场景 5: 多 Tab 数据同步
  context1 新增用户 → context2 刷新列表 → expect context2 看到新用户

场景 6: 移动端 viewport
  setViewportSize(375, 812) → expect sidebar 隐藏或 drawer, 表格可横向滚动

场景 7: beforeunload
  编辑弹窗有未保存内容 → page.close({runBeforeUnload: true})
  → expect dialog text '未保存'
```

---

### B. Bash 自动化脚本

> 所有脚本完整实现见 → [scripts 设计文档](../scripts/design.md)

| 脚本 | 说明 |
|------|------|
| `scripts/chaos.sh` | MySQL/Redis 中断 + 恢复测试，6 个场景 |
| `scripts/deploy-check.sh` | 镜像大小、生产配置、安全头、隐藏文件保护 |
| `scripts/bench.sh` | Vegeta 30s/100QPS 压测，CI 模式 P99 > 50ms → fail |
| `scripts/check-traces.sh` | Jaeger API 拉取 trace→验证 span 层级 ≥ 3 |
| `scripts/check-metrics.sh` | Prometheus API 验证 6 个指标均有数据 |
| `scripts/check-no-panic.sh` | grep 扫描 panic/log.Fatal/fmt.Println |
| `scripts/check-frontend-constraints.sh` | grep 扫描 hardcode URL、直接 fetch/axios |

```
---

### C. 代码层硬性约束

→ 详见 [guardrails 设计文档](../backend/guardrails/design.md)

13 条编译/Lint 约束在 CI 中强制通过。违反任意一条，整个 pipeline 阻断。包括：

| # | 约束 | 执行层 |
|---|------|--------|
| 1 | `core/` 不得 import `extends/` | depguard (golangci-lint) |
| 2 | `domain/` 零外部依赖 | depguard |
| 3 | 所有 error 必须检查 | errcheck |
| 4 | context 必须是第一个参数 | revive |
| 5 | 导出函数必须有注释 | revive |
| 6 | 禁止 `panic` 在非 main.go 中 | grep check |
| 7 | 禁止 `log.Fatal` 在非 main.go 中 | grep check |
| 8 | `tsc --noEmit` 零 error | TypeScript strict |
| 9 | 禁止 `any` 类型 | ESLint |
| 10 | 禁止 `console.log` | ESLint |
| 11 | 禁止组件中直接 `fetch`/`axios` | grep check |
| 12 | 测试覆盖率 ≥ 80% | CI gate |
| 13 | Wire 生成代码与提交一致 | CI gate |

