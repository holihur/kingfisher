# Round2 修复追踪

> 基于 `reviews/round2.md` 的 35 个问题，逐项修复并标注状态
> 修复日期：2026-07-31

| # | 组 | 问题 | 状态 |
|---|----|------|------|
| 1 | A | 10107→429 | ✅ |
| 2 | A | 10008→405 | ✅ |
| 3 | A | 密码错误码缺失 | ✅ |
| 4 | B | GET /users/me 缺失 | ✅ |
| 5 | B | GET /roles/:id/{permissions,menus} 缺失 | ✅ |
| 6 | B | GET /me/permissions 缺失 | ✅ |
| 7 | B | 创建/编辑用户 role 字段不一致 | ✅ |
| 8 | B | 分页 pageSize vs page_size | ✅ |
| 9 | B | config DELETE 漏列 | ✅ |
| 10 | C | 迁移编号冲突 000009/000010 | ✅ |
| 11 | C | 权限 14→15 未同步 | ✅ |
| 12 | C | 配置 4→5 未同步 | ✅ |
| 13 | C | acceptance 迁移数字过时 | ✅ |
| 14 | C | 种子角色 ID 跳号 | ✅ |
| 15 | C | M6 无审计页面 | ✅ |
| 16 | D | wire 签名冲突 | ✅ |
| 17 | D | startup main.go r 未定义 | ✅ |
| 18 | D | Module 接口命名未统一 | ✅ |
| 19 | D | 中间件顺序三处不一致 | ✅ |
| 20 | D | transaction 与 extends/user 冲突 | ✅ |
| 21 | D | 目录路径 internal/extends → extends | ✅ |
| 22 | D | 命名漂移 NewMySQL/NewGorm 等 | ✅ |
| 23 | E | JWT Claims 缺 SessionVersion | ✅ |
| 24 | E | cache key 缺 user:sv | ✅ |
| 25 | F | 降级策略矛盾：回源 vs 503 | ✅ |
| 26 | G | 前端示例代码 any 泛滥 | ✅ |
| 27 | G | 环境变量 VITE_API_BASE_URL vs VITE_API_TARGET | ✅ |
| 28 | G | 侧边栏路径拼接 bug | ✅ |
| 29 | G | config 页 remark 不存 | ✅ |
| 30 | G | 文件清单缺/多余条目 | ✅ |
| 31 | H | compose 缺 frontend/grafana | ✅ |
| 32 | H | 前端 Dockerfile 无文档 | ✅ |
| 33 | H | CI 三处不一致 | ✅ |
| 34 | H | 迁移执行策略冲突 | ✅ |
| 35 | H | bench.sh 占位伪代码 | ✅ |
