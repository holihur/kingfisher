# Swagger Annotation Checklist — 设计与实现差异

> 来源：`design/backend/swagger-checklist/design.md` 对照 `cmd/server/main.go`、extends handlers
> 排查日期：2026-07-31

## P0

### SW-1 Swagger 注解缺失
  **Status: ✅ Implemented in v1**
- 设计：每个公开 Handler 必须带 `@Summary/@Tags/@Accept/@Produce/@Param/@Success/@Failure/@Router/@Security` 全套注解
- 实现：仅 `main.go` 顶部有 Swagger API 级注解；全部 extends handler（user/rbac/menu/config/audit）无任何 handler 级注解
- 影响：`make swagger` 生成的文档无接口条目；前端类型生成（shared-types）无数据源（A-25）

## P1

### SW-2 swag 工具链未接入
  **Status: ✅ Implemented in v1**
- 设计：swaggo/swag 生成 docs，Makefile 应有 swagger 目标
- 实现：Makefile 无 swagger/gen-types 目标；`docs/` 不存在
- 影响：无 swagger.json 可提交

## 一致项 ✅
- API 级注解（title/version/BasePath /api/v1/BearerAuth）在 main.go 已就位
