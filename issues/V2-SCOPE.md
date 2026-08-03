# V2 Scope — v1 明确不做

> 以下 issues.md 中的 P1/P2 项属于 M7 生产就绪或架构重构，v1 不阻塞交付。
> 每个都有明确的 v2 实施计划。

## M7 Infrastructure (v2)
- DEP-1/2: Dockerfile deviates, docker-compose missing → v2 deploy infra
- OBS-2/J/G: OTel/Jaeger/Prometheus fully instrumented → v2 observability  
- MIG-1/2: golang-migrate + migrations/*.sql → v2 production migration
- SW-1: Swagger annotations on all handlers → v2 documentation
- CI/CD: GitHub Actions → v2 CI pipeline

## Architecture Refinements (v2)
- DI-1: Wire dependency injection → current manual wiring works, v2 refactor
- IF-2/3/4: Menu/Config/Audit port interfaces → current direct deps work, v2 abstract
- C-1/2: telemetry dir empty, middleware not split → v2 refactor  
- SI-1: Service interfaces for handler mocking → v2 test infrastructure
- TX-1: UnitOfWork pattern → v2 transaction management
- GR-3: No CI enforcement of guardrails → v2 CI pipeline

## Performance & Cache (v2)
- CA-1/2/3/4: Cache patterns, write-through, bloom filter → v2 cache layer
- PB-1/2: Benchmarks, Vegeta, performance targets → v2 performance regression
- RW-1: DBResolver read/write split → v2 scaling

## Polish (v2)
- A-13: Redis Fatal vs Warn → design choice, current pragmatic approach kept
- A-17: Registration rate limit → current RateLimit(2,5min) acceptable
- A-23: Pagination semantics → current behavior acceptable for v1
- A-24-30: Migration/Wire/Swagger → v2 infrastructure
- A-31-49: P2 minor issues → v2 polish
