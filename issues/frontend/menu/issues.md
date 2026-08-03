# Frontend Menu — 设计与实现差异

> 来源：`design/frontend/menu/design.md` 对照 `web/src/pages/menu/MenuManage.tsx`
> 排查日期：2026-07-31

## P1

### FM-1 TreeSelect 父级数据为空
- 设计：选择上级菜单时树形可选全部目录/菜单
- 实现：`treeData={[{ title: '根菜单', value: 0, children: [] }]}`——只有根，无法选择已有菜单为父级
- 影响：无法创建二级以上菜单（只能通过「添加子项」间接实现）

## P2

### FM-2 删除不禁用有子节点项
- 设计：有子节点的菜单删除按钮禁用/拦截（后端 ErrMenuHasChildren）
- 实现：Popconfirm 提示「有子节点将无法删除」但按钮可点；后端也未校验（EM-4）
- 影响：删除父级产生孤儿数据

### FM-3 无 URL 展开同步
- 设计：根据当前 URL 自动展开树
- 实现：树为扁平 ProTable（flatten），无展开概念
- 影响：设计交互（树形展开）被扁平表格替代，体验差异

## P2

### FM-4 拖拽排序未实现（可选）
- 设计：拖拽排序（可选）
- 实现：无 dnd 依赖
- 影响：设计标注可选，低优先级

## 一致项 ✅
- 树形展示（缩进 + 类型 Tag 目录/菜单/按钮）✅
- 新增根菜单/添加子项/编辑/删除 + 权限控制（menu:create/update/delete）✅
- 表单含 名称/路由/图标/排序/类型/权限标识/状态 等字段 ✅
