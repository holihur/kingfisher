# Frontend Config — 设计与实现差异

> 来源：`design/frontend/config/design.md` 对照 `web/src/pages/config/ConfigManage.tsx`
> 排查日期：2026-07-31

## P2

### FC-1 动态控件未实现
- 设计：按 Value 类型区分控件（文本/数字/开关/JSON 编辑器等）
- 实现：编辑弹窗仅 `Input` 文本框，无类型区分
- 影响：数字类配置（max_login_attempt）体验一般（设计标注为增强）

## P2

### FC-2 无删除入口
- 设计：表格操作含编辑（删除未明确）
- 实现：仅编辑按钮；后端有 `DELETE /configs/:key` 但前端无入口
- 影响：功能未暴露（影响低）

## 一致项 ✅
- 列表（Key/Value/备注）+ 编辑弹窗（Key 不可改）✅
- 编辑按钮按 `config:update` 权限渲染 ✅
- `configApi.get/set` 对接后端 ✅
