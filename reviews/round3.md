# Kingfisher 项目评审 — Round 3（修复核查）

> 评审日期：2026-07-31
> 评审对象：round2 之后 15 个设计文件 + 新增 `reviews/round2-fixes.md`（声称 35/35 修复）
> 方法：逐项对照修复表，回到原始文档验证代码/数字/接口是否真实落地

---

## 一、结论速览

| 评分对象 | 分数 | 一句话结论 |
|----------|------|------------|
| 设计文档 | **86 / 100** | 有真实进步（错误码/JWT/wire/部署），但修复表可信度差：35 项声称修复，实际仅 11 项落地 |
| 项目整体（参考） | 65 / 100 | 仍为纯设计阶段 |

**最重要的发现**：`round2-fixes.md` 把 35 个问题全部标 ✅，实际核查结果约为 **11 修 / 10 半修 / 14 未修**。修复表只有结论、没有改动证据，后续若按"全部已修"直接进 M1，会在端点契约和种子数据上立刻踩坑。

---

## 二、35 项修复核查表

### ✅ 已核实修复（11 项）

| # | 组 | 问题 | 证据 |
|---|----|------|------|
| 1 | A | 10107→429 | `errcode` HTTPStatus 已加 `code==10007 || code==10107 → 429` |
| 2 | A | 10008→405 | 已加 `code==10008 → 405` |
| 3 | A | 密码错误码 | 新增 10108/10109/10110 |
| 8 | B | 分页 pageSize | `request/design.md` getList 已改 `page_size` |
| 16 | D | wire 签名 | di 与 startup 统一为 `InitializeApp("config/config.yaml")` |
| 17 | D | startup r 未定义 | 已补 `r := coreRouter.NewEngine(cfg, logger)` |
| 18 | D | Module 接口命名 | overview 统一 `core.Module` + `Register`，且注册了 5 个模块（含 audit） |
| 23 | E | JWT 缺 SessionVersion | Claims 加 `SessionVersion int json:"sv"`，GenerateToken 加参数 |
| 28 | G | 侧边栏路径拼接 | 已改 `item.path.startsWith("/") ? item.path : parentPath + "/" + item.path` |
| 31 | H | compose 缺 frontend/grafana | 已补两个服务 |
| 35 | H | bench.sh 占位伪代码 | CI 段已补真实 vegeta 命令 |

### ⚠️ 部分修复（10 项）

| # | 组 | 问题 | 现状 |
|---|----|------|------|
| 7 | B | 用户角色字段 | 未见修改（validation/frontend-user 不在改动文件内），仍 code vs role_id |
| 11 | C | 权限 14→15 | acceptance 已改 15，但**种子 SQL 仍只有 14**（无 audit:list/15），down.sql 未清理 |
| 12 | C | 配置 4→5 | acceptance 仍"返回 4 个预设配置"，种子仍 4 项；config 文档预设 5 项 |
| 13 | C | 迁移数字 | acceptance 已改 10 SQL/8 表；PROGRESS 仍"migrations/ (8 个 SQL 文件)" |
| 19 | D | 中间件顺序 | overview+middleware 已统一（并新增 SecurityHeaders）；**M1 里程碑仍写旧顺序** "Trace→Recovery→…→RateLimit"，M1/PROGRESS 文件清单也未加 security_headers.go |
| 21 | D | internal/extends 路径 | scripts 已修；guardrails（4 处）、perf-bench（2 处）、swagger-checklist（2 处）仍用 `internal/extends` |
| 25 | F | 降级策略 | L241 加了"miss 回源 vs 连接失败拒绝"的澄清，但 L250 Chaos 仍写"缓存 Get 失败→回源 DB 请求正常"，仍与 503 决策冲突 |
| 27 | G | 环境变量 | request 改用 `VITE_API_BASE_URL`(axios baseURL)，local-dev 保留 `VITE_API_TARGET`(proxy)；overview 仍写 `VITE_API_TARGET`，两变量分工未说明（baseURL 若配绝对地址则绕过 proxy，CORS 语义变） |
| 30 | G | 文件清单 | overview 仍列 dynamic.tsx/app.ts/permission.ts/IconSelect.tsx，PROGRESS 未同步 |
| 32 | H | 前端 Dockerfile/nginx | compose 加了 frontend，但 `build: context: ./kingfisher-web + dockerfile: deploy/Dockerfile` 引用的是**后端** Dockerfile；nginx.conf 仍无文档 |

### ❌ 声称已修但未落地（14 项）

| # | 组 | 问题 | 证据 |
|---|----|------|------|
| 4 | B | GET /users/me | `extends/user` Handler 仍无 GetMe，api-contract 仍列出该端点 |
| 5 | B | GET /roles/:id/{permissions,menus} | `extends/rbac` 仍只有 PUT 分配，无读取端点 |
| 6 | B | GET /me/permissions | 前后端文档均无；前端 auth store 仍调 `userApi.getPermissions()` |
| 9 | B | config DELETE 漏列 | api-contract Config 段仍只有 GET/GET/:key/PUT/:key |
| 10 | C | 审计迁移编号 | `extends/audit` 仍写 `000009_create_audit_logs`，migration 文档是 `000010` |
| 14 | C | 种子角色 ID 跳号 | seed 仍为 (1, 3, 4)，前端 rbac 表格画 1/2/3 |
| 15 | C | M6 无审计页面 | M6/overview/PROGRESS 均无 audit 页面任务 |
| 20 | D | transaction 跨模块冲突 | transaction 仍用 `roleRepo`（来自 extends/rbac），acceptance L185 仍禁 `extends/user` import rbac |
| 22 | D | 命名漂移 | mysql 仍 `NewMySQL`、redis 仍 `NewRedis`；di/startup 用 `NewGorm`/`NewRedisClient` |
| 24 | E | cache key 缺 user:sv | cache 文档仍无该 key 约定（extends/user 实际使用） |
| 26 | G | 前端 any | 6 份前端文档共 16 处 `any` 原样保留（request 的 `<any, ApiResponse<...>>`、`catch (err: any)` 等） |
| 29 | G | config remark | frontend/config 仍 `remark: ''  // 从另一个接口补充`，后端 GetAll 仍返回 map 无 remark |
| 33 | H | CI 三处不一致 | deploy 文档 CI 仍是精简 3-job（lint/test/build），guardrails 的 8 层 + frontend + Playwright 未同步 |
| 34 | H | 迁移执行策略冲突 | bootstrap"生产手动迁移"与 compose `/docker-entrypoint-initdb.d` 首次自动迁移并存 |

---

## 三、修复引入的新问题

1. **503 无错误码映射**：acceptance 四处出现"返回 503"（M2 L175、M3 L241/251、M7 L659），但 errcode 的 `HTTPStatus()` 没有 503 分支，且 503 响应体写 `code:10006`——而 `HTTPStatus(10006)` 映射的是 **500**。需要新增如 `ErrServiceUnavailable`（如 10009）→ 503。
2. **acceptance L241 文本合并瑕疵**：同一行出现两段重复的 `{"code":10006,...}` 与"不降级放行"，疑似合并冲突未清理。
3. **SecurityHeaders 引入未同步**：middleware 文档与 overview engine 已加，但 M1 里程碑"6 个中间件"、PROGRESS M1 文件清单未更新（现在应为 8 个 core 中间件）。
4. **frontend 服务 Dockerfile 引用错误**：compose 里 frontend 用 `dockerfile: deploy/Dockerfile`（后端多阶段构建），配 `context: ./kingfisher-web`，构建必然失败——应指向独立的 `deploy/Dockerfile.frontend`。

---

## 四、评分明细

| 维度 | 权重 | 得分 | 依据 |
|------|------|------|------|
| 架构与模块设计 | 25 | 23 | Module/Register 统一 + audit 入册；残留 transaction 跨模块冲突 |
| 可执行性 | 20 | 17 | wire/startup 已通；但 3 个前端依赖端点缺失、frontend Dockerfile 引用错误 |
| 验收与测试设计 | 20 | 17 | 错误码映射已修；降级策略仍自相矛盾、503 无映射 |
| 完整度 | 20 | 17 | 种子/迁移/前端审计页仍缺；修复表 14 项未落地 |
| 一致性与细节 | 15 | 12 | 修复表可信度差（35/35 声明 vs 11 实际）；新引入 4 处瑕疵 |
| **合计** | **100** | **86** | |

> round2 = 85 → round3 = 86。真实修复（错误码、JWT sv、wire、startup、Module、compose、路径拼接、bench）带来 +2 实质提升；修复表失实与新增瑕疵约 -1。

---

## 五、下轮建议（评审只记录，不代改）

1. **修复表改证据制**：`round2-fixes.md` 每项必须附"改动文件 + 关键行"，验收时逐条复核——否则 35/35 ✅ 会误导进入实现阶段。
2. **优先补 14 项未落地**：尤其 B 组 3 个端点（或删前端依赖）、C 组种子/编号、G26 any、H33/34。
3. **补 503 错误码**并清理 L241 重复句。
4. **修 frontend Dockerfile 引用**，补 nginx.conf 文档。
5. M1/PROGRESS 文件清单与 SecurityHeaders/8 个中间件同步。

---

## 六、评审记录

- 核查范围：round2 后修改的 15 个文件（按 mtime 识别）+ round2-fixes.md + PROGRESS.md。
- 验证方式：逐项回到原始文档 grep/精读，不以修复表结论为准。
- 本报告"部分修复/未修复"判定均以文档当前内容为准，未含猜测。
