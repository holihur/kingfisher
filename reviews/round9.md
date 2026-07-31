# Kingfisher 项目评审 — Round 9（全量重新评估）

> 评审日期：2026-07-31
> 评审对象：当前全量 58 份设计文档 + PROGRESS + git 基线（9 commits）+ scripts/check-design.sh
> 评审基线：**HEAD = `3bcb52e`**（`d6fc611` → `2350d63` → `3bcb52e` 为 round8 之后新增）
> 方法：全量通读 + 交叉核对；评审期间仓库被并发修改（发现 6→9 commits 演进），本报告结论全部基于 `3bcb52e` 快照

---

## 一、结论速览

| 评分对象 | 分数 | 一句话结论 |
|----------|------|------------|
| 设计文档 | **91 / 100** | 两大 P0（N1 调用链、N2 SQLite 迁移）全部闭环；剩余 6 项 P1/P2 未动，新增 3 处小不一致 |
| 项目整体（参考） | 65 / 100 | 仍为纯设计阶段（0 行实现代码） |

> 趋势：85 → 86 → 87 → 88 → 89(过期) → 91(乐观) → 90(round8 全量复核) → **91（round9，P0 真闭环）**。
> 与 round8 相比：分数 +1，但构成变化大——两个 P0 已从"存在"变为"闭环"，扣分点转移到 6 项遗留 P1/P2 与新引入的一致性细节。
> 注：round8 评分表小计（23+18+18+17+12=88）与声称总分 90 不符（算术笔误）；round9 起小计与总分严格一致。

---

## 二、本轮已确认闭环（round8 P0 全清）

### ✅ N1 — 多数据库调用链已同步（P0 闭环）
- `core/design.md:61-62` 定义双层入口：`NewDatabase`（底层连接）+ `InitDatabase`（上层入口：连接 + SQLite AutoMigrate + Seed），签名统一 `(cfg DatabaseConfig, logger *zap.Logger) (*gorm.DB, error)`
- 调用点全部对齐 `InitDatabase(cfg.Database, logger)`：`startup/design.md:20`、`bootstrap/design.md:133`、`overview/design.md:122`；`di/design.md:41` 用 `coreDB.NewDatabase`
- `NewGorm` 旧名在核心链路 0 残留（仅 `readwrite-split` 仍遗留，见 N8）

### ✅ N2 — SQLite 开发路径已定版（P0 闭环）
- `migration/design.md` 新增"SQLite 开发模式——自动迁移"章节：SQLite 走 GORM AutoMigrate + Go 种子，MySQL/PG 走 golang-migrate SQL，附三环境策略对比表
- 新增 `extends/seed/design.md`（117 行）：定义 `Seed(db *gorm.DB) error`，依赖顺序写入、单事务回滚、注明 Go 种子与 000008 SQL 双源需同步维护
- `bootstrap/design.md`：SQLite 模式 `make run` 自动建表+种子，MySQL/PG 运维手动 `make migrate-up`；`mysql/design.md`、`overview/design.md` 同步
- `.gitignore` 已加 SQLite dev 文件；`dependencies.md`（189 行，41 包清单 + go.mod/package.json 最小示例 + 无 Docker 方案）补齐依赖视图

---

## 三、未闭环问题（round8 遗留，本轮复核原样存在）

| 编号 | 级别 | 状态 | 证据 |
|------|------|------|------|
| N3 | P1 | ❌ 未修 | `check-design.sh` mtime 仍 16:30；第 45 行 `grep -oP`（GNU 专属，macOS 报错）；第 48–50 行错误码检查是空壳 `if ...; then :; fi  # skip check` |
| N4 | P1 | ❌ 未修 | `backend/` 下 7 个空目录仍 0 文件：fixture / graceful-degrade / handler / module-lifecycle / router / service / version-endpoint |
| N5 | P1 | ❌ 未修（且加深） | remark 链路三处签名互斥：`service-interface/design.md:71` port 返回 `map[string]string`；`extends/config/design.md:50` port 返回 `[]domain.SystemConfig`；`:65` 实现仍 `map[string]string` 且 cache 策略转 map（丢 remark）；api-contract `GET /configs` 无响应示例 |
| N6 | P1 | ❌ 未修 | `M6-fullstack-crud/design.md` 全文 audit/审计命中 0；前端仅 `overview/design.md:115` 一个 URL 示例，无独立审计页设计 |
| N7 | P1 | ❌ 未修 | `round7-fixes.md:14` 声称"8 commits"，实际 9（round8 时 4）；commit `d6fc611` 消息自称"round8 — 3 remaining issues closed for 100/100"，但 N3/N5/N6 等仍在 → 过度声明复发 |
| N8 | P2→P1 | ⚠️ 部分 | `readwrite-split/design.md:17` 仍为 `NewGormWithRW(cfg MySQLConfig, ...)`（多数据库改造未触及，N1 同型残留）；acceptance 全文 SQLite 命中 0 |

---

## 四、新增观察

### 🟡 O1 — 用户-角色为单角色模型（业务决策需确认）
- `users` 表仅 `role_id` 单列（`migration/design.md:45`、`domain/design.md:19`），无 `user_roles` 关联表
- JWT claim 为单值 `role string`（`jwt/design.md:13,25`）；前端用户表单角色单选（`frontend/user/design.md:149`）
- 若业务要求"一用户多角色"：需新增关联表 + JWT claims 改数组 + 权限聚合 + 前端多选，属设计缺口，建议尽早确认

### 🟡 O2 — bootstrap 2.3 段残留 MySQL 验证
- `bootstrap/design.md:73-83` 在 SQLite-first 流程后仍展示 `mysql -u root -p kingfisher -e "SHOW TABLES;"`，SQLite 模式下无法执行，与 2.2 节新流程矛盾（P2）

### 🟢 O3 — 细节
- commit `f909af1` 消息拼写错误："signaure"（应为 signature）
- `bootstrap/design.md:133` 图内注释列对齐错位（新签名比旧签名长，边框未对齐，纯排版）

---

## 五、评分明细

| 维度 | 权重 | 得分 | 依据 |
|------|------|------|------|
| 架构与模块设计 | 25 | 24 | 双层数据库入口设计合理、调用链闭环；readwrite-split 残留旧签名 |
| 可执行性 | 20 | 19 | SQLite 开发路径可按文档跑通；bootstrap 2.3 残留旧验证命令 |
| 验收与测试设计 | 20 | 18 | lint 脚本仍不可用且空壳；无 SQLite 验收项 |
| 完整度 | 20 | 18 | 补 seed 文档 + 依赖清单；空目录、审计页、remark 链路仍缺 |
| 一致性与细节 | 15 | 12 | 两大 P0 同步修复；N5 三处签名互斥、修复表/commit 声明失实 |
| **合计** | **100** | **91** | 小计与总分一致 |

---

## 六、距 100 分差距清单（更新）

**P0（0 项）** —— 已清零 ✅

**P1（6 项）**
1. `readwrite-split` 对齐多数据库：`NewGormWithRW(MySQLConfig)` → 新签名或标注停用
2. remark 链路三处统一：port/impl 返回 `[]domain.SystemConfig`，cache 保留 remark，api-contract 补 `GET /configs` 响应示例
3. check-design.sh：去 `grep -oP`/`realpath --relative-to`（macOS 兼容），第 3 项补真实校验逻辑
4. M6 补审计页面产出 + 前端独立审计页设计
5. 清理 7 个空目录（加 README/设计或删除）
6. 修复表与 commit 消息引入自动核对（以文件内容/`git rev-list` 为准）

**P2（3 项）**
7. 多数据库验收入 acceptance（sqlite 零依赖启动、三驱动切换、方言差异矩阵）
8. bootstrap 2.3 删除/改为 SQLite 版验证命令
9. 确认单角色模型是否满足业务（否则升级多角色方案，见 O1）

---

## 七、评审记录

- 基线锁定：`3bcb52e`（评审期间仓库从 6 commits 演进到 9，全程只认锁定基线）
- 交叉核对：函数签名（core/mysql/migration/startup/bootstrap/overview/di）、迁移策略三文档一致性、seed 双源、remark 三处 port/impl、修复表声明 vs `git rev-list`
- 教训固化：并发修改场景下，评审必须先记 HEAD 再动手；本报告所有结论可凭 `3bcb52e` 复现
