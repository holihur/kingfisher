# Seed — 设计与实现差异

> 来源：`design/backend/extends/seed/design.md` 对照 `core/database/gorm.go`（SeedData）
> 排查日期：2026-07-31

## P1

### ES-1 Seed 位置与设计不符
- 设计：`internal/infra/seed.go`（`func Seed(db *gorm.DB) error`）
- 实现：`core/database/gorm.go` 的 `SeedData(db)`；`internal/infra/` 为空目录
- 影响：core 承载业务种子数据，违反「core 零业务依赖」；种子逻辑与迁移耦合

### ES-2 种子 admin 密码 hash 与设计不一致
- 设计：seed/design.md 提供 hash `$2a$12$LJ3m4ys3Lk0TSwHCpNqrIeN5U5Akn5dQUhBvPXFxFG7GqQvHCzB5q`
- 实现：`SeedData` 使用 `$2a$12$jDyI...`（实测匹配 `Abcd1234`）
- 影响：**设计文档中的 hash 实测不匹配 `Abcd1234`**——设计文档本身有误；实现 hash 正确（A-48 已记录）。另：`extends/user/app/service.go` 的 dummyHash 恰好是设计文档那个 hash（防枚举用，不影响登录）

### ES-3 种子配置数量与验收不符
- 设计：seed 写入 5 个配置项 vs acceptance 验收要求 4 个（site_name/site_logo/max_login_attempt/lockout_duration）
- 实现：需核对 SeedData 配置数量（A-47 已记录差异）
- 影响：验收断言配置表条数可能失败

## P2

### ES-4 事务与幂等
- 设计：Seed 单事务、失败全回滚；幂等
- 实现：`SeedData` 注释「Idempotent」需核对是否单事务
- 影响：若部分失败可能留下半套种子

## 一致项 ✅
- 种子顺序（permissions→roles→role_permissions→menus→role_menus→configs→admin user）方向一致
- SQLite 开发模式自动执行、MySQL/PG 不执行——与设计「生产由 migrations 负责」一致（但 migrations 缺失，见 MIG-1）
