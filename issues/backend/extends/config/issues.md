# Extends/Config — 设计与实现差异

> 来源：`design/backend/extends/config/design.md` 对照 `extends/config/`
> 排查日期：2026-07-31

## P0

### EC-1 无权限校验
- 设计：写操作需 `config:update` 权限
- 实现：`register.go` 的 PUT/DELETE 未挂 `RequirePerm`（A-3）
- 影响：viewer 可修改系统配置

## P1

### EC-2 缓存未实现
- 设计：Write-Through（写 DB + 写缓存）；`config:all` 缓存
- 实现：无缓存（见 CA-1/CA-2）
- 影响：每次请求直查 DB

### EC-3 读接口未做按键缓存失效联动
- 设计：`Set`/`Delete` 后失效相关缓存
- 实现：无缓存联动（与 EC-2 同源）

## P2

### EC-4 `Set` 用 `FirstOrCreate` 的 upsert 行为
- 设计：`PUT /configs/:key` 语义为「新增或更新」需明确
- 实现：`ConfigRepo.Set` 需核对（若用 FirstOrCreate，`gorm.io` 对 struct 的 Updates 可能不更新零值）
- 影响：value 设为空字符串等场景可能不生效（需验证）

### EC-5 无 port 接口
- 设计：`port/repository.go` 定义 ConfigRepository
- 实现：Service 依赖 `*adapter.ConfigRepo`（见 IF-3）

## 一致项 ✅
- SystemConfig domain（key/value/remark）与设计一致
- 预设配置项（site_name 等）种子已写入（数量差异见 A-47/ES-3）
