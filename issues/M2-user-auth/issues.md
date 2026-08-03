# M2 用户注册登录 — 设计与实现差异

> 来源：`design/M2-user-auth/design.md` 对照实现
> 排查日期：2026-08-03 ｜ 详见各模块 issue

## P0

### M2-1 登录 token 角色硬编码 viewer（EU-1）
- 期望：admin 登录 → token 含 admin 角色
- 现状：`GenerateToken(..., "viewer", ...)` 硬编码
- 影响：一切按角色判断的后端/前端逻辑失真

### M2-2 session_version 踢出未生效（JWT-2）
- 期望：改密/踢出后旧 token 立即失效
- 现状：sv 递增了，但 Auth 中间件不比对

### M2-3 删除用户接口必 500（EU-3）
- 期望：DELETE /users/:id 成功
- 现状：UPDATE deleted_at 无该列 → SQLite 报错

## P1

### M2-4 密码强度校验缺失（SEC-4/VAL-1）
### M2-5 注册限流与设计不符（SEC-1）
### M2-6 无 logout 接口（EU-6）

## 结论
- ✅ 已达标：注册/登录/刷新接口可用、bcrypt、防用户枚举、登录限流（固定窗口）
- ❌ 验收核心场景（角色声明、踢出、删除用户）失败
