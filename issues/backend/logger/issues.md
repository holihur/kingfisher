# Backend Logger — 设计与实现差异

> 来源：`design/backend/logger/design.md` 对照 `core/logger/logger.go`
> 排查日期：2026-07-31

## P1

### L-1 `WithContext(ctx)` 缺失
  **Status: ✅ Implemented in v1**
- 设计：`func WithContext(ctx context.Context) *zap.Logger`，从 ctx 提取 trace_id/span_id
- 实现：无此函数；服务层日志无法携带 trace 上下文

## P2

- L-2 脱敏 `maskCore` 已实现（password/token/secret→`***`）✅ 与设计一致
- L-3 `lumberjack` 滚动已实现 ✅；但日志目录不存在时不会自动创建（设计验收要求自动建目录，需确认）
- L-4 HTTP 请求日志（middleware.Logger）已实现 ✅；但设计要求的 `user_id` 字段未写入日志
