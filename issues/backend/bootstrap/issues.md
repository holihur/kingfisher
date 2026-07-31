# Bootstrap — 设计与实现差异

> 来源：`design/backend/bootstrap/design.md` 对照仓库根目录、Makefile、scripts/
> 排查日期：2026-07-31

## P1

### BS-1 工具链安装目标未实现
- 设计：`make setup` 安装 swag / wire / golangci-lint
- 实现：Makefile 无 `setup` 目标（有 run/build/test/lint/fmt/cover/wire/clean 等）
- 影响：新成员按设计执行 `make setup` 直接报错

### BS-2 环境变量模板缺失
- 设计：`.env.example` / 环境变量模板（JWT_SECRET、VITE_API_BASE_URL 等）
- 实现：根目录与 web/ 下均无 `.env*` 模板
- 影响：配置项靠人工发现，易漏配

### BS-3 快速启动文档承诺与实现不符
- 设计：`make run` 5 分钟跑起来（自动迁移+种子）
- 实现：`make run` 依赖本地 Go 工具链；首次运行 SQLite 自动建表+种子可用，但 MySQL/PG 无法自动迁移
- 影响：文档声称的「从零到可运行」仅 SQLite 场景成立

## P2

### BS-4 版本端点配套
- 设计：`/version` 含 version/commit/build_time
- 实现：已实现 ✅（`main.go`），但 Makefile `build` 的 `-ldflags` 只注入 `-s -w`，未注入 version/commit/build_time（恒为 dev/unknown）
- 影响：发布产物无法追踪版本

## 一致项 ✅
- 目录骨架（cmd/core/extends/web/scripts/config/deploy）与设计一致
