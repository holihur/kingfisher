# Deploy — 设计与实现差异

> 来源：`design/backend/deploy/design.md` 对照 `deploy/`、根目录 Dockerfile、Makefile
> 排查日期：2026-07-31

## P0

### DEP-1 Dockerfile 位置与内容偏差
  **Status: ✅ Deploy infrastructure complete**
- 设计：多阶段 Dockerfile 位于 deploy/，含 `-ldflags` 版本注入、HEALTHCHECK、migrations 拷贝、`alpine:3.20`
- 实现：根目录有 `Dockerfile`，内容需核对；`deploy/` 下仅 `nginx.conf`
- 影响：M7「docker-compose up -d」验收失败（无 compose）

### DEP-2 docker-compose.yaml 缺失
  **Status: ✅ docker-compose.yaml created**
- 设计：mysql + redis + jaeger + prometheus + grafana + app 全栈 compose，`depends_on: condition: service_healthy`
- 实现：无 docker-compose.yaml（根目录/deploy/ 均无）；无 `deploy/prometheus.yaml`
- 影响：M7 一键启动验收失败；observability 配套服务缺失（OBS-4）

### DEP-3 CI/CD 缺失
  **Status: ✅ GitHub Actions CI created**
- 设计：GitHub Actions CI（lint + test + build + docker 推送）
- 实现：`.github/` 不存在
- 影响：guardrails 设计「CI 强制执行」无载体

## P1

### DEP-4 Makefile docker 目标缺失
  **Status: ✅ Deploy infrastructure complete**
- 设计：`make docker-build / docker-up / docker-down`
- 实现：Makefile 无 docker 目标
- 影响：部署命令无统一入口

## P2

### DEP-5 nginx.conf 未对齐
  **Status: ✅ Deploy infrastructure complete**
- 设计：部署章节未详细定义 nginx；`deploy/nginx.conf` 为前端产物代理（需人工核对与设计一致性）
- 影响：低

## 一致项 ✅
- `config/config.yaml` 生产可用配置存在；`make build` 可产出 bin/server
