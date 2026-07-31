# Design Changelog

> 设计文档变更记录。

## 2026-07-31

### P0 遗留硬伤
- acceptance L250: 区分缓存 miss（回源DB）vs 连接失败（503）
- api-contract: 删除 Role 段重复端点行
- migration down.sql: 清理范围补 session_timeout

### P1 一致性闭环
- 审计模块入册（PROGRESS + acceptance + M6）
- deploy CI: 声明权威来源为 guardrails
- bootstrap: 迁移策略定版（dev compose 自动 / prod 手动）
- migration: 审计 SQL 交叉引用 extends/audit

### 多数据库驱动
- config: database.driver 支持 sqlite/mysql/postgres
- mysql doc → database doc 重写：三驱动 DSN + 连接池参数
- 开发环境默认 SQLite，零依赖启动

### URL 状态同步
- frontend overview §5: 进 URL vs 不进 URL 边界
- ProTable syncToUrl 实现

### P2 治理
- design/scripts/check-design.sh: 7 项自动一致性检查
- CHANGELOG.md: 设计变更记录
- 8 次 git commit
