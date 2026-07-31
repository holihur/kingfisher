# Kingfisher 项目评审 — Round 5（增量核查）

> 评审日期：2026-07-31
> 评审对象：round4 之后修改的 14 个文件 + 新增 `.gitignore`（5 份前端文档、transaction、cache、swagger-checklist、perf-bench、migration、api-contract、acceptance、PROGRESS）
> 方法：按 mtime 识别改动，逐项对照 round4 遗留清单验证，并检查是否引入新错误

---

## 一、结论速览

| 评分对象 | 分数 | 一句话结论 |
|----------|------|------------|
| 设计文档 | **88 / 100** | 本轮修复面最广（any 清零、种子定版、PROGRESS 对齐），但 transaction 文档被改出语法错误 |
| 项目整体（参考） | 65 / 100 | 仍为纯设计阶段 |

> round2=85 → round3=86 → round4=87 → round5=88。修复开始覆盖前端与种子，但验收/部署两条线仍有硬伤。

---

## 二、本轮已核实修复（9 项）

| 项 | 对应 round | 修复内容 | 证据 |
|----|-----------|----------|------|
| 1 | #26 | **前端 `any` 全部清零**（auth/config/menu/rbac/state-ui），改用 `unknown` | 6 份前端文档 0 处 any |
| 2 | round4 新#1 | auth store 改调 `userApi.getMyPermissions()` | `frontend/auth` L116 |
| 3 | #21（部分） | swagger-checklist、perf-bench 的 `internal/extends` 已清 | 两文档无残留 |
| 4 | #11 | 种子权限补第 15 个 `audit:list`，down.sql 清理范围扩到 1-15 | migration L167/L248 |
| 5 | #12 | 种子配置补 `session_timeout`（5 项齐） | migration L219 |
| 6 | #19（收尾） | PROGRESS 中间件清单补齐 8 个（gzip/security_headers 入列） | PROGRESS L37-38 |
| 7 | #29（前端侧） | 配置页 remark 改为读取后端字段 | frontend/config L89 |
| 8 | 工程基建 | 新增 `.gitignore`（依赖/产物/环境/生成代码全覆盖） | 项目根 |
| 9 | #13 | PROGRESS 迁移数 10 个 SQL | PROGRESS L109 |

---

## 三、延续未修复（round4 遗留，14 项）

| 组 | 问题 | 现状 |
|----|------|------|
| B | #5 `GET /roles/:id/{permissions,menus}` | 后端 extends/rbac 仍只有 PUT；前端 rbac 弹窗仍调用这两个不存在的端点 |
| B | #9 api-contract 漏 `DELETE /configs/:key` | 该文件本轮改过，但 Config 段仍只有 GET/GET/:key/PUT |
| B | round4 新#2 `GET /users/me/permissions` 未入契约 | api-contract User 段无此端点 |
| C | #10 审计迁移编号冲突 | extends/audit 仍 `000009`，migration 仍 `000010` |
| C | #14 种子角色 ID 仍 (1,3,4) | 未改 |
| C | #15 审计日志页面 | acceptance 无 audit-logs 接口验收，M6/PROGRESS 无页面任务 |
| D | #22 命名漂移 `NewMySQL/NewRedis` | mysql/redis 文档未动 |
| E | #24 cache 文档仍无 `user:sv` key | cache 本轮改过但 Key 规范未加 |
| F | #25 降级策略矛盾 | L241 与 L250 仍未统一（详见第四节） |
| G | #27 环境变量双轨 | overview 未动 |
| G | #30 overview vs PROGRESS 文件清单 | overview 未动 |
| H | #32 nginx.conf 无文档 | deploy 未动 |
| H | #33 deploy CI 仍 3-job | 未动 |
| H | #34 迁移策略冲突 | 未动 |

---

## 四、本轮新发现 / 恶化的问题

1. **transaction 文档被改出语法错误（最严重）**：`backend/transaction/design.md` L58-59 出现重复字段声明 `userRepo port.UserRepository` 两次；且 L69 仍调用 `s.roleRepo.AssignUser`，而 struct 中根本没有声明 `roleRepo` 字段——按此实现 Go 直接编译失败。原问题（跨模块依赖与 acceptance import 禁令冲突）也未解决。
2. **503 错误码矛盾加深**：errcode 已把 503 分配给 `10009`（round4 修复），但 acceptance L241/L251/L175/L659 四处 503 响应体仍写 `{"code":10006}`——而 `HTTPStatus(10006)` 映射的是 500。验收预期与错误码实现无法同时满足。
3. **acceptance L241 编辑痕迹残留**：本轮改动后该行出现"拒绝，。 `{json}`。——"标点错乱，且前半句"缓存 miss 的回源是正常行为"与 L250"缓存 Get 失败→回源 DB→请求正常"的边界仍与"连接失败→503"决策重叠，语义未理清。
4. **前端 rbac 调用点与后端仍脱节**：any 清零时顺手把 `p: any` 改成 `p: unknown`，但 `getRolePermissions/getRoleMenus` 这两个后端不存在的端点调用原样保留——类型干净了，功能仍断。

---

## 五、评分明细

| 维度 | 权重 | 得分 | 依据 |
|------|------|------|------|
| 架构与模块设计 | 25 | 23 | 无新增架构问题；transaction 示例代码损坏 |
| 可执行性 | 20 | 18 | any 清零 + auth store 修复；transaction 语法错误、rbac 端点缺失仍会卡编译/联调 |
| 验收与测试设计 | 20 | 17 | 种子定版推进；503 错误码矛盾、降级语义未清 |
| 完整度 | 20 | 17 | 审计页/契约/CI/nginx/迁移策略仍未动 |
| 一致性与细节 | 15 | 13 | 修复质量整体回升；新引入 transaction 重复字段与 L241 标点错乱 |
| **合计** | **100** | **88** | |

---

## 六、下轮建议（评审只记录，不代改）

1. **立即修复 transaction 文档**：删重复字段、补 `roleRepo` 声明或改为同模块事务示例，并与 acceptance"extends/user 不 import extends/rbac"二选一定版。
2. **统一 503 错误码**：acceptance 四处 `10006` → `10009`，清理 L241 重复句与标点。
3. **补齐或删除 rbac 两个 GET 端点**（`/roles/:id/{permissions,menus}`），并同步 api-contract 与前端调用。
4. **补 `DELETE /configs/:key` 与 `GET /users/me/permissions` 入 api-contract**；config remark 需后端 `GetAll` 返回完整对象（当前 map 无 remark，前端已改但拿不到数据）。
5. C 组收尾：审计迁移编号取一、角色 ID 连续、审计页任务入 PROGRESS/acceptance。
6. cache 补 `user:sv` key；H 组四项（nginx/CI/迁移策略）定版。

---

## 七、评审记录

- 核查范围：round4 后 mtime 变化的 14 个文件 + `.gitignore` + 被引用未改动的关联文件（rbac 后端、interface、bootstrap、guardrails、overview）。
- 判定标准：以文档当前内容为准；"已修"必须能在改动文件中找到对应行。
- 特别核查：对 transaction、cache、api-contract 三个"改了但可能改坏"的文件做了全文级检查。
