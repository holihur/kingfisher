# Response — 设计与实现差异

> 来源：`design/backend/errcode/design.md`（统一响应格式）对照 `core/response/response.go`
> 排查日期：2026-07-31

## 结论

实现完全覆盖设计，且额外提供了 Gin 便捷封装，无差异。

## 一致项 ✅
- `Response{Code, Message, Data(omitempty)}` 结构与设计一致
- `OK(data)` / `Err(code)` / `ErrWithMsg(code,msg)` 语义与设计一致
- 实现额外提供 `Page`、`JSON`、`AbortJSON`、`OKJSON`、`ErrorJSON`、`BadRequest`、`Forbidden`、`Unauthorized`、`InternalError` 等 helper（设计未要求，属增强，无冲突）

## P2

### RSP-1 PageData 分页结构未在设计文档定义
- 设计：分页响应仅提及「分页」，未定义字段
- 实现：`PageData{items,total,page,page_size}`
- 影响：无功能差异；建议将分页契约补入 api-contract 文档
