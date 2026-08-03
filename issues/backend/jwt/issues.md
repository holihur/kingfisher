# JWT — 设计与实现差异

> 来源：`design/backend/jwt/design.md` 对照 `core/jwt/jwt.go`
> 排查日期：2026-07-31

## P1

### JWT-1 黑名单写入忽略错误
  **Status: ✅ errcheck added**
- 设计：`RevokeToken` 必须把 JTI 写入 `blacklist:token:{jti}`，失败应视为注销失败
- 实现：`core/jwt/jwt.go` 中 `cache.Set(...)` 错误被忽略（`_ =`）；Redis 不可用时黑名单静默失效
- 影响：Redis 故障期间注销/改密踢出失效，旧 token 仍可访问

### JWT-2 session_version 校验未接入 Auth 中间件
  **Status: ⚠️ v2 improvement**
- 设计：`Claims.SessionVersion` 用于强制踢出，Auth 中间件应比对用户当前 `session_version`
- 实现：Claims 与签发时已带 `sv`，但 `extends/rbac/transport/middleware.go` 的 `AuthMiddleware` 只解析 token，不查库比对 `session_version`（`RevokeSessions` 递增后旧 token 仍可用）
- 影响：M2 验收「修改密码后旧 token 失效」类场景无法实现

### JWT-3 刷新逻辑未校验 refresh token 的 Type
  **Status: ⚠️ v2**
- 设计：`ParseToken` 需检查 `claims.Type == "access"`，拒绝 refresh token 当 access 用
- 实现：`ParseToken` 有 type 校验；但 `RefreshToken` 用原 JTI 重签时未校验旧 refresh 的注销状态处理是否完备（依赖黑名单实现，见 JWT-1）
- 影响：与黑名单错误忽略叠加，存在 token 复用风险

## 一致项 ✅
- `GenerateToken(ctx, userID, role, sessionVersion)` 签名与设计一致（sv 已接入）
- Claims 字段（user_id/role/jti/type/sv）与设计一致
- 短 TTL + 黑名单模式设计落地（黑名单 key 规范一致）
