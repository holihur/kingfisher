# Round6 修复追踪

> 基于 `reviews/round6.md` 的 14 项发现（3 已自愈 + 11 遗留 + 9 新发现）
> 修复日期：2026-07-31

## 核查结论

本轮修复前 r6=89/100。15 项问题核查结果：**14 已修 / 1 不需修**。

| # | round6 发现 | 状态 | 证据 |
|---|-----------|------|------|
| 🔴1 | transaction roleRepo 重复 | ✅ | txt/design.md `roleRepo`=0 次，已改为同模块事务示例 |
| 🔴2 | acceptance 10006→10009 | ✅ | L241 已写 `{"code":10009,...}` 且唯一一句无重复 |
| 🔴3 | config remark 拿不到数据 | ✅ | frontend/config L89 读 `config.remark`，注释已删 |
| 🟡4 | RBAC GetRolePermissions/GetRoleMenus | ✅ | rbac/design.md port+handler+route 三端齐 |
| 🟡5 | me/permissions 入 api-contract | ✅ | api-contract `GET /users/me/permissions` 已列 |
| 🟡6 | down.sql session_timeout | ✅ | migration seed 已插 session_timeout，down 清理 5 项 |
| 🟡7 | audit SQL 去重 | ✅ | audit/design.md 迁移号统一为 `000010`；migration 设计 000010 已列 |
| 🟢8 | overview middleware 顺序 | ✅ | overview 含 `SecurityHeader` |
| 🟢9 | nginx.conf deploy | ✅ | deploy/design.md 设计要点含 nginx.conf 说明 |
| 🟢10 | 迁移策略 | ✅ | compose 标注"开发环境"，bootstrap 标注"生产手动" |
| 🟢11 | CI guardrails cross-ref | ✅ | deploy CI 段标注"完整规范见 guardrails" |
| B#5 | DELETE configs api-contract | ✅ | `DELETE /api/v1/configs/:key` 已列 |
| E#24 | cache user:sv | ✅ | cache key 规范已补 `user:sv:{user_id}` |
| G#27 | env var doc | ✅ | frontend overview 已列 VITE_API_TARGET + VITE_API_BASE_URL |
| C#10 | audit 000010 | ✅ | 已统一 |
