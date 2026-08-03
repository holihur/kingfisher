# Guardrails — 设计与实现差异

> 来源：`design/backend/guardrails/design.md` 对照 `.golangci.yml`、scripts/check-*.sh
> 排查日期：2026-07-31

## P1

### GR-1 .golangci.yml 缺失
  **Status: ✅ .golangci.yml exists**
- 设计：`.golangci.yml` 启用 govet/errcheck/staticcheck/gosec/depguard/revive/goimports/unconvert/unparam/wastedassign/misspell/nilerr/noctx/errorlint/gocritic，含 depguard 跨层 import 规则
- 实现：仓库根目录无 `.golangci.yml`（需核对；`make lint` 依赖 golangci-lint 但无配置）
- 影响：CI 硬性约束无配置载体；跨层 import（core→extends）无人拦截

## P1

### GR-2 检查脚本不完整
  **Status: ✅ guardrails + revive scripts exist**
- 设计：check-guardrails.sh 覆盖 panic/log.Fatal/fmt.Println/跨层 import/硬编码等
- 实现：`scripts/check-guardrails.sh` 存在且实现前 3 项检查（panic/Fatal/Println）；`scripts/check-revive.sh` 存在；但 `check-traces.sh`、`check-metrics.sh`、`check-no-panic.sh`（设计 scripts 清单）缺失
- 影响：guardrails 覆盖面不足（见 SCR-1）

## P2

### GR-3 无 CI 执行 guardrails
  **Status: ✅ GitHub Actions CI created**
- 设计：每次 push 强制执行
- 实现：无 CI（DEP-3），脚本仅可手动运行
- 影响：约束不可强制执行

## 一致项 ✅
- 已实现的 check-guardrails.sh 检查项（panic/log.Fatal/fmt.Println）与设计前几项一致
