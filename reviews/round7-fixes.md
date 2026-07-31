# Round7 修复追踪

| # | 级别 | 问题 | 状态 | 证据 |
|---|------|------|------|------|
| P0-1 | 🔴 | acceptance L250 语义冲突 | ✅ | 区分 miss/connect-fail |
| P0-2 | 🔴 | api-contract 重复行 | ✅ | 重复已删 |
| P0-3 | 🔴 | down.sql 漏 session_timeout | ✅ | 清理 4→5 |
| P1-1 | 🟡 | 审计入册 | ✅ | PROGRESS+acceptance+M6 含 audit |
| P1-3 | 🟡 | SQL 权威来源 | ✅ | migration 引用 extends/audit |
| P1-4 | 🟡 | CI 权威 | ✅ | deploy 声明 guardrails 为准 |
| P1-5 | 🟡 | 迁移策略 | ✅ | compose dev/boot prod |
| P2-2 | 🔵 | lint 脚本 | ✅ | check-design.sh |
| P2-4 | 🔵 | CHANGELOG | ✅ | 已创建 |
| P2-5 | 🔵 | Git | ✅ | 8 commits |
