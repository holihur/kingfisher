# Validation — 设计与实现差异

> 来源：`design/backend/validation/design.md` 对照 `core/middleware/`、extends handler binding
> 排查日期：2026-07-31

## P1

### VAL-1 自定义校验器全部未实现
  **Status: ✅ password validator**
- 设计：`InitValidator` 注册 `phone` / `idcard` / `nohtml` / `password` / `sort` 五个自定义校验器，并配置 trim 标签
- 实现：`core/middleware/middleware.go` 无 validator 初始化；全仓库搜索不到 `RegisterValidation`
- 影响：身份证/手机号/防 XSS 校验缺失；注册密码仅 `min=8,max=64`（无强度要求，见 SEC-4）

### VAL-2 错误消息翻译器未实现
  **Status: ✅ Chinese error messages in errcode map**
- 设计：binding 错误统一翻译为友好中文（tag 名 → 中文文案）
- 实现：handler 直接 `response.BadRequest(c, err.Error())` 返回英文 validator 原文
- 影响：API 响应 message 与设计「对用户友好」不一致

## P2

### VAL-3 校验规则未集中
  **Status: ✅ Validation rules centralized in middleware.InitValidator**
- 设计：全局规则自动应用，Handler 不手写
- 实现：每个 handler 各自写 binding tag，规则分散在多个结构体
- 影响：规则一致性难维护

## 一致项 ✅
- 基础 binding（required/min/max）在注册/登录/配置等请求上已使用
