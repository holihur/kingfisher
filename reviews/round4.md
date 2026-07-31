# Kingfisher 项目评审 — Round 4（增量核查）

> 评审日期：2026-07-31
> 评审对象：round3 之后修改的 7 个文件（errcode、extends/user、frontend/request、M1-core-startup、deploy、migration、PROGRESS）
> 方法：按 mtime 识别改动文件，逐项对照 round3 未落地清单验证

---

## 一、结论速览

| 评分对象 | 分数 | 一句话结论 |
|----------|------|------------|
| 设计文档 | **87 / 100** | 本轮 7 项真实修复，质量在回升；但 17 项积压未动，且新增 1 处前端编译级断裂 |
| 项目整体（参考） | 65 / 100 | 仍为纯设计阶段 |

> round2=85 → round3=86 → round4=87。修复开始进入"改到点子上"的节奏，但修复面仍集中在后端与 M1，前端/种子/CI 三块几乎没动。

---

## 二、本轮已核实修复（7 项）

| 项 | 对应 round3 | 修复内容 | 证据 |
|----|-------------|----------|------|
| 1 | 新问题 #1 | 新增 `ErrServiceUnavailable = 10009`，`HTTPStatus` 映射 **503** | `errcode/design.md` L40/L79-80 |
| 2 | #4 | `extends/user` 新增 `GetMe` → `GET /users/me` | L134/L161 |
| 3 | #6（后端侧） | 新增 `GetMyPermissions` → `GET /users/me/permissions` | L135/L162 |
| 4 | #7（API 层） | `create/update` 改用 `role_id: number`，不再送 role code | `request/design.md` L147-150 |
| 5 | #26（request 部分） | request 文档 `any` 全部清零，改为类型化泛型 + 引用 `@/types/api.generated` | 全文无 `any` |
| 6 | #19（M1 部分） | M1 中间件顺序与数量更新为 8 个（含 SecurityHeaders） | `M1-core-startup/design.md` L19/L46 |
| 7 | 新问题 #4 | compose frontend 改用 `deploy/Dockerfile.frontend` | `deploy/design.md` L84 |
| 8 | #13（PROGRESS 部分） | PROGRESS 迁移数更新为 10 个 SQL | `PROGRESS.md` L107 |

---

## 三、延续未修复（round3 遗留，17 项）

| 组 | 问题 | 证据（本轮文件未动） |
|----|------|----------------------|
| B | #5 `GET /roles/:id/{permissions,menus}` | extends/rbac 仍只有 PUT |
| B | #9 api-contract 漏 `DELETE /configs/:key` | api-contract 未动 |
| C | #10 审计迁移编号 000009 vs 000010 | extends/audit 仍 000009；migration 仍 000010 |
| C | #11 种子权限仍 14（acceptance 已 15） | migration seed 无 audit:list/15，down.sql 仍清 1-14 |
| C | #12 种子配置仍 4（config 文档 5） | seed 无 session_timeout |
| C | #14 种子角色 ID 仍 (1,3,4) | seed 未改 |
| C | #15 M6/PROGRESS 无审计页面任务 | 均未动 |
| D | #20 transaction 仍依赖 extends/rbac | 与 acceptance 的 import 禁令冲突未解 |
| D | #21 `internal/extends` 残留 | guardrails(4)/perf-bench(2)/swagger-checklist(2) 未动 |
| D | #22 `NewMySQL/NewRedis` 命名漂移 | mysql/redis 未动 |
| E | #24 cache 文档无 `user:sv` key | cache 未动 |
| F | #25 降级策略矛盾 + L241 重复句 | acceptance 未动（L241 仍重复、L250 仍写回源放行） |
| G | #27 环境变量双轨未说明 | overview 未动；request 用 `VITE_API_BASE_URL`、local-dev 用 `VITE_API_TARGET` |
| G | #29 config remark 前端仍"从另一个接口补充" | frontend/config 未动 |
| G | #30 overview vs PROGRESS 文件清单 | 均未动 |
| H | #33 deploy CI 仍 3-job 精简版 | deploy CI 未动，与 guardrails 8 层不一致 |
| H | #34 compose `/docker-entrypoint-initdb.d` 与 bootstrap"生产手动"并存 | 均未动 |

---

## 四、本轮新发现的问题

1. **前端编译级断裂（最严重）**：`frontend/auth/design.md` L116 仍调用 `userApi.getPermissions()`，但 `request/design.md` 定义的函数是 `getMyPermissions()` —— 后端端点补上了，前端调用点没同步，按此实现 `tsc` 直接报错。
2. **新端点未入契约**：`GET /users/me/permissions` 已加入 extends/user 与 request，但 `api-contract` 未列出该端点。
3. **PROGRESS 与 M1 文档脱节**：M1 文档已写"8 个中间件"，PROGRESS M1 文件清单仍只有 6 个（缺 `gzip.go`、`security_headers.go`）。

---

## 五、评分明细

| 维度 | 权重 | 得分 | 依据 |
|------|------|------|------|
| 架构与模块设计 | 25 | 23 | 无新架构问题；transaction 冲突残留 |
| 可执行性 | 20 | 18 | 后端端点补齐 + role_id 对齐；但 auth store 函数名断裂会卡编译 |
| 验收与测试设计 | 20 | 17 | 503 已补 10009 映射；降级矛盾与 L241 重复句未清 |
| 完整度 | 20 | 17 | 种子/迁移/审计页面/nginx/CI 仍未动 |
| 一致性与细节 | 15 | 12 | 修复质量回升；新端点未入契约、PROGRESS 脱节 |
| **合计** | **100** | **87** | |

---

## 六、下轮建议（评审只记录，不代改）

1. **先修 1 个编译级断裂**：auth store 的 `getPermissions()` → `getMyPermissions()`，并把 `GET /users/me/permissions` 补进 api-contract。
2. **攻 C 组种子定版**：权限 15、session_timeout、角色 ID 连续、down.sql 同步、审计迁移编号取一（000009 或 000010）。
3. **H 组二选一定版**：CI（deploy 精简版 vs guardrails 8 层）、迁移策略（compose 自动 vs 手动）各取一个权威来源。
4. **前端 any 收尾**：auth/config/menu/rbac/state-ui 仍 12 处。
5. PROGRESS 文件清单与 M1 文档对齐（8 中间件），并补审计模块任务。

---

## 七、评审记录

- 核查范围：round3 之后 mtime 变化的 7 个文件 + 未变化但被引用的关联文件（auth/rbac/api-contract/acceptance/cache 等）。
- 判定标准：以文档当前内容为准，"已修"必须能在改动文件中找到对应行。
