# Audit Gap Report — 缺失项追踪报告映射

> 来源：`design/backend/audit/gap-report.md`（23 项 P0/P1/P2 缺失追踪）对照 `issues/acceptance/issues.md`
> 排查日期：2026-07-31

## 说明

`design/backend/audit/gap-report.md` 是设计侧的缺失项追踪报告（P0×7、P1×7、P2×9）。本文件不重复造问题，仅记录映射关系，避免重复劳动。

## 映射

| gap-report 缺失项 | 已落 issue | 状态 |
|---|---|---|
| PUT /users/me/password 修改密码 | 已实现（`ChangePassword` handler 存在） | ✅ 已闭合（与 gap-report 描述相反——设计文档滞后于实现） |
| 操作审计日志零覆盖 | EA-1/EA-2/EA-3/EA-5 | 未闭合（Audit 中间件/Flush/LOGIN 仍缺） |
| 其他 P0/P1 项（RBAC、菜单过滤、权限校验、限流、迁移等） | A-2/A-3/A-4/A-20/A-21/A-24/A-25 等 | 未闭合，见 acceptance 与各模块 issue |

## 结论

- gap-report 声称的「PUT /users/me/password 缺失」已被实现，属**设计文档滞后**（实现有 `ChangePassword`，且 `api-contract` 也列出了该接口）✅
- 审计模块的设计（`design/backend/extends/audit/design.md`）与实现差异详见 `issues/backend/extends/audit/issues.md`
