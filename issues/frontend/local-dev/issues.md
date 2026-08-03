# Frontend Local Dev — 设计与实现差异

> 来源：`design/frontend/local-dev/design.md` 对照 `web/vite.config.ts`
> 排查日期：2026-07-31

## P1

### FLD-1 未用 loadEnv 分层
- 设计：`loadEnv(mode, cwd)` + `VITE_API_TARGET` 环境变量决定 proxy target
- 实现：proxy target 硬编码 `http://localhost:8080`
- 影响：切换后端地址需改配置

### FLD-2 无 .env 文件
- 设计：`.env.development` / `.env.production`（VITE_API_BASE_URL 等）
- 实现：`web/` 下无任何 `.env*` 文件（与 FOV-3、FRQ-1 联动）
- 影响：环境变量分层机制形同虚设

### FLD-3 无 docker-compose.dev.yaml
- 设计：本地一键起全栈（前端+后端+MySQL+Redis）
- 实现：无该文件（见 DEP-2）
- 影响：M5 联调环境需手动起服务

## P2

### FLD-4 `open: true` 未设置
- 设计：dev server 自动打开浏览器
- 实现：`server` 未配置 `open`
- 影响：需手动打开 5173

## 一致项 ✅
- port 5173 ✅、`@` alias ✅、`/api` proxy（changeOrigin）✅、`/swagger` proxy（额外）✅
