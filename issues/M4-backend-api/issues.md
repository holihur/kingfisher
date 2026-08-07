# M4 后端 API 全量就绪 — 设计与实现差异

> 来源：`design/M4-backend-api/design.md` 对照实现
> 排查日期：2026-08-03 ｜ 详见各模块 issue

## P0

### ✅ M4-1 migrations/ 空目录（MIG-1）
- 期望：10 组 SQL 迁移文件
- 现状：无任何 .sql

### ✅ M4-2 Wire 注入未实现（DI-1）
### ✅ M4-3 Swagger 注解缺失（SW-1）
### ✅ M4-4 审计日志零写入（EA-1）

## P1

### ✅ M4-5 事务 UnitOfWork 未实现（TX-1）
### ✅ M4-6 缓存模式未实现（CA-1/CA-2）
### ✅ M4-7 Service 接口化未实现（SI-1）
### ✅ M4-8 `POST /users` 契约与实现均缺失（AC-1/EU-2）

## 结论
- ✅ 已达标：菜单树/配置 CRUD/审计查询接口可调通（SQLite）
- ❌ M4 依赖的迁移/DI/Swagger/审计写入四大件全部缺失，验收失败
