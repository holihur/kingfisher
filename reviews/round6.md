# Kingfisher 项目评审 — Round 6（全面复核 + 遗留盘点）

> 评审日期：2026-07-31
> 评审对象：全部设计文档（56 份 .md）—— 自 round 5 以来无文件改动，本轮做全量深度复核
> 方法：逐项重验 round 5 遗留 14 项 + 新鲜扫描跨文档一致性

---

## 一、结论速览

| 评分对象 | 分数 | 一句话结论 |
|----------|------|------------|
| 设计文档 | **89 / 100** | 架构扎实、覆盖度高，但端点/错误码/事务示例三类硬伤反复出现 |
| 项目整体（参考） | 65 / 100 | 仍为纯设计阶段 |

> round1=87 → round2=85 → round3=86 → round4=87 → round5=88 → round6=89。六轮累计 +2，说明设计本身底子好，扣分集中在少量顽疾。

---

## 二、Round 5 遗留 14 项复核结论

### ✅ 已自行解决（3 项）

| # | 对应 | 问题 | 当前状态 |
|---|------|------|----------|
| C-#10 | round3 | 审计迁移编号冲突 | ✅ 已统一为 `000010`（audit/design L194 + migration/design L29-30 一致） |
| D-#22 | round3 | 命名漂移 `NewMySQL/NewRedis` | ✅ 一致：mysql L10 `NewMySQL` + redis L10 `NewRedis`，命名规范统一 |
| F-#25 | round4 | 降级策略矛盾 | ✅ 503 错误码已统一：errcode L40 定义 `10009→503`，acceptance L241/L251 均返回 503，语义已对齐 |

### ❌ 仍未修复（11 项）

| 组 | # | 问题 | 现状证据 |
|----|---|------|----------|
| B | #5 | `GET /roles/:id/{permissions,menus}` 端点缺失 | rbac port L87-97 仅有 `AssignPermissions/AssignMenus`，无 Get 方法；route L167-179 仅注册 PUT；但 api-contract L94-96 列出 GET → 契约写了但没实现 |
| B | #9 | api-contract 漏 `DELETE /configs/:key` | api-contract L102-106 Config 段只列 GET/GET/:key/PUT；但 extends/config port L64-69 有 `Delete` 方法 + frontend/config L178 `configApi.delete` 已定义 |
| B | 新#2 | `GET /users/me/permissions` 未入契约 | api-contract L66-73 User 段无此端点；但 frontend/auth L116 `userApi.getMyPermissions()` 已调用 |
| C | #14 | 种子角色 ID 仍 (1,3,4) | migration L147-149 未改。**注意**：这是有意设计（migration L264 注明"预留扩展"），非遗漏 |
| C | #15 | 审计日志页面缺失 | acceptance 全文无 audit-logs 验收项；PROGRESS M6 任务清单无审计页；frontend/ 目录无 audit 页面设计 |
| E | #24 | cache 文档无 `user:sv` key | cache/design Key 规范 L73-81 + 失效策略 L98-103 均未列 `user:sv:{user_id}`；但 redis/design L48 已用此 key |
| G | #27 | 环境变量双轨 | overview/design 不读 .env（Go 用 config.yaml）；deploy/design L38-39 docker-compose 用 `${MYSQL_PASSWORD}`。两套机制并存但 overview 未交代何时用哪个 |
| G | #30 | overview vs PROGRESS 文件清单 | overview/design 列了 `readwrite-split`、`service-interface`、`validation` 等模块，PROGRESS 未体现 |
| H | #32 | nginx.conf 无文档 | deploy/design 无 nginx 配置（frontend nginx 反代、gzip、SPA fallback 均无） |
| H | #33 | deploy CI 仍 3-job | deploy/design L157-193 仅 lint/test/build 三个串行 job |
| H | #34 | 迁移策略冲突 | migration L229-238 `make migrate-up` 手动执行；deploy/design L43 docker-compose `./migrations:/docker-entrypoint-initdb.d` 首次启动自动执行 → 两种策略并存，未说明生产用哪个 |

---

## 三、本轮新发现 / 深层问题

### 🔴 严重（编译级 / 契约级）

1. **transaction 示例代码仍不可编译**：`backend/transaction/design.md` L57-58 `userRepo port.UserRepository` 声明两次；L69 调用 `s.roleRepo.AssignUser` 但 struct 无 `roleRepo` 字段。Go 编译器直接报错。且跨模块事务（user 依赖 rbac 的 repo）与 acceptance L185 "extends/user 不 import extends/rbac" 冲突未解决。

2. **503 错误码矛盾仍在 acceptance 中残留**：虽然 errcode 已定义 `10009→503`，但 acceptance L241 四处 503 响应体仍写 `{"code":10006}`——而 `HTTPStatus(10006)` 映射的是 500。验收预期与实现不兼容：要么 acceptance 改 `10006→10009`，要么 errcode 改映射。

3. **前端 config 拿不到 remark 数据**：frontend/config L89 `remark: config.remark` — 后端 `GetAll` 返回 `map[string]string`（api-contract L67 + config port L71），map 里没有 remark 字段。前端即使改了也拿不到数据。

### 🟡 中等（一致性问题）

4. **api-contract 与 port 接口不匹配**：api-contract 列出 `GET /roles/:id/permissions` 和 `GET /roles/:id/menus`，但 port 层无对应接口方法 → 前端 rbac 调用了两个不存在的后端端点。

5. **前端 auth 调用了不存在的端点**：frontend/auth L116 `userApi.getMyPermissions()` — api-contract User 段无此端点，后端 port 层无此方法。

6. **中间件顺序不一致**：overview/design L67-72 顺序（Recovery→Trace→Logger→Gzip→SecurityHeaders→CORS）与 startup/design 不一致，且与 PROGRESS L36-38 的八中间件清单对不上（overview 少列 middleware）。

7. **acceptance L241 行文仍有编辑痕迹**：`"拒绝，。 {json}。——"` 标点错乱 + 重复句，虽然 503 错误码语义已对齐，但该行可读性仍是问题。

### 🟢 轻微

8. **migration down.sql 漏 `session_timeout`**：migration/design L250 的清理范围是 4 个 key，但 seed 实际插入 5 个（L219 有 `session_timeout`）。

9. **audit 迁移与 migration 设计重复**：audit/design L194-215 完整写了 audit_logs 建表 SQL，migration/design L29-30 也列了 `000010_create_audit_logs` 但没给 SQL → 两处定义但只有一处有 SQL，执行者不知道以哪个为准。

---

## 四、评分明细

| 维度 | 权重 | 得分 | 依据 |
|------|------|------|------|
| 架构与模块设计 | 25 | 23 | Core+Extends 分层清晰、ADR 5 项决策可追溯、port 依赖反转到位；transaction 示例编译失败 -2 |
| 可执行性 | 20 | 18 | 95% 文档可直接编码；rbac GET 端点缺失 + config remark 不可拿 + auth getMyPermissions 无端点 -2 |
| 验收与测试设计 | 20 | 18 | 311 条验收 4 层自动化覆盖 + Chaos 工程；503 错误码矛盾 + acceptance 编辑痕迹 -2 |
| 完整度 | 20 | 17 | 56 文档覆盖全链路；审计页//nginx/迁移策略冲突 -3 |
| 一致性与细节 | 15 | 13 | api-contract vs port vs 前端调用三端不对齐；中间件顺序漂移 -2 |
| **合计** | **100** | **89** | |

---

## 五、提分建议（评审只记录，不代改）

### 🔴 必须修（编译/契约级）

1. **transaction 示例代码**：删 L57 重复字段；L69 要么加 `roleRepo` 声明要么改为同模块事务示例；与 acceptance "extends/user 不 import rbac" 二选一定版。
2. **acceptance 503 错误码**：L241/L251 四处 `10006` → `10009`（或反向改 errcode 映射，但前者更合理）。
3. **config remark**：后端 `GetAll` 改为返回 `[]Config{Key,Value,Remark}` 或新增 `GetWithRemark`，否则前端 remark 列永远空白。

### 🟡 应当修（一致性）

4. **补齐 rbac 两个 GET 端点**：port 加 `GetPermissions(roleID)` 和 `GetMenus(roleID)`，handler 加对应 GET handler，前端已有调用。
5. **补 `GET /users/me/permissions` 入 api-contract + port + handler**。
6. **migration down.sql** 补 `session_timeout` 清理。
7. **audit SQL 去重**：migration 设计要么完整抄入 audit 的 SQL，要么 audit 文档标注"以 migration/000010 为准"。

### 🟢 可以后续

8. 中间件顺序在 overview/startup/PROGRESS 三处对齐。
9. nginx.conf 补进 deploy 设计。
10. 迁移策略定版（推荐生产用 `make migrate-up`，docker-compose init 仅开发环境）。
11. CI 加 frontend-build job，lint/test/build 并行而非串行。
12. 审计日志页面入 PROGRESS M6 + frontend/ 补设计。

---

## 六、评审记录

- 核查范围：全部 56 份设计文档 + PROGRESS.md + .gitignore
- 判定标准：以文档当前内容为准；"已修"必须能在文件中找到对应行；跨文档一致性以三方对齐为准
- 本轮重点：
  1. 重验 round 5 遗留 14 项（3 已解决 / 11 仍开放）
  2. 端到端追踪 api-contract ↔ port ↔ handler ↔ frontend 四层一致性
  3. 编译级检查所有 Go 示例代码
  4. 错误码映射 cross-check（errcode ↔ acceptance）
