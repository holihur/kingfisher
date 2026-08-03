# M7 生产就绪 — 设计与实现差异

> 来源：`design/M7-production-ready/design.md` 对照实现
> 排查日期：2026-08-03 ｜ 详见各模块 issue

## P0

### M7-1 docker-compose 缺失（DEP-2）
- 期望：`docker-compose up -d` 一键起 mysql+redis+app(+jaeger/prometheus/grafana)
- 现状：无 compose 文件，命令必失败

### M7-2 CI/CD 缺失（DEP-3）
- 期望：GitHub Actions lint+test+build
- 现状：`.github/` 不存在

## P1

### M7-3 可观测性缺失（OBS-1/2/3/4）
### M7-4 性能基准无载体（PB-1）
### M7-5 测试金字塔未达成（TEST-1）
### M7-6 ADR 部分未兑现（ADR-1/2）
### M7-7 读写分离未启用（RW-1，设计标注可选）

## 结论
- ✅ 已达标：Dockerfile（根目录）、nginx.conf、Makefile build/test/lint
- ❌ M7 一键启动与 CI 两大验收全部缺失，生产就绪不成立
