# Kingfisher 项目评审 — Round 8（全量重新审查）

> 评审日期：2026-07-31
> 评审对象：当前全量 56 份设计文档 + PROGRESS + git 基线（4 commits）+ scripts/check-design.sh + CHANGELOG
> 方法：全量通读 + 交叉核对；重点审查本轮新增的"多数据库驱动"特性与 7 个新子目录

---

## 一、结论速览

| 评分对象 | 分数 | 一句话结论 |
|----------|------|------------|
| 设计文档 | **90 / 100** | P0/P1 大量闭环，但新特性"多数据库驱动"自身未闭环，引入两处 P0 级不一致 |
| 项目整体（参考） | 65 / 100 | 仍为纯设计阶段（0 行实现代码） |

> 趋势：85 → 86 → 87 → 88 → 89(round6，快照过期) → 91(round7，乐观估计) → **90（round8 全量复核修正）**。
> round7 的 91 未验证新增内容；全量复核发现多数据库特性与 lint 脚本存在问题，修正为 90。

---

## 二、当前变更盘点（已确认）

### git 基线
- `.git` 已建立，**4 个 commit**（round7-fixes 声称"8 commits"不实）：
  `56654ab init → 97e9fba round5/6 修复 → 63f450a URL 状态同步 → f96a84b 多数据库驱动 + round7 P0/P1`
- `git status` 干净，56 份文档已全部入库。

### 新增文件
| 文件 | 内容 |
|------|------|
| `design/scripts/check-design.sh` | 设计一致性 lint（6 项检查，见 N3 问题） |
| `design/CHANGELOG.md` | 设计变更记录（P2-4 ✅） |
| `reviews/round6-fixes.md` / `round7-fixes.md` | 修复追踪表（含证据列，进步；但仍有失实项） |

### 已核实闭环（round7 P0/P1 主体）
- ✅ P0-1：acceptance L241/L250 语义统一（缓存 miss→回源 DB；连接失败→503）
- ✅ P0-2：api-contract 重复端点已清（rg 验证 0 重复）
- ✅ P0-3：down.sql 清理范围补 `session_timeout`（5 个 key）
- ✅ 端点契约：`DELETE /configs/:key`、`GET /users/me/permissions`、`GET /roles/:id/{permissions,menus}` 已入契约
- ✅ 错误码：`10009 → 503` 全链统一；L241 重复句已清
- ✅ 审计入册（2/3 层）：PROGRESS 有模块任务 ✅、acceptance 有 audit-logs 验收 ✅、**M6 页面缺失 ❌（见 N6）**
- ✅ cache `user:sv:{user_id}`、env 双变量说明、审计迁移编号统一、CI 权威指向 guardrails、迁移策略标注（dev compose 自动 / prod 手动）
- ✅ P2：check-design.sh（存在）、CHANGELOG（存在）、git 基线（存在）

---

## 三、新发现的问题（全量复核）

### 🔴 N1 — 多数据库命名/签名未闭环（P0）
`mysql/design.md` 已定义 `NewDatabase(cfg DatabaseConfig, ...)`，但调用链四处仍用旧签名：
- `startup/design.md` L20：`coreDB.NewGorm(cfg.MySQL, logger)`
- `bootstrap/design.md` L134：`database.NewGorm(cfg.MySQL, logger)`
- `core/design.md` L61：`func NewGorm(cfg MySQLConfig, ...)`
- `di/design.md` L41：`coreDB.NewGorm`

→ 按文档实现 Go 直接编译失败（函数不存在 + 参数类型不匹配）。改名只改了定义处，调用处未同步——与 round3 的 wire 签名问题同型复发。

### 🔴 N2 — SQLite 开发路径的迁移/种子机制冲突（P0）
- `migration/design.md` 全部建表 SQL 为 **MySQL 方言**（`ENGINE=InnoDB`、`utf8mb4`、`DATETIME(3)`、反引号 `key`），SQLite 无法执行。
- `mysql/design.md` 说"SQLite 开发用 GORM AutoMigrate"，但 `bootstrap/design.md` 开发流程是 `make migrate-up`（golang-migrate 执行 SQL 文件）→ 两者矛盾。
- 种子数据（000008 的 MySQL INSERT）在 SQLite 下如何加载**未定义**。
- 结论：开发环境默认 SQLite 的宣称，当前文档无法支撑执行。

### 🟡 N3 — check-design.sh 两个缺陷
1. **macOS 不可运行**：用了 GNU 专属 `grep -P` 与 `realpath --relative-to`，实测直接报 `grep: invalid option -- P`。只能在 Linux/CI 跑，需注明或改兼容写法。
2. **第 3 项"错误码映射检查"是空壳**：循环体为 `if ! grep -q ...; then :; fi # skip check`，实际不检查任何东西。声称 7 项/6 项检查，至少 1 项虚设。

### 🟡 N4 — 7 个空子目录
`backend/` 下新增 `fixture/ graceful-degrade/ handler/ module-lifecycle/ router/ service/ version-endpoint/` 全部为空（0 文件，git 不跟踪）。若为规划占位，应加 README 或设计文件；否则删除。

### 🟡 N5 — config remark 数据链路仍未闭环
前端 `config.remark` 已依赖，但后端 `interface` 与 `extends/config` 的 `GetAll` 仍返回 `map[string]string`（无 remark），api-contract 也无 GET /configs 的完整响应示例。round6-fixes 声称已修，实际只修了前端侧。

### 🟡 N6 — 审计页面第 3 层缺失
round7-fixes 声称 P1-1"PROGRESS+acceptance+M6 含 audit"，但 `M6-fullstack-crud/design.md` 全文无 audit/审计，前端也无独立审计页设计（仅 overview 一行 URL 示例）。失实。

### 🟡 N7 — 修复表数字失实（复发）
round7-fixes 写"8 commits"，实际 4；P1-1 写"M6 含 audit"，实际无。与 round2-fixes 的 35/35 同型问题——修复表仍有"报喜不报忧"倾向。

### 🟢 N8 — 细节
- acceptance 无任何 SQLite 相关验收项（新特性的验收缺失）。
- `readwrite-split/design.md` 仍引用 `MySQLConfig`/旧结构（多数据库改造未触及）。
- `fixture` 等空目录若用于文档拆分，与 CHANGELOG"mysql doc → database doc 重写"描述不一致（实际是原地重写，未建 database/ 目录）。

---

## 四、评分明细

| 维度 | 权重 | 得分 | 依据 |
|------|------|------|------|
| 架构与模块设计 | 25 | 23 | 多数据库驱动方向正确；但 NewDatabase 调用链未同步 |
| 可执行性 | 20 | 18 | P0/P1 大量闭环；SQLite dev 路径执行会挂（迁移方言） |
| 验收与测试设计 | 20 | 18 | 503/降级/审计/URL 同步已补；无 SQLite 验收、lint 第 3 项空壳 |
| 完整度 | 20 | 17 | 审计 3 层缺 M6 页；remark 后端链路缺；7 个空目录 |
| 一致性与细节 | 15 | 12 | 修复表失实（8 commits / M6）；NewGorm 残留；lint 空壳 |
| **合计** | **100** | **90** | |

---

## 五、距 100 分差距清单（衔接 round7 路线图）

**P0（2 项）**
1. 统一 `NewDatabase` 命名与 `DatabaseConfig` 签名（startup/bootstrap/core/di 四处）。
2. 定版 SQLite 开发路径：migration 方言策略 + 种子加载机制 + bootstrap 流程，三处一致。

**P1（4 项）**
3. check-design.sh：改 macOS/Linux 双兼容；第 3 项错误码检查补真实逻辑。
4. remark 后端 `GetAll` 返回完整对象 + api-contract 补 GET /configs 响应示例。
5. M6 补审计页面产出文件 + 前端 audit 页面设计。
6. 清理 7 个空目录（加内容或删除）。

**P2（2 项）**
7. 修复表引入"自动核对"（用 check-design.sh 输出比对声明）。
8. 多数据库验收入 acceptance（sqlite 零依赖启动、三驱动切换、方言差异矩阵）。

---

## 六、评审记录

- 全量通读 56 份文档 + PROGRESS + CHANGELOG + check-design.sh + git 历史。
- 交叉核对：NewDatabase 定义 vs 调用链、migration 方言 vs SQLite 宣称、修复表声明 vs 文件内容、api-contract vs port vs 前端调用。
- 本轮再次验证 P2-1"评审防过期"的价值：round6 基于过期快照、round7 乐观估计未验证新增内容——本报告所有结论均基于 16:32 之后的文件快照。
