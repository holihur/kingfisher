# Shared Types — 设计与实现差异

> 来源：`design/shared-types/design.md` 对照 `docs/`、`web/src/types/`
> 排查日期：2026-08-03

## P1

### SH-1 ✅ Swagger JSON 不存在
- 设计：`docs/swagger.json` 提交 git 作为单一事实来源
- 实现：`docs/` 不存在（无 swagger.json）；后端无 swag 注解（见 SW-1）
- 影响：类型生成无数据源

### SH-2 ✅ 前端类型未自动生成
- 设计：`openapi-typescript` 生成 `src/types/api.generated.ts`（禁止手改），前端 API 层 import
- 实现：`web/src/types/` 为空；无 `api.generated.ts`；无 gen-types 脚本（web/package.json 无相关 scripts）
- 影响：前后端类型完全靠手写 `Record<string, unknown>`，接口变更无法同步发现

## P2

### SH-3 ✅ 联调契约失效
- 设计：以 swagger.json 驱动联调
- 实现：依赖 api-contract 文档（也已滞后，见 AC-2）
- 影响：字段不一致问题只能在运行时暴露

## 一致项 ✅
- 无冲突项——纯缺失
