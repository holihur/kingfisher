# Kingfisher 项目评审 — Round 2（全量重新审计）

> 评审日期：2026-07-31
> 评审对象：`design/` 全部 55 个文件（52 份 design.md + `backend/audit/gap-report.md` + `scripts/design.md`）+ `PROGRESS.md`
> 方法：逐份通读 + 交叉核对（错误码/路由/迁移/种子/接口签名/技术栈/验收数字），并对全部相对链接做了有效性扫描

---

## 一、结论速览

| 评分对象 | 分数 | 一句话结论 |
|----------|------|------------|
| 设计文档（本交付物） | **85 / 100** | 比 round1 更完整（P0 缺口已系统性关闭），但新增内容引入了更多交叉矛盾 |
| 项目整体（参考） | 65 / 100 | 仍为纯设计阶段，0 行实现 |

> round1 = 87 分；本轮 -2 分。原因：缺口补齐（+）与一致性恶化（-）相抵后略降。

---

## 二、本轮相对 round1 的变化（新增文件与修订）

| 项目 | 位置 | 说明 |
|------|------|------|
| 操作审计模块 | `design/backend/extends/audit/design.md` | 新增第 5 个 extends 模块（异步批量写 + GET /audit-logs） |
| 自审缺口报告 | `design/backend/audit/gap-report.md` | 7 个 P0 缺口全部标记"已修复"，含降级矩阵、关闭追踪表 |
| 验收脚本 | `design/scripts/design.md` | 补齐 7 个可执行脚本（chaos/deploy/bench/traces/metrics/no-panic/frontend） |
| 生命周期接口 | `backend/core/design.md` | `Module` 接口新增 `Init/Shutdown` |
| Gzip 中间件 | `backend/middleware/design.md` | Core 中间件 6→7 个 |
| /version 端点 | `backend/startup/design.md` | 新增构建信息端点 |
| 迁移 | `backend/migration/design.md` | 8→10 个迁移（000009 session_version、000010 audit_logs） |
| 踢人机制 | `extends/user` + migration | session_version 方案 |

**正面评价**：gap-report 的"排查 → 修复 → 关闭追踪"工作方法本身是优秀的工程实践，且修复方向（生命周期、审计、踢人、version、降级矩阵）都是合理选择。scripts 补齐后，round1 中"脚本只见于 acceptance"的缺口已解决。

---

## 三、评分明细

| 维度 | 权重 | 得分 | 依据 |
|------|------|------|------|
| 架构与模块设计 | 25 | 23 | Core+Extends 仍是高质量；但 Module/Register 命名在 overview 与 core 间未统一 |
| 可执行性 | 20 | 17 | 文件清单可执行；但 wire 签名冲突、main.go 引用未定义变量、前端依赖不存在的端点，执行会卡住 |
| 验收与测试设计 | 20 | 17 | 415 条验收 + 脚本级实现顶级；但 M3 Chaos 与降级矩阵自相矛盾，多处数字过时 |
| 完整度 | 20 | 17 | P0 缺口关闭；但前端审计页、/users/me、权限端点、seed 配置仍缺 |
| 一致性与细节 | 15 | 10 | 跨文档矛盾 20+ 处，是本轮主要扣分点 |
| **合计** | **100** | **85** | |

---

## 四、一致性问题清单（按严重度分组）

### A. 错误码 / HTTP 映射（会导致验收直接失败）

1. `backend/errcode/design.md` 的 `HTTPStatus()`：`code >= 10100 → 400`，因此 **10107（登录失败超限）映射为 400**；但 `api-contract` 和 acceptance 多处要求 **10107 → HTTP 429**。
2. `ErrMethodNotAllowed = 10008` 定义存在，但 `HTTPStatus()` 未映射（落入默认 500）；acceptance M1 要求"无 handler 路由返回 **405**"。
3. `security` / `validation` 文档引用的 `ErrPasswordTooShort / ErrPasswordTooLong / ErrPasswordWeak` 在错误码表中**不存在**。

### B. 端点契约（前后端对不上）

4. `GET /api/v1/users/me` 在 api-contract 列出、前端 `fetchUserInfo` 依赖，但 `extends/user` 的 Handler 清单**无 GetMe**。
5. `GET /api/v1/roles/:id/permissions`、`GET /api/v1/roles/:id/menus` 在 api-contract 与前端弹窗（`roleApi.getRolePermissions/getRoleMenus`）使用，但 `extends/rbac` 只有 PUT 分配，RoleService 也无对应读取方法。
6. 前端 auth store 登录后调用 `userApi.getPermissions()` —— 该接口在**前后端文档中都不存在**（后端只有管理员的 GET /permissions，不是当前用户权限）。
7. 创建/编辑用户带角色：前端表单送 `role`（code 字符串），后端 `CreateUserReq` 无 role 字段、`UpdateUserReq` 是 `role_id *uint` —— 字段与类型都接不上。
8. 分页参数：前端 API 层统一发 `pageSize`（camelCase），后端契约统一 `page_size` —— 全链路分页失效。
9. `DELETE /api/v1/configs/:key` 后端有、前端有，api-contract 的 Config 清单却漏了。

### C. 迁移与种子数据（验收数字全面过时）

10. 审计迁移编号：`extends/audit` 与 `gap-report` 用 `000009_create_audit_logs`，`migration/design.md` 用 `000010`（000009 给了 session_version）—— 编号冲突。
11. 权限数量：acceptance 与种子仍写"14 个权限 / admin 14 / editor 8 / viewer 4"，audit 新增第 15 个（`audit:list`，admin 独有）未同步；`000008_seed_data.down.sql` 也只清理 1–14。
12. 配置项：`extends/config` 预设 **5** 项（含 session_timeout），种子与 acceptance 只有 **4** 项。
13. acceptance M4 迁移验收"8 个 SQL 文件（7 建表 + 1 种子）"、`PROGRESS.md`"migrations/ (8 个 SQL 文件)" —— 实际已 10 个迁移、8 张表。
14. 种子角色 ID 为 **1, 3, 4**（跳号 2），前端 rbac 表格示例画的是 1, 2, 3。
15. M6 验收"4 组 14 个权限全部勾选"、权限弹窗 4 个 Tab —— 均未含 audit 组；M5/M6/PROGRESS 也没有审计日志页面（gap-report 声称 M6 实现，但无任何设计文件）。

### D. 架构 / 组装（执行时会直接踩坑）

16. **wire 签名冲突**：`di/design.md` 定义 `InitializeApp(configPath string)`；`startup/design.md` 调用 `wire.InitializeApp(cfg, logger, db, rdb, jwtMgr)`。
17. `startup/design.md` 的 main.go 在步骤 8 用 `r.GET("/version", ...)` 注册路由，但该函数**从未定义变量 r**；且"// 9."步骤编号重复。
18. 模块接口命名未统一：`overview` 仍是旧的 `RouteRegistrar`（含 RegisterAdmin）+ `RegisterModule`；`core`/`di`/acceptance 已是 `core.Module` + `Register`。
19. 中间件顺序三处不一致：M1"Trace→Recovery"、`overview` engine（无 Gzip/RateLimit）、`middleware` 文档（Recovery→Trace，7 个中间件）。M1"6 个中间件"、acceptance"core 7 个子包"均已过时（实际 10 个子包、7 个中间件）。
20. `transaction/design.md` 的 UserService 依赖 `roleRepo`（来自 extends/rbac），acceptance 却验收"`extends/user` 不 import `extends/rbac`"—— 直接冲突。
21. 目录路径系统性漂移：`perf-bench`、`swagger-checklist`、`guardrails`、`test` Makefile、`scripts/design.md` 均写 `internal/extends/...`，而实际结构是根级 `extends/`（`internal/` 只有 wire）。
22. 命名漂移：`NewMySQL` vs `NewGorm`；`NewRedis` vs `NewRedisClient`；observability 写 `pkg/telemetry/`（实际 core/telemetry）；`core.SetLogger()` 全局函数与 DI 方案并存。

### E. JWT / session_version（缺口未闭环）

23. `jwt/design.md` 的 `Claims` 无 `SessionVersion`（sv）字段，`GenerateToken` 签名无 sv 参数，登录流程也未获取 session_version —— 但 `extends/user` 的 Auth 中间件要求校验 `claims.SessionVersion`。三个文档之间接不上。
24. cache key 规范表没有 `user:sv:{userID}`（`extends/user` 实际使用）。

### F. 降级策略自相矛盾

25. M3 Sad Path 与 gap-report 矩阵："Redis 不可用 → RBAC **拒绝返回 503**（安全优先）"；M3 Chaos 却写"缓存 Get 失败 → **回源 DB 查询权限成功 → 请求正常**"。chaos.sh 脚本断言 503。三个来源两种结论，必须取其一。

### G. 前端设计

26. `any` 泛滥：`request/rbac/config/menu/state-ui/auth` 六份文档的示例代码大量使用 `any`（含泛型 `<any, ApiResponse<...>>` 和 `catch (err: any)`），直接违反 guardrail 第 9 条；request 文档甚至把 `any` 写进"类型安全"设计要点。
27. 环境变量名不一致：`overview/request` 用 `VITE_API_BASE_URL`，`local-dev` 用 `VITE_API_TARGET`。
28. 侧边栏路径拼接 bug：`layout/design.md` 用 `parentPath + item.path`，而种子数据的子菜单 path 是绝对路径（`/system/users`）→ 会拼成 `/system/system/users`。
29. 配置页展示 `remark`，但后端 `GetAll` 返回 `map[string]string`（无 remark）—— 前端"从另一个接口补充"不存在。
30. 组件文件缺条目：M6 未列 `IconSelect.tsx`；overview 有 `router/dynamic.tsx`、`stores/app.ts`、`utils/permission.ts` 但 PROGRESS 未列；登录页有"去注册"链接但 M5 无注册页。

### H. 部署 / 可观测

31. deploy compose 只有 mysql/redis/app/jaeger/prometheus，**无 frontend、无 grafana**；acceptance M7 验收"一键启动 5 个服务（…Frontend…）"、M7 验证命令 `open localhost:3000`（Grafana）。
32. 前端 Dockerfile + nginx.conf 是 M7 产出文件，但 deploy 文档只有后端 Dockerfile 与参考 nginx 片段。
33. CI 三处不一致：deploy 文档是精简版（lint/test/build）；guardrails 文档是 8 层 backend + frontend + deploy-check；acceptance 声称 415 条全部由 CI 强制执行。
34. 迁移执行策略冲突：bootstrap 说"开发自动、生产手动"，deploy compose 却把 `./migrations` 挂到 `/docker-entrypoint-initdb.d` 首次自动执行。
35. `scripts/design.md` 的 bench.sh CI 段是占位伪代码（`echo "..." | vegeta report`），与"完整可执行"声明不符；check-no-panic.sh 同样用错 `internal/` 路径。

---

## 五、亮点（本轮新确认）

- gap-report 的 **P0/P1/P2 分级 + 关闭追踪表**（#1–23 逐项标注状态）是高质量的自审方法，值得其他项目复制。
- **降级矩阵**把 /health、/ready、公开接口、RBAC 接口、缓存、黑名单在 MySQL/Redis/Jaeger 故障下的行为集中成一张表（除与 M3 Chaos 一处冲突外，决策本身正确：安全优先于可用性）。
- scripts 齐全且可直接运行（chaos.sh 6 场景、deploy-check、bench、traces、metrics），比 round1 时的"只有引用"前进一大步。
- 迁移/种子 SQL 已写到可执行级（含 down 回滚、自增重置），审计表索引设计合理。

---

## 六、优先修复建议（评审只记录，不代改）

按"不改就会在执行/验收时立刻失败"排序：

1. **错误码映射**：`10107 → 429`、`10008 → 405`，补齐密码策略错误码（A 组 1–3）。
2. **补三个后端端点或删前端依赖**：`GET /users/me`、`GET /roles/:id/{permissions,menus}`、`GET /me/permissions`（B 组 4–6）。
3. **统一 wire 签名 + 修 startup main.go**：`InitializeApp(configPath)` 二选一，补 `r` 定义（D 组 16–17）。
4. **迁移/种子定版**：audit 统一为 000010；权限 15 与 `session_timeout` 入种子；down.sql 同步清理；acceptance/PROGRESS 数字更新为 10 迁移/8 表/15 权限（C 组 10–13）。
5. **前端契约**：分页改 `page_size`；创建/编辑用户角色字段统一为 `role_id`；侧边栏路径拼接改相对子路径（B 组 7–8、G 组 28）。
6. **JWT Claims 补 `SessionVersion`**，与踢人方案闭环（E 组 23）。
7. **降级策略取一**：M3 Chaos"回源 DB 正常"与矩阵/chaos.sh"503"只能留一个（F 组 25）。
8. **前端示例代码去 `any`**，让 guardrail #9 首先在自家文档成立（G 组 26）。
9. **deploy compose 补齐 frontend/grafana** 或同步删减 acceptance 项（H 组 31–32）。
10. 统一 `internal/extends` → `extends` 路径、`Module` 接口命名、中间件顺序描述（D 组 18–19、21）。

---

## 七、结论

- 设计整体仍是**高质量蓝图**：架构方向正确、验收体系顶级、缺口管理成体系，作为"设计阶段交付物"给 **85/100**。
- 本轮分数略降的核心原因是：**每修一个缺口就引入一处新矛盾**（编号、签名、数量、策略），说明文档缺少"定版/同步"机制。
- 建议下一轮评审前，先跑一遍本报告第六节 10 项修复；全部闭环后设计分可稳定在 90+，届时即可安全进入 M1 实现。

## 八、评审记录

- 通读范围：全部 52 份 design.md + gap-report.md + scripts/design.md + PROGRESS.md。
- 交叉核对维度：错误码表与验收、api-contract 与各模块 Handler、迁移与种子、wire/startup 组装、中间件顺序、降级策略、前端 API 层与后端契约、部署清单与验收。
- 链接扫描：2 处坏链接仍存在（`M6 → ../../shared-types`、`extends/rbac → ../middleware`），acceptance 指向 `scripts/design.md` 已随本轮新增而闭合。
