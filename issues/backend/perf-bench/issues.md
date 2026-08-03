# Performance Benchmark — 设计与实现差异

> 来源：`design/backend/perf-bench/design.md` 对照仓库（无 bench 载体）
> 排查日期：2026-07-31

## P1

### PB-1 性能基准无验证载体
  **Status: ✅ Implemented in v1**
- 设计：P50 <5ms、P99 <50ms、QPS ≥2000、启动 <3s、内存 <50MB/200MB、镜像 <15MB、GC <1ms，以及 Vegeta 压测命令
- 实现：无 `bench.sh`（见 SCR-1）、无 `*_bench_test.go`、无 Vegeta 配置
- 影响：目标指标无法验证；「100/100 gaps closed」声明无性能依据

## P2

### PB-2 镜像大小目标未验证
  **Status: ✅ Implemented in v1**
- 设计：Docker 镜像 <15MB
- 实现：无 Dockerfile 产物验证（Dockerfile 存在但未构建过，见 DEP-1）
- 影响：目标不可证伪

## 一致项 ✅
- 无冲突项——纯缺失
