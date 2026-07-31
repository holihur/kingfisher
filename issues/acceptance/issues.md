# Acceptance 验收项 — 设计与实现差异

> 来源：`design/acceptance/design.md`（1020 行验收清单）对照当前实现
> 排查日期：2026-07-31 ｜ 约定：P0=验收阻断（必须修复），P1=重要差异，P2=轻微/建议

## P0 · 功能验收失败

### A-1 `GET /metrics` 未实现
- 设计：M1 要求 `/metrics` 返回 Prometheus 格式，至少含 `http_requests_total`
- 实现：`core/telemetry/` 为空目录，`cmd/server/main.go` 只注册 `/version /health /ready`，无 `/metrics`
- 影响：M1 自动化验收第一项即失败；observability 全部指标/追踪缺失

### A-2 viewer 访问 `/api/v1/users` 不会 403
- 设计：M3 关键验收「viewer 登录 → GET /api/v1/users → 403」
- 实现：`extends/user/transport/register.go` 的 users 路由未挂 `RequirePerm`；RBAC 中间件虽注入 permissions，但只有 roles 路由做了 `RequirePerm` 控制（`extends/rbac/transport/register.go`）
- 影响：鉴权验收核心场景全部失败，viewer/editor 可读写用户/菜单/配置

### A-3 `PUT /api/v1/configs/:key` 无权限校验
- 设计：M3 验收「viewer 登录 → PUT /api/v1/configs/site_name → 403」
- 实现：`extends/config/transport/register.go` 未挂 `RequirePerm("config:update")`
- 影响：viewer 可修改系统配置

### A-4 菜单树未按角色过滤
- 设计：M3 验收「editor 登录 → GET /menus/tree 只返回有权限的菜单；viewer 侧边栏只剩 Dashboard」
- 实现：`extends/menu/app/service.go` 的 `GetTree` 无角色过滤，返回全部菜单
- 影响：前端侧边栏对所有角色渲染全部菜单；M3 关键验收失败

### A-5 RBAC 权限缓存命中返回空权限
- 设计：权限缓存 30min，命中回读缓存
- 实现：`extends/rbac/app/service.go` 的 `strSlice()` 是占位符（恒返回 nil）；Redis 可用时第二次请求起 `GetUserPermissions` 命中缓存返回空权限 → 所有接口 403
- 影响：启用 Redis 后整个后台全部 403；P0 级功能故障

### A-6 RBAC 中间件 Redis 不可用时行为与设计相反
- 设计：M3 Sad Path「RBAC 中间件 Redis 不可达/超时 → 拒绝请求返回 503（安全优先）」
- 实现：`extends/rbac/transport/middleware.go` 中 `perms, _ := roleSvc.GetUserPermissions(...)` 忽略 error，返回空权限 → 403 而非 503；同时缓存命中时 strSlice 返回 nil（见 A-5）
- 影响：故障语义与验收不符；且因忽略错误掩盖了缓存 bug

### A-7 `POST /api/v1/auth/logout` 缺失
- 设计：api-contract 要求 logout；JWT 黑名单 RevokeToken 已实现
- 实现：`extends/user/transport/register.go` 未注册 `/auth/logout`，`RevokeToken` 无调用方
- 影响：注销 token 功能不可用

### A-8 `GET /api/v1/users` 不返回角色
- 设计：api-contract 用户列表响应含 `role: {id,name,code}`
- 实现：`extends/user/adapter/mysql/repo.go` 的 `FindAll` 不关联 roles 表；前端用户列表无「角色」列（设计 M6 列：ID/用户名/邮箱/角色/状态/创建时间/操作）
- 影响：用户管理页面缺角色列；契约不一致

### A-9 删除用户 500（软删除未实现）
- 设计：M2 数据验收要求软删除（`deleted_at`），删除后同名注册需清理 `deleted_at` 才能重用
- 实现：`core/database/models.go` 的 `UserPO` 无 `DeletedAt` 字段（AutoMigrate 不建该列）；`extends/user/adapter/mysql/repo.go` 的 `Delete` 却执行 `Update("deleted_at", &now)` → SQLite 报 `no such column` → 500
- 影响：删除用户功能不可用；设计「软删除 + 同名可重用」语义未落地

### A-10 前端「新增用户」接口 404
- 设计：M6 验收「点击新增用户 → 提交 → 列表刷新」
- 实现：`web/src/api/user.ts` 调 `POST /users`，但后端 `extends/user/transport/register.go` 未注册 `POST /users`（api-contract 也无此接口——设计侧同样缺失）
- 影响：用户管理「新增」端到端不可用；设计与实现共同缺口

## P1 · 重要差异

### A-11 `/ready` 不返回 503
- 设计：MySQL/Redis down 时 `/ready` 返回 503 + `mysql/redis: down`
- 实现：`cmd/server/main.go` readyHandler 恒 200，只改字段 `"down"`
- 影响：健康检查语义错误，chaos 脚本 `assert_eq 503` 会失败

### A-12 MySQL 启动无重试
- 设计：M1「MySQL 不可用 → 重试 3 次后 Fatal，日志含每次重试间隔」
- 实现：`core/database/gorm.go` 一次失败直接返回 error → main Fatal
- 影响：Docker 编排下 backend 先于 MySQL 启动会立即退出（依赖 restart 策略兜底）

### A-13 Redis 不可用降级而非 Fatal
- 设计：M1「Redis 连接超时 → Fatal，不跳过（核心依赖）」
- 实现：`cmd/server/main.go` warn + `redisCache = nil` 降级运行
- 影响：与验收及 bootstrap 设计冲突（实现选择了可用性优先）

### A-14 `JWT_SECRET` 校验仅限 release
- 设计：M1「JWT_SECRET 为空或 = change-me-in-production → 启动拒绝」
- 实现：`core/config/config.go` 仅在 `Server.Mode == "release"` 时拒绝占位符
- 影响：debug 模式带默认密钥启动，验收场景失败

### A-15 请求体 10MB 限制未实现
- 设计：M1「请求体 > 10MB 返回 413」；security 设计「Recovery 中 MaxBytesReader」
- 实现：无任何 `MaxBytesReader`；`server.max_request_body` 配置字段未被使用
- 影响：超大请求无保护

### A-16 密码强度校验未实现
- 设计：security/validation「密码必须 ≥8 位且含大小写+数字（10108/10109/10110）」
- 实现：`RegisterReq`/`ChangePwdReq` 仅 `min=8,max=64`，无强度校验
- 影响：弱密码可注册

### A-17 注册限流与设计不一致
- 设计：security「注册限流 1 次/5min per IP + 同邮箱 3 次/h」
- 实现：`extends/user/transport/register.go` 注册挂 `RateLimit(2, 5min)`；无邮箱级限流
- 影响：垃圾注册防护弱于设计

### A-18 登录后 token 角色硬编码 viewer
- 设计：`GenerateToken(user.ID, user.Role)`，token 携带真实角色
- 实现：`extends/user/app/service.go` Login 调 `GenerateToken(ctx, user.ID, "viewer", ...)` 写死 viewer
- 影响：admin/editor 登录后 token 里 role=viewer；虽权限走 DB 查询，但 token 语义错误，`c.GetString("role")` 全为 viewer

### A-19 Auth 中间件不校验 session_version
- 设计：M2「修改密码/踢人后旧 token 立即失效（sv 校验）」
- 实现：`extends/rbac/transport/middleware.go` AuthMiddleware 只注入 user_id/role，不比对 `user:sv:*`
- 影响：RevokeSessions/ChangePassword 的踢人逻辑形同虚设

### A-20 审计日志零写入
- 设计：M4「审计日志记录所有操作；audit 中间件」
- 实现：`extends/audit/transport/` 无 middleware.go；所有 handler（user/menu/role/config）都不调用 `AuditService.Log`；`AuditService.worker` 无输入数据
- 影响：审计表永远为空，合规验收失败

### A-21 菜单删除/创建业务校验缺失
- 设计：M4「有子节点返回 10203；path 重复返回 10201；parent_id 不存在返回参数错误」
- 实现：`extends/menu/app/service.go` Delete 直接删、Create 无 path 唯一性/父级校验；`MenuRepo.HasChildren` 未使用
- 影响：脏数据（孤儿子节点、重复 path）

### A-22 角色业务校验缺失
- 设计：M3「code 重复→10301；删除有用户角色→10303；分配不存在 permission→参数错误；角色层级保护」
- 实现：`extends/rbac/` Create/Delete/AssignPermissions 均无上述校验，handler 忽略 service error
- 影响：重复 code 返回 200、可删在用角色、可分配不存在权限

### A-23 分页/排序语义与验收不符
- 设计：M4「page=-1→10001；page_size=1000→截断 100；sort 白名单校验」
- 实现：`extends/user/transport/handler.go` page<1→1（静默修正）、pageSize>100→20（非 100）；无 sort 支持
- 影响：边界行为与验收不一致

### A-24 迁移与 Wire 完全缺失
- 设计：M4「migrations/ 10 个 SQL 文件；make migrate-up/down；internal/wire + wire_gen.go；make wire」
- 实现：`migrations/` 空目录；无 `cmd/migrate`；`internal/wire` 不存在（Makefile 有 `wire` 目标但会失败）；`make swagger` 目标不存在
- 影响：生产 MySQL 部署无迁移入口；M4 验收大面积失败

### A-25 Swagger 未落地
- 设计：M4「make swagger → /swagger/index.html 可见全部接口」
- 实现：无 swag 注解（handler 均无 `@Summary/@Success/@Router`）、无 `docs/`、无 `/swagger` 路由、Makefile 无 swagger 目标
- 影响：Swagger 验收全部失败；shared-types 也依赖 swagger.json

### A-26 审计日志 PUT/DELETE 返回 404 而非 405
- 设计：M6「PUT/DELETE /audit-logs/:id → 405」
- 实现：`extends/audit/transport/register.go` 只注册 GET；PUT/DELETE 无路由 → 404
- 影响：验收期望 405，实际 404

### A-27 前端 404/403 页面缺失
- 设计：M5/M6「未定义路由 → 404 页；无权限 → 403 页」
- 实现：`web/src/router/index.tsx` `*` 路由跳 `/login`；无 403 页面
- 影响：直接访问 `/abc` 被重定向到登录而非 404

### A-28 路由懒加载未实现
- 设计：M5「各页面独立 chunk，首屏只加载当前页」
- 实现：`web/src/router/index.tsx` 全部静态 import，单 bundle
- 影响：首屏体积大；验收「独立 chunk」失败

### A-29 组件库缺失
- 设计：M5「Skeleton/EmptyState/ErrorResult 三种状态组件 + IconSelect」
- 实现：`web/src/components/` 仅有 `PermissionBtn.tsx`
- 影响：组件库验收失败；页面错误/空状态无统一兜底

### A-30 站点配置不生效
- 设计：M6「编辑 site_name → 登录页/页头标题变化；max_login_attempts → 登录上限变化」
- 实现：`web/src/pages/login/index.tsx` 标题固定「Kingfisher」；后端限流写死 `5`（`extends/user/transport/register.go` RateLimit(5, 1min)），不读 `system_configs`
- 影响：配置驱动 UI/行为未实现

## P2 · 轻微/建议

- A-31 `gzip` 中间件为空实现（`core/router/engine.go` 占位 `c.Next()`），未实际压缩、未排除 /metrics /health
- A-32 SecurityHeaders 缺 `Content-Security-Policy / Cache-Control / Pragma`（`core/middleware/middleware.go`）
- A-33 RateLimit 用 INCR 计数器而非设计要求的 ZSET 滑动窗口；无 `X-RateLimit-Limit/Remaining` 响应头；全局限流 `requests_per_minute` 未接入 engine
- A-34 中间件未按设计拆分为 8 个独立文件（request_id/trace/recovery/logger/gzip/security_headers/cors/ratelimit），全部在 `middleware.go`；RBAC 中间件也应为 3 个文件（auth/rbac/require_perm）
- A-35 Trace 未创建 OTel span，仅透传/生成 `X-Trace-ID`；日志无 `trace_id` 注入
- A-36 登录时 Redis 不可用 → 限流失效可接受（实现符合 design 的降级说明），但 A-6 的 503 语义仍不符
- A-37 `Refresh` 时 Redis 不可用 → 黑名单检查忽略 error（`core/jwt/jwt.go` `revoked, _ :=`），设计要求拒绝刷新 503
- A-38 菜单树缓存（`menu:tree` 10min）、配置缓存（write-through）、用户信息缓存、空值缓存均未实现
- A-39 LIKE 关键词未转义 `%`/`_`（`extends/user/adapter/mysql/repo.go` `FindAll`），验收要求转义
- A-40 用户管理搜索无防抖；多 Tab 数据协同未实现（M6 浏览器 Edge Case）
- A-41 `document.title` 路由联动未实现；面包屑仅「首页/最后一段」，设计要求完整层级
- A-42 响应式 <768px 侧边栏抽屉未实现；MenuManage TreeSelect 数据为空（`treeData={[{根菜单}]}`），父级选择不可用
- A-43 前端「删除自己」按钮无 disabled+Tooltip（仅后端拒绝）
- A-44 URL 状态同步部分实现（ProTable `syncToUrl`），无 `replaceState` 语义、无 expanded 参数同步
- A-45 登录 redirect 回跳未实现（`AuthGuard` 不带 `?redirect=`，登录后固定跳 /dashboard）
- A-46 注册页/路由缺失（登录页无「去注册」链接）
- A-47 seed 配置数矛盾：acceptance 说「4 个预设配置」，`core/database/gorm.go` 种子为 5 个（含 session_timeout），migration 设计也是 5 个——设计文档内部不一致
- A-48 设计文档中 admin 密码 hash `$2a$12$LJ3m...` 实测不匹配 `Abcd1234`；实现种子 hash `$2a$12$jDyI...` 实测匹配——设计文档 seed SQL 有误
- A-49 覆盖率 ≥80%、`npm audit`、gitleaks、golangci-lint CI gate、Playwright、chaos/bench/deploy-check 脚本等「自动化策略」承诺无落地载体（见 `issues/scripts`、`issues/backend/test`、`issues/backend/deploy`）
