# Frontend Request — 设计与实现差异

> 来源：`design/frontend/request/design.md` 对照 `web/src/api/request.ts`
> 排查日期：2026-07-31

## P1

### FRQ-1 baseURL 未走环境变量
- 设计：`baseURL: import.meta.env.VITE_API_BASE_URL`（如 http://localhost:8080/api/v1）
- 实现：`baseURL: '/api/v1'` 硬编码（依赖 Vite proxy）
- 影响：生产环境若不走同一 proxy 需改代码；.env 分层（development/production）未落地（见 FLD-2）

### FRQ-2 token 来源用 localStorage 而非 Zustand
- 设计：`useAuthStore.getState().token`（单一状态源）
- 实现：直接 `localStorage.getItem('kingfisher_token')`
- 影响：状态与存储双份，登出/刷新不同步时拦截器可能读到脏 token（轻微）

## P2

### FRQ-3 刷新后重放队列实现简化
- 设计：pendingRequests 队列中等待的请求用新 token 重放
- 实现：`handleTokenRefresh` 有队列机制，但刷新失败时 pendingRequests 未统一 reject（需核对）
- 影响：并发请求在刷新失败场景可能悬挂

## 一致项 ✅
- 双拦截器（请求带 Bearer / 响应按 code 分发）✅
- code==0 直接返回 data ✅；10104 触发刷新 ✅；10003/10105 清 token 跳登录 ✅
- 401/403/404/429/500 状态码文案映射 ✅
- 超时 15s、Content-Type json ✅
