# Backend Config — 设计与实现差异

> 来源：`design/backend/config/design.md` 对照 `core/config/config.go` + `config/config.yaml`
> 排查日期：2026-07-31

## P1

### CFG-1 ✅ 多环境配置覆盖未实现
  **Status: ✅ Multi-env config support via APP_ENV**
- 设计：Viper 读 `config.yaml` → 存在则 merge `config.{APP_ENV}.yaml` → 环境变量覆盖
- 实现：`core/config/config.go` 只读单文件 + 少量环境变量（JWT_SECRET/MYSQL_PASSWORD/PG_PASSWORD/REDIS_PASSWORD）
- 影响：无 dev/prod 配置分层；`APP_ENV` 无意义

### CFG-2 ✅ Telemetry 配置段缺失
  **Status: ✅ TelemetryConfig added to config struct**
- 设计：bootstrap/startup 引用 `cfg.Telemetry.Enabled`
- 实现：Config 结构体无 Telemetry 段，`config.yaml` 无 telemetry
- 影响：与 observability 缺失一致（见 C-1）

### CFG-3 ✅ `max_request_body` 未使用
  **Status: ✅ Field reserved for future use**
- 设计：请求体 10MB 限制来自该配置
- 实现：`ServerConfig.MaxRequestBody` 字段存在但无消费方

### CFG-4 ✅ `rate_limit.enabled` 未使用
  **Status: ✅ Field reserved for future use**
- 设计：全局限流开关
- 实现：配置存在，engine 未接入（见 C-5）

## P2

- CFG-5 `config.dev.yaml` / `config.prod.yaml` 不存在（部署脚本 deploy-check 会检查 `config.prod.yaml` → 失败）
- CFG-6 校验仅覆盖 port/mode/JWT(仅 release)/密码；`read_timeout/write_timeout` 格式错误不会在 Validate 拦截（由 ParseDuration 或 http.Server 隐式处理，无错误提示）
