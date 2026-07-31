# Errcode — 错误码体系

## 职责

统一定义业务错误码，映射到 HTTP 状态码和用户可见消息。

## 错误码分段规则

```
格式: A A B B C  (5 位)
      ┬ ┬ ┬ ┬ ┬
      │ │ └─┴─── 模块内序号 (00-99)
      │ └─────── 模块号 (01-99)
      └───────── 错误大类: 1=客户端 2=服务端
```

| 段 | 范围 | 类别 |
|------|------|------|
| 通用 | 10000-10099 | 参数、认证、服务器 |
| 用户 | 10100-10199 | 登录、注册、资料 |
| 菜单 | 10200-10299 | 菜单管理 |
| 角色 | 10300-10399 | 角色、权限 |
| 配置 | 10400-10499 | 系统配置 |

## 完整错误码表

```go
const (
    // 0 = 成功
    CodeSuccess = 0

    // 通用 10000-10099
    ErrInvalidParam     = 10001  // 参数错误
    ErrUnauthorized     = 10003  // 未认证
    ErrForbidden        = 10004  // 无权限
    ErrNotFound         = 10005  // 资源不存在
    ErrInternal         = 10006  // 服务器内部错误
    ErrTooManyRequest   = 10007  // 请求过于频繁
    ErrMethodNotAllowed = 10008  // 方法不允许
    ErrServiceUnavailable = 10009 // 服务暂时不可用（依赖不可达）

    // 用户模块 10100-10199
    ErrUserExists       = 10101  // 用户已存在
    ErrUserNotFound     = 10102  // 用户不存在
    ErrPasswordWrong    = 10103  // 密码错误
    ErrTokenExpired     = 10104  // Token 过期
    ErrTokenInvalid     = 10105  // Token 无效
    ErrUserDisabled     = 10106  // 用户已禁用
    ErrLoginFailed      = 10107  // 登录失败次数过多（HTTP 429）
    ErrPasswordTooShort = 10108  // 密码过短
    ErrPasswordTooLong  = 10109  // 密码过长
    ErrPasswordWeak     = 10110  // 密码强度不足

    // 菜单模块 10200-10299
    ErrMenuExists       = 10201  // 菜单已存在
    ErrMenuNotFound     = 10202  // 菜单不存在
    ErrMenuHasChildren  = 10203  // 菜单有子节点，不可删除

    // 角色模块 10300-10399
    ErrRoleExists       = 10301  // 角色已存在
    ErrRoleNotFound     = 10302  // 角色不存在
    ErrRoleInUse        = 10303  // 角色被用户使用中

    // 配置模块 10400-10499
    ErrConfigNotFound   = 10401  // 配置不存在
    ErrConfigKeyExists  = 10402  // 配置键已存在
)
```

## 错误码 → HTTP 状态码映射

```go
func HTTPStatus(code int) int {
    switch {
    case code == 0:
        return 200
    case code == 10007 || code == 10107:     // 限流/登录锁定 → 429
        return 429
    case code == 10009:                      // 依赖不可达 → 503
        return 503
    case code == 10003 || code == 10104 || code == 10105:
        return 401
    case code == 10004:
        return 403
    case code == 10005:
        return 404
    case code == 10008:                      // 方法不允许
        return 405
    case code == 10001 || code >= 10100:     // 其他参数/业务错误 → 400（10107 已在前面被 429 捕获）
        return 400
    default:
        return 500
    }
}
```

## 统一响应格式

```go
type Response struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}

func OK(data any) *Response {
    return &Response{Code: 0, Message: "success", Data: data}
}
func Err(code int) *Response {
    return &Response{Code: code, Message: errMsg[code]}
}
func ErrWithMsg(code int, msg string) *Response {
    return &Response{Code: code, Message: msg}
}
```

## 设计要点

- 错误码 5 位数字，第 1 位区分客户端/服务端
- `errMsg` map 存所有 message，Handler 层不硬编码文案
- `Data` 为 `nil` 时不序列化（omitempty）
- 业务错误不暴露内部细节，message 对用户友好
