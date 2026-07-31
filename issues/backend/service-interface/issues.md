# Service Interface — 设计与实现差异

> 来源：`design/backend/service-interface/design.md` 对照 `extends/*/transport/handler.go`
> 排查日期：2026-07-31

## P1

### SI-1 Handler 依赖具体 Service
- 设计：`extends/{module}/port/service.go` 定义 AuthService/UserService/MenuService/RoleService 接口，Handler 依赖接口以便 mock
- 实现：全部 Handler 持有具体 struct——`*app.AuthService`、`*app.UserService`、`*app.RoleService`、`*app.MenuService`、`*app.ConfigService`、`*app.AuditService`
- 影响：Handler 单测需实例化整个依赖链（Service→Repo→DB），无隔离性；与设计「Service 接口化」目标完全未达成

## P1

### SI-2 模块间调用走具体实现而非 port
- 设计：extends 之间跨模块调用需走 port 接口
- 实现：`extends/menu`、`extends/config`、`extends/audit` 内部直接 import `adapter/mysql` 具体仓库
- 影响：模块无法独立发布/替换实现

## 一致项 ✅
- 问题描述部分（Handler 依赖具体 Service）与设计文档「问题」章节完全吻合——设计已预见到现状，但方案未实施
