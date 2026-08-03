# Frontend User — 设计与实现差异

> 来源：`design/frontend/user/design.md` 对照 `web/src/pages/user/UserList.tsx`
> 排查日期：2026-07-31

## P1

### FU-1 ✅ 无角色列
- 设计：表格列含「角色」（管理员/编辑/访客）
- 实现：列仅 ID/用户名/邮箱/状态/创建时间/操作，无角色列（后端 List 未返回 role 信息，见 DOM-1）
- 影响：M6 验收「角色列」缺失

### FU-2 ✅ 新增用户必然失败（后端 404）
- 设计：新增用户表单提交成功
- 实现：`userApi.create` 调 `POST /users`，后端未注册（EU-2/AC-1）
- 影响：M6「新增用户」点击提交必 404

### FU-3 ✅ 表单未独立成文件
- 设计：`pages/user/UserForm.tsx` 独立表单组件
- 实现：表单内嵌在 `UserList.tsx` 的 Modal 中
- 影响：目录结构与设计不一致；复用性差

## P2

### FU-4 ✅ 状态展示用 Tag 而非 Badge
- 设计：状态用 Badge（绿点/红点）
- 实现：`<Tag color>启用/禁用</Tag>`
- 影响：视觉细节差异

### FU-5 ✅ 搜索无防抖
- 设计：ProTable 搜索（防抖/enter 触发）
- 实现：依赖 ProTable 默认行为，未显式防抖配置
- 影响：轻微（输入即查询）

## 一致项 ✅
- ProTable + 搜索（keyword）+ 分页（page/page_size）✅
- 编辑/删除按钮按权限渲染（user:update/user:delete）✅
- 删除自己二次确认（后端拦截，前端 Popconfirm）✅
- 编辑表单（用户名/邮箱/状态/角色选择）基本齐全 ✅
