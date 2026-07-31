# Logger — 日志系统

## 职责

基于 Zap 的结构化日志封装，提供统一字段规范，支持 JSON/Console 输出、日志滚动、敏感信息脱敏。

## 对外接口

```go
func New(cfg LogConfig) (*zap.Logger, error)
func WithContext(ctx context.Context) *zap.Logger   // 从 ctx 提取 trace_id 等
```

## 日志级别

| Level | 使用场景 |
|------|----------|
| Debug | 开发调试，生产默认关闭 |
| Info | 请求日志、关键操作（登录、注册、数据变更） |
| Warn | 可恢复异常（缓存 miss、限流触发） |
| Error | 需要关注的错误（DB 错误、第三方超时） |
| Fatal | 启动失败，进程退出 |

## 字段规范

```go
// 通用字段（所有日志都应携带）
zap.String("trace_id", traceID)     // 链路追踪 ID
zap.String("span_id", spanID)       // Span ID
zap.String("request_id", reqID)     // HTTP 请求 ID
zap.Int64("user_id", userID)        // 当前用户（登录接口为空）

// 错误日志
zap.Error(err)                      // error 对象
zap.String("stack", stacktrace)     // 堆栈（仅 Error 级别）

// HTTP 请求日志（middleware/logger 自动记）
zap.String("method", c.Request.Method)
zap.String("path", c.Request.URL.Path)
zap.Int("status", c.Writer.Status())
zap.Duration("latency", latency)
zap.String("ip", c.ClientIP())
zap.String("user_agent", c.Request.UserAgent())
```

## 敏感数据脱敏

```go
// 自定义 Zap Core 实现脱敏
type MaskCore struct { zapcore.Core }

func (c *MaskCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
    for i, f := range fields {
        if f.Key == "password" || f.Key == "token" || f.Key == "secret" {
            fields[i] = zap.String(f.Key, "***")
        }
    }
    return c.Core.Write(entry, fields)
}
```

## 日志滚动（lumberjack）

```yaml
log:
  format: json
  output: file
  file_path: logs/app.log
  max_size: 100     # MB，超限自动切割
  max_backups: 10   # 保留 10 个旧文件
  max_age: 30       # 30 天后删除
  compress: true    # gzip 压缩旧文件
```

## 使用示例

```go
// Service 层
logger := logger.WithContext(ctx)
logger.Info("user login success",
    zap.Uint("user_id", user.ID),
    zap.String("username", user.Username),
)

// 错误
logger.Error("failed to create user",
    zap.Error(err),
    zap.String("username", req.Username),
)
```

## 设计要点

- 生产环境用 JSON 格式，方便 ELK/Loki 采集
- 每个请求的日志通过 `request_id` 串联
- 密码、token、secret 写入日志时自动替换为 `***`
