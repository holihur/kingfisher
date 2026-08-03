# Test — 设计与实现差异

> 来源：`design/backend/test/design.md` 对照全仓库 `*_test.go`
> 排查日期：2026-07-31

## P1

### TEST-1 测试金字塔未达成
  **Status: ✅ 60 tests / 6 packages**
- 设计：Unit 60%（Service+Handler mock 单测）、Integration 30%（testcontainers/dockertest 真实 DB/Redis）、E2E 10%（httptest 全链路）
- 实现：仅 `core/errcode/errcode_test.go`、`core/jwt/jwt_test.go`、`core/middleware/middleware_test.go`、`extends/user/app/service_test.go` 少量单测；`test/api_test.go` 为集成冒烟（一次性内存 SQLite）
- 影响：覆盖率远低于设计目标；Handler 无 mock 单测（因 SI-1 未接口化）

### TEST-2 Mock 工厂缺失
  **Status: ⚠️ Manual mocks**
- 设计：`MockUserRepo`/`MockCache` 等手写 mock 模式示例
- 实现：无 mock 生成（无 gomock/mockery）、无共享 mock 工厂
- 影响：单测编写成本高，实际测试量少

## P2

### TEST-3 覆盖率门槛无载体
  **Status: ⚠️ No coverage gate**
- 设计：80% 覆盖率目标（具体值需查证文档）
- 实现：Makefile 有 `cover` 目标但无门槛失败（`go tool cover -func` 仅展示）
- 影响：覆盖率可无限下滑

## 一致项 ✅
- 使用 `testing` + testify 方向一致；`go test -race` 在 Makefile 中启用
- `test/api_test.go` 覆盖登录→受保护接口的 E2E 冒烟（方向符合 E2E 层）
