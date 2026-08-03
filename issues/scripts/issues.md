# Scripts — 设计与实现差异

> 来源：`design/scripts/design.md` 对照 `scripts/`
> 排查日期：2026-08-03

## P1

### SCR-1 自动化脚本大部分缺失
- 设计：8 个脚本（chaos.sh / deploy-check.sh / bench.sh / check-traces.sh / check-metrics.sh / check-no-panic.sh / check-frontend-constraints.sh + smoke.sh）
- 实现：`scripts/` 仅有 `smoke.sh`、`check-guardrails.sh`、`check-revive.sh`
- 影响：混沌测试、部署检查、压测、trace/metrics 校验、panic 扫描、前端约束扫描全部无载体；acceptance 承诺的「自动化验收 100%」无法成立

### SCR-2 check-design.sh 未实现
- 设计：`design/scripts/check-design.sh` 为设计侧验收脚本（README 配套）
- 实现：根目录 `scripts/` 无 check-design.sh
- 影响：设计文档与实现一致性无法自动校验

## P2

### SCR-3 smoke.sh 覆盖有限
- 设计：smoke 覆盖关键链路（登录→列表→CRUD→RBAC 403）
- 实现：`smoke.sh` 覆盖 health/version/login/register/users 基础项；RBAC 403 场景依赖 Redis 可用才执行
- 影响：CI 无 Redis 时鉴权链路不验证

## 一致项 ✅
- `smoke.sh` 存在且实现基础检查（curl + HTTP 状态断言）✅
- `check-guardrails.sh` 覆盖 panic/log.Fatal/fmt.Println ✅（见 GR-2）
