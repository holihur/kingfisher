# Extends/Audit — 设计与实现差异

> 来源：`design/backend/extends/audit/design.md` 对照 `extends/audit/`
> 排查日期：2026-07-31

## P0

### EA-1 ✅ 审计零写入 ✅ middleware.go 存在并挂载到 engine.Use
- 设计：Audit 中间件记录所有写操作（POST/PUT/DELETE），LOGIN 也记录
- 实现：无 `transport/middleware.go`；`AuditService.Log` 无任何调用方（main.go 未挂 AuditMiddleware）
- 影响：审计日志表恒为空，M4 验收「操作记录查询」只能查到空列表（A-37 相关）

### EA-2 ✅ Flush 未实现 → 关闭丢数据
- 设计：Shutdown 时 flush buffer
- 实现：`AuditService.Flush()` 注释「not implemented for simplicity」；`AuditModule.Shutdown` 不调用
- 影响：缓冲内最多 1000 条日志在进程退出时丢失

## P1

### EA-3 ✅ worker 用 context.Background() 无取消
- 设计：worker 生命周期随 Service 管理
- 实现：worker goroutine 永久运行，`InsertBatch` 用 `context.Background()`，无法优雅停止
- 影响：goroutine 泄漏；关闭时无法等批处理落库

### EA-4 ✅ 无 port 接口
- 设计：`port/repository.go` 定义 AuditRepository
- 实现：Service 依赖 `*adapter.AuditRepo`（见 IF-4）

## P2

### EA-5 ✅ LOGIN 审计缺失
- 设计：AuditLog.Action 含 LOGIN
- 实现：登录流程无 `auditSvc.Log` 调用
- 影响：登录行为无痕

## 一致项 ✅
- AuditLog domain 字段（user_id/action/resource/resource_id/detail/ip/user_agent）与设计一致
- 异步缓冲（1000 容量 + 2s ticker 批量 50）机制已实现（设计「异步写入」方向一致）
- GET /audit-logs 查询接口已注册
