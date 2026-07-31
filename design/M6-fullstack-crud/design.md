# M6 — 全栈 CRUD 闭环

## 目标

在浏览器上完整操作 用户/菜单/角色/配置 的增删改查。

## 覆盖模块

| 模块 | 设计文档 | 说明 |
|------|----------|------|
| 用户管理 | [user](../frontend/user/design.md) | ProTable + Modal 表单 + 搜索分页 |
| 菜单管理 | [menu](../frontend/menu/design.md) | 树形 Table + TreeSelect 父级 |
| 角色权限 | [rbac](../frontend/rbac/design.md) | Tabs 权限分配 + Tree 菜单分配 |
| 系统配置 | [config](../frontend/config/design.md) | 键值对编辑 + 动态控件 |
| 审计日志 | [audit](../../design/backend/extends/audit/design.md) | 操作记录查询按资源/操作筛选 |
| 类型共享 | [shared-types](../../shared-types/design.md) | Swagger → TS 类型自动生成 |

## 验证

```bash
# 用户管理
/ 点击 "系统管理 > 用户管理"
# → 表格显示所有用户（admin 在第一条）
# → 搜索框输入 "admin"，点击搜索，过滤成功
# → 点击 "新增用户" → 填写表单 → 提交 → 表格刷新
# → 点击某行的 "编辑" → 修改 email → 提交 → 数据更新
# → 点击 "删除" → 确认 → 该行消失

# 菜单管理
# → 展开树形表格，看到 15 条菜单
# → 点击系统管理行的 "添加子项" → 表单一填 → 提交 → 树刷新
# → 编辑菜单名称 → 提交 → 更新成功

# 角色管理
# → 看到 3 个角色（admin/editor/viewer）
# → 点击 admin 行的 "权限" → 弹窗 Tab 页显示 4 组 14 个权限，全部勾选
# → 点击 editor 行的 "菜单" → Tree 显示 15 个菜单 → 勾选 Dashboard + 用户管理 → 提交
# → viewer 登录后侧边栏只剩 Dashboard

# 系统配置
# → 看到 4 个配置项
# → 编辑 site_name → 保存 → 页面标题变化
```

## 产出文件

| 文件 | 说明 |
|------|------|
| `kingfisher-web/src/api/user.ts` | 用户 API |
| `kingfisher-web/src/api/menu.ts` | 菜单 API |
| `kingfisher-web/src/api/role.ts` | 角色 API |
| `kingfisher-web/src/api/config.ts` | 配置 API |
| `kingfisher-web/src/pages/user/UserList.tsx` | 用户列表 |
| `kingfisher-web/src/pages/user/UserForm.tsx` | 用户表单 |
| `kingfisher-web/src/pages/menu/MenuManage.tsx` | 菜单管理 |
| `kingfisher-web/src/pages/menu/MenuForm.tsx` | 菜单表单 |
| `kingfisher-web/src/pages/role/RoleList.tsx` | 角色列表 |
| `kingfisher-web/src/pages/role/RoleForm.tsx` | 角色表单 |
| `kingfisher-web/src/pages/role/PermissionModal.tsx` | 权限分配 |
| `kingfisher-web/src/pages/role/MenuAssignModal.tsx` | 菜单分配 |
| `kingfisher-web/src/pages/config/ConfigManage.tsx` | 配置列表 |
| `kingfisher-web/src/pages/config/ConfigEditModal.tsx` | 配置编辑 |
| `kingfisher-web/src/pages/dashboard/index.tsx` | Dashboard |
| `kingfisher-web/src/pages/audit/AuditLogList.tsx` | 审计日志 |
## 验收

→ [acceptance 验收清单](../acceptance/design.md)
