# Extends/Menu — 设计与实现差异

> 来源：`design/backend/extends/menu/design.md` 对照 `extends/menu/`
> 排查日期：2026-07-31

## P0

### EM-1 ✅ 菜单树未按角色过滤 ✅ handler 层 filterTreeByPerms 权限过滤
- 设计：`GetTree` 按角色权限返回菜单（viewer 只剩 Dashboard）
- 实现：`GetTree` 查全表 → buildTree，无角色过滤、无权限过滤
- 影响：M3 验收「editor/viewer 菜单过滤」失败（A-4）；所有角色看到相同菜单

## P1

### EM-2 ✅ 菜单缓存缺失
- 设计：Cache-Aside `menu:tree` 10min
- 实现：无缓存（见 CA-1）
- 影响：每次请求全量查菜单表

### EM-3 ✅ Create/Update 无校验
- 设计：菜单名/路径/权限标识唯一性、父级存在性校验
- 实现：直接 repo.Create/Update
- 影响：重复路径、悬空父级可写入

### EM-4 ✅ Delete 无子节点保护
- 设计：`ErrMenuHasChildren`——有子节点不可删
- 实现：`MenuService.Delete` 直接 repo.Delete；`ErrMenuHasChildren` 常量未使用
- 影响：级联孤儿菜单（A-39）

## P2

### EM-5 ✅ 无 port 接口
- 设计：`port/repository.go` 定义 MenuRepository
- 实现：Service 依赖 `*adapter.MenuRepo`（见 IF-2）

### EM-6 ✅ `HasChildren` 未使用
- 设计：删除前查子节点
- 实现：无此方法调用（repo 层可能未实现）

## 一致项 ✅
- buildTree 递归构建（ParentID=0 顶级）与设计一致
- Menu domain 字段（Type 1/2/3、Permission、Status、Children）与设计一致
