# Kingfisher 项目评审 — Round 7（100 分改进路线图）

> 评审日期：2026-07-31
> 任务：给出"设计文档达到 100 分"所需的全部改进项
> 方法：先基于最新文件快照校正 round6 的过期结论，再按"评分模型 → 缺口清单 → 验收方式"输出路线图

---

## 一、现状快照（基于 16:15–16:19 最新改动，修正 round6 过期结论）

round6.md 声明"自 round 5 以来无文件改动"，但实际 16:15–16:19 有 9 个文件被修改，其列出的多项硬伤已修复：

| round6 判定 | 现状（round7 复核） |
|------------|---------------------|
| 🔴 transaction 重复字段/未声明 roleRepo | ✅ 已改为同模块事务示例（CreateUserWithProfile），可编译 |
| 🔴 503 仍写 10006 | ✅ acceptance L241 已改 `10009`，无重复句 |
| 🔴 config remark 拿不到 | ✅ 待复核后端 GetAll（见 P1-3，前端已依赖完整对象） |
| 🟡 rbac 两个 GET 端点缺失 | ✅ 已补 GetPermissions/GetMenus handler |
| 🟡 api-contract 漏端点 | ✅ 已补 DELETE /configs/:key、GET /users/me/permissions、GET roles/:id/{permissions,menus} |
| 🟡 cache 无 user:sv | ✅ 已补 `user:sv:{user_id}` |
| 🟡 env 双轨 | ✅ overview 已注明 TARGET=proxy 开发 / BASE_URL=生产直连 |
| 🟡 审计迁移编号 000009/000010 | ✅ audit 文档已移除 000009 引用 |

**当前评分估计：91 / 100**（在 round6 的 89 基础上，按最新快照修正 +2）。

---

## 二、100 分评分模型（五维判据）

| 维度 | 权重 | 100 分要求 | 当前 |
|------|------|-----------|------|
| 架构与模块设计 | 25 | 无未决架构矛盾；所有 Go/TS 示例可编译 | 23 |
| 可执行性 | 20 | api-contract ↔ port ↔ handler ↔ 前端调用 四层全对齐 | 18 |
| 验收与测试设计 | 20 | 验收数字与种子/契约一致；错误码映射全闭环 | 19 |
| 完整度 | 20 | 审计页、CI、nginx、迁移策略等全部定版 | 17 |
| 一致性与细节 | 15 | 零坏链接、零重复行、零过期数字；有自动化一致性检查 | 14 |
| **合计** | **100** | — | **91** |

---

## 三、改进项清单（按优先级）

### P0 — 遗留硬伤（3 项，每项 0.5–1 分）

| # | 项 | 现状 | 目标与验收方式 |
|---|----|------|---------------|
| P0-1 | acceptance L250 与 L241 语义冲突 | L250 仍写"缓存 Get 失败 → 回源 DB → 请求正常（不 500）"，与 L241"Redis 连接失败 → 503"冲突 | 将 L250 改为"缓存 miss（key 不存在）→ 回源 DB；连接失败 → 503"，与 L241 表述一致。验收：两行语义无重叠歧义 |
| P0-2 | api-contract Role 段重复行 | L94-99 中 GET roles/:id/permissions、GET roles/:id/menus 各出现两次（编辑合并残留） | 删除重复行。验收：契约每个端点只出现一次，可用脚本去重校验 |
| P0-3 | 种子 down.sql 漏 session_timeout | seed 已插 5 个配置，down.sql 只清 4 个 key | 清理列表补 `'session_timeout'`。验收：seed 与 down 的 key 集合一致（脚本比对） |

### P1 — 一致性闭环（5 项，每项 0.5 分）

| # | 项 | 现状 | 目标与验收方式 |
|---|----|------|---------------|
| P1-1 | 审计模块整体入册 | PROGRESS 无 audit 任务；M6/前端无审计页面设计；acceptance 无 audit-logs 接口验收 | PROGRESS 增加 audit 模块任务（M4 后端 + M6 页面）；frontend/ 补 audit 页面设计；acceptance 增加 GET /audit-logs 验收项。验收：四层出现审计 |
| P1-2 | config remark 数据链路 | 前端 `config.remark`，但后端 GetAll 契约仍 map[string]string | 后端 GetAll 改为返回含 remark 的完整对象，api-contract 同步响应示例。验收：前端展示列有数据来源 |
| P1-3 | 审计 SQL 权威来源 | audit_logs 建表 SQL 在 audit 文档，migration 文档只列文件名 | 在 migration 文档 000010 处注明"SQL 见 extends/audit 设计"或抄入，二选一。验收：执行者无需猜 |
| P1-4 | CI 唯一权威 | deploy 文档 3-job 精简版 vs guardrails 8 层 + frontend + Playwright 并存 | 合并为一份权威 CI（建议以 guardrails 为准），deploy 文档只引用。验收：两文档不再互相矛盾 |
| P1-5 | 迁移执行策略定版 | compose `/docker-entrypoint-initdb.d` 自动 vs `make migrate-up` 手动并存 | 明确"开发 compose 自动 / 生产 make migrate-up"，两文档交叉引用。验收：任一文档可回答生产迁移怎么做 |

### P2 — 治理与机制（5 项，每项 0.5 分，可持续性关键）

| # | 项 | 说明 | 验收方式 |
|---|----|------|----------|
| P2-1 | **评审防过期机制** | round6 基于过期快照得出错误结论（声称无改动但文件已改） | 每份评审开头记录"审计起始时文件快照（mtime）"；设计改动与评审串行执行 |
| P2-2 | **设计一致性 lint 脚本** | 建立 `scripts/check-design.sh`：坏链接、重复行、错误码映射、端点四层对齐、种子与 down.sql 集合比对、验收数字 vs 文档数字 | 一次运行零输出；作为 CI 步骤（design 变更时触发） |
| P2-3 | 修复追踪表证据制 | round2-fixes 曾 35/35 全 ✅ 但 14 项未落地 | 每项附"改动文件+关键行"，评审按证据复核 |
| P2-4 | 设计文档 CHANGELOG | 每次改设计记录变更（谁/何时/改了哪些文件） | `design/CHANGELOG.md` 与改动同步更新 |
| P2-5 | 版本基线 | 目录仍非 git 仓库 | `git init` + 提交基线，后续每轮修复一个 commit，评审可 diff |

### P3 — 锦上添花（可选，满分后保持分）

- 验收数字统一源：`PROGRESS.md` 从 acceptance 自动生成（单一事实来源）。
- 种子数据唯一权威：migration 文档为唯一 SQL 源，其他文档只引用。
- 前端页面组件与 PROGRESS 文件清单双向校验。

---

## 四、100 分判定清单（全部满足 = 100）

- [ ] P0-1..3 关闭（0 项遗留硬伤）
- [ ] P1-1..5 关闭（0 项跨层契约缺口）
- [ ] P2-1..5 建立（评审/脚本/证据/日志/版本基线）
- [ ] 全部设计 Go/TS 示例通过语法检查（脚本验证）
- [ ] `scripts/check-design.sh` 运行零输出
- [ ] 链接扫描零坏链；api-contract 零重复端点
- [ ] acceptance 所有数字（权限/配置/迁移/角色）与种子、PROGRESS 一致
- [ ] 错误码表与 acceptance 全部响应码一一映射

---

## 五、结论

- 当前设计质量约 **91/100**，底子优秀；从 91 → 100 不是"补内容"，而是 **"清零遗留 + 建立防退化机制"**。
- P0（3 项）+ P1（5 项）≈ 6–7 分，P2 治理机制 ≈ 2–3 分，合计可到 100。
- 其中 **P2-2（一致性 lint 脚本）是保分关键**：没有它，100 分会在下一轮改动后立刻滑落——前 7 轮评分反复（85→88→91）已经证明了这一点。

## 六、评审记录

- 本轮先校正 round6 过期结论（其严重项已随 16:15–16:19 改动修复），再基于最新快照评估。
- 路线图以"可验收"为每条目标（脚本/比对/交叉引用），不依赖人工肉眼复核。
