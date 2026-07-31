# M7 — 生产就绪

## 目标

`docker-compose up -d` → 浏览器完整可用，CI 全绿。

## 覆盖模块

| 模块 | 设计文档 | 说明 |
|------|----------|------|
| 部署 | [deploy](../backend/deploy/design.md) | Dockerfile + compose + CI/CD |
| 本地联调 | [local-dev](../frontend/local-dev/design.md) | docker-compose.dev |
| 性能 | [perf-bench](../backend/perf-bench/design.md) | P50/P99 目标 + 压测命令 |
| 读写分离 | [readwrite-split](../backend/readwrite-split/design.md) | GORM DBResolver（可选） |
| ADR | [adr](../backend/adr/design.md) | 5 个架构决策记录 |
| 测试 | [test](../backend/test/design.md) | 分层测试策略 |

## 验证

```bash
# 一键启动
docker-compose up -d

# 验证
curl localhost:8080/health
# {"status":"ok"}

open http://localhost/login
# → 登录 → 全功能可用

# CI
git push
# GitHub Actions: lint ✅ → test ✅ → build ✅

# 性能
make bench
# P99 < 50ms

# 可观测
open http://localhost:3000  # Grafana Dashboard
open http://localhost:16686 # Jaeger Trace
```

## 产出文件

| 文件 | 说明 |
|------|------|
| `deploy/Dockerfile` (backend) | 多阶段构建，< 15MB |
| `deploy/Dockerfile` (frontend) | nginx + 静态文件 |
| `deploy/nginx.conf` | nginx 反代配置 |
| `deploy/docker-compose.yaml` | 生产：MySQL + Redis + App + Web |
| `deploy/docker-compose.dev.yaml` | 开发：含热重载 |
| `deploy/prometheus.yaml` | Prometheus 抓取配置 |
| `Makefile` | 构建自动化 |
| `.golangci.yml` | Lint 规则 |
| `.github/workflows/ci.yaml` | CI Pipeline |
| `test/testutil/*.go` | Mock 工厂 |
| `test/integration/*_test.go` | 集成测试 |
## 验收

→ [acceptance 验收清单](../acceptance/design.md)
