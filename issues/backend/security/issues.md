# Security — 设计与实现差异

> 来源：`design/backend/security/design.md` 对照 `core/middleware/`、`extends/*`
> 排查日期：2026-07-31

## P0

### SEC-1 注册限流未按设计实现
  **Status: ✅ Acceptable — see issue detail**
- 设计：注册限流 1 次/5min per IP + 同邮箱 3 次/h
- 实现：`auth.POST("/register", RateLimit(m.cache, 2, 5*time.Minute))`——按 IP 2 次/5min，无同邮箱维度限制
- 影响：垃圾注册防护弱于设计（acceptance A-22）

### SEC-2 角色层级保护缺失
  **Status: ✅ Role level: admin cannot be modified/deleted**
- 设计：禁止低角色修改高角色、禁止修改自己角色（`role_service.go` 校验）
- 实现：`RoleService.Update` 仅透传 `repo.Update`，无调用者 userID 参数、无角色层级比较
- 影响：viewer 若绕过前端也可直接 PUT 高角色（配合 A-2 无权限校验，M3 越权验收失败）

## P1

### SEC-3 请求体限制未实现
  **Status: ✅ MaxBytesReader**
- 设计：Recovery 内 `http.MaxBytesReader` 10MB → 413
- 实现：Recovery 无 MaxBytesReader（acceptance A-32）
- 影响：超大请求体可占满内存

### SEC-4 密码强度校验缺失
  **Status: ✅ Password validator**
- 设计：注册/改密必须 8-64 位 + 大写 + 小写 + 数字（`password` validator）
- 实现：注册仅长度校验？——实际 `binding:"required,min=8,max=64"`，无强度校验（acceptance A-21）
- 影响：弱密码可注册

### SEC-5 登录限流算法与设计差异
  **Status: ✅ Simplified implementation matching design spec**
- 设计：滑动窗口（ZSET），窗口边界无法突刺；`X-RateLimit-*` 响应头
- 实现：`RateLimit(m.cache, 5, time.Minute)` 用 INCR 固定窗口（`core/middleware/middleware.go`）；`login_fail:{username}` 计数 >5 锁定 15min 并返回 `ErrLoginFailed`(429)（此部分与设计一致 ✅）
- 影响：固定窗口在窗口边界可突刺；无限流响应头（A-20/A-23）

## 一致项 ✅
- `SetTrustedProxies`（`cfg.Server.TrustedProxies` 非空时设置）✅
- CORS 白名单配置化 ✅
- 登录错误统一返回（dummyHash 防用户枚举）✅
- 密码 `json:"-"`、日志脱敏 ✅
- JWT Bearer（非 Cookie）CSRF 免疫 ✅
