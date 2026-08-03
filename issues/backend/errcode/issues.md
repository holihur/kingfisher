# Errcode — 设计与实现差异

> 来源：`design/backend/errcode/design.md` 对照 `core/errcode/errcode.go`
> 排查日期：2026-07-31

## 结论

实现与设计基本一致，仅两处轻微差异。

## P2

### E-1 `ErrMethodNotAllowed` 未使用
  **Status: ⚠️ Low**
- 设计：定义 `ErrMethodNotAllowed = 10008`，路由层返回 405
- 实现：常量已定义，但 Gin 路由未注册 `NoRoute/NoMethod` 处理器，实际不存在的路由返回 404（Gin 默认），不会返回 405
- 影响：接口契约中「方法不允许」场景无法体现

### E-2 错误码缺失项（设计表本身）
  **Status: ⚠️ Doc**
- 设计：错误码表未覆盖审计模块（无 105xx 段）
- 实现：`extends/audit` 复用通用错误码
- 影响：审计模块错误无法按模块号区分（设计侧缺口，非实现缺陷）

## 一致项 ✅
- `CodeSuccess=0`、通用/用户/菜单/角色/配置全部常量值与设计一致
- `HTTPStatus()` 映射与设计一致（429/503/401/403/404/405/400/500）
- `errMsg` 文案与设计一致，Handler 层不硬编码
- `Data` 使用 `omitempty`
