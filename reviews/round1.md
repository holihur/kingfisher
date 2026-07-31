# Kingfisher 项目评审 — Round 1

> 评审日期：2026-07-31
> 评审对象：`/Users/jzx/Desktop/kingfisher`（当前处于设计阶段，尚未开始实施）
> 评审范围：52 份设计文档（8656 行）+ `PROGRESS.md`（82 个任务，完成 0）

---

## 一、结论速览

| 评分对象 | 分数 | 一句话结论 |
|----------|------|------------|
| 项目整体 | **65 / 100** | 一流的蓝图，还没动工 |
| 设计文档（本次交付物） | **87 / 100** | 高质量蓝图，扣分在契约未写完与少量自相矛盾 |

> 澄清：用户确认"当前项目就是设计，还没开始实施"，因此以**设计文档**为主要评分对象，
> 项目整体分仅作参考（0 行代码、无 git、无 CI 拉低了整体分）。

---

## 二、项目整体评分（参考）：65 / 100

| 维度 | 权重 | 得分 | 说明 |
|------|------|------|------|
| 设计深度与完整度 | 30 | 27 | 52 份设计文档、8656 行，覆盖 M1–M7 全里程碑 |
| 架构合理性 | 25 | 21 | Core + Extends 插件化、依赖方向清晰、有 ADR 记录取舍 |
| 验收与自动化设计 | 15 | 13 | 约 415 条验收用例、Playwright/chaos/bench 脚本级可执行 |
| 一致性细节 | 10 | 6 | 少量坏链接、符号误用、示例违反自家规则 |
| 实现与可运行性 | 20 | 0 | `0/82` 任务，无任何可运行产物 |
| **合计** | **100** | **65** | |

---

## 三、设计文档评分（主评分）：87 / 100

| 维度 | 权重 | 得分 | 依据 |
|------|------|------|------|
| 架构与模块设计 | 25 | 23 | Core+Extends 插件化、依赖方向清晰、ADR 记录取舍 |
| 可执行性 | 20 | 18 | 每个里程碑有产出文件清单 + 验证命令，82 个任务粒度合理 |
| 验收与测试设计 | 20 | 18 | 约 415 条分层验收、Playwright/chaos/bench 脚本级可执行 |
| 完整度 | 20 | 16 | 全链路覆盖但 API 契约只详述了约 5 个端点 |
| 一致性与细节 | 15 | 12 | 少量坏链接 + 示例代码违反自家 guardrail |
| **合计** | **100** | **87** | |

---

## 四、亮点

1. **蓝图密度极高**
   - 52 份文档、8656 行，从 `core/` 子包划分到 `extends/` 模块文件级规划。
   - `design/M1-core-startup/design.md` 等里程碑文档列出全部产出文件（如 M1 的 19 个文件），做到"执行者零决策"。

2. **验收标准是顶级水平**
   - `design/acceptance/design.md` 按 Happy/Sad/Edge/Chaos 分层计数（约 150/60/41/20 条）。
   - 已写好 `chaos.sh`、`deploy-check.sh`、`bench.sh` 脚本级内容，可直接转 CI。
   - 明确"100% 自动化、不接受人工验收"。

3. **治理意识强**
   - `design/backend/guardrails/design.md`：13 条 CI 硬约束（depguard 防反向依赖、覆盖率 ≥80% 等）。
   - `design/backend/adr/design.md`：5 个 ADR，含替代方案对比与否决理由。

4. **前后端契约有闭环思路**
   - `design/shared-types/design.md`：以 Swagger JSON 为单一事实来源，前端自动生成类型。

---

## 五、扣分点

### 1. API 契约未达自家标准（完整度 -4）
- `design/backend/api-contract/design.md` 自述"为每个 API 接口提供完整的请求/响应示例"。
- 实际只详述了约 5 个端点（login、users、menus/tree、roles/permissions、configs/:key），
  其余预计共 20+ 端点只有模块标题。
- 后果：实现者仍需自行决定错误码、分页格式、字段命名，与"零决策"承诺不符。

### 2. 示例代码违反自家规则（一致性 -2）
- `design/frontend/auth/design.md` 登录页示例使用 `catch (err: any)`。
- 违反 `design/backend/guardrails/design.md` 第 9 条"禁止 `any` 类型（ESLint）"。

### 3. 轻微不一致（一致性 -1）
- 坏链接 2 处：
  - `design/backend/extends/rbac/design.md` → `../middleware/design.md`（目标不存在）
  - `design/M6-fullstack-crud/design.md` → `../../shared-types/design.md`（相对路径错误）
- M4 引入 `wire` 与 M1 手工注册的组装方式并存，未说明取舍。
- `PROGRESS.md` 在 `0/19` 的里程碑上标 ✅，易误读。

### 4. 范围偏重（可执行性风险）
- 脚手架设计包含 Grafana、OTel、读写分离、vegeta P99 基准、415 条验收、80% 覆盖率门槛。
- 作为设计文档很豪华，但执行时是明显的成本/风险项，建议标注 P1/P2 分期。

---

## 六、提分建议（按优先级）

1. 补齐 API 契约：把 user/menu/rbac/config 全部端点按现有模板写完（最实质的完整性缺口）。
2. 修复 `err: any` 示例和 2 处坏链接——设计文档自身要成为 guardrails 的"第一遵守者"。
3. 在 `PROGRESS.md` 或 acceptance 中标注重型项优先级，Grafana/读写分离标为"后续迭代"。
4. 实施阶段：`git init` 建立版本基线 → 启动 M1 的 19 个文件 → 按 acceptance 逐条验证 `curl /health` → 同步搭 `.golangci.yml` 和 GitHub Actions。

---

## 七、评审记录

- 核查方式：目录结构、文件统计、链接有效性扫描、代表性文档精读（overview/acceptance/guardrails/adr/api-contract/migration/frontend-auth 等）。
- 无 TBD/TODO/待定项，文档完成度高。
- 设计阶段分数会随 API 契约补全与瑕疵修复上升到 90+；项目整体分将随 M1 落地快速提升。
