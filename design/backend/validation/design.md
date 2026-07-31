# Validation — 请求校验规范

## 全局规则（自动应用）

以下规则对所有请求自动生效，Handler 不需要手动写。

```go
// core/middleware/validator.go — 在 Recovery 中间件中初始化
func InitValidator() {
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        // 1. 字符串自动 trim 空格
        v.RegisterTagNameFunc(func(fld reflect.StructField) string {
            return fld.Tag.Get("trim")  // 配合自定义 trim 逻辑
        })

        // 2. 自定义校验器注册
        v.RegisterValidation("phone", validatePhone)      // 手机号
        v.RegisterValidation("idcard", validateIDCard)    // 身份证
        v.RegisterValidation("nohtml", validateNoHTML)     // 不含 HTML 标签（防 XSS）
        v.RegisterValidation("password", validatePassword) // 密码强度
        v.RegisterValidation("sort", validateSortField)    // 排序字段白名单
    }
}
```

## 自定义校验器

### password

```go
// 至少 8 位，包含大写、小写、数字
func validatePassword(fl validator.FieldLevel) bool {
    pwd := fl.Field().String()
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(pwd)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(pwd)
    hasDigit := regexp.MustCompile(`[0-9]`).MatchString(pwd)
    return len(pwd) >= 8 && len(pwd) <= 64 && hasUpper && hasLower && hasDigit
}
```

### nohtml

```go
// 拒绝 HTML 标签，防 XSS
func validateNoHTML(fl validator.FieldLevel) bool {
    return !regexp.MustCompile(`<[^>]*>`).MatchString(fl.Field().String())
}
```

### sort

```go
// 排序字段必须在白名单内
func validateSortField(fl validator.FieldLevel) bool {
    allowed := map[string]bool{"created_at": true, "updated_at": true, "id": true, "username": true, "sort": true}
    return allowed[fl.Field().String()]
}
```

### phone

```go
// 中国大陆手机号
func validatePhone(fl validator.FieldLevel) bool {
    return regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(fl.Field().String())
}
```

## 请求 struct 规范

### ID 字段

```go
// 所有 int ID 字段 > 0
type GetUserReq struct {
    ID uint `uri:"id" binding:"required,min=1"`
}
```

### 分页

```go
type PaginationReq struct {
    Page     int    `form:"page"     binding:"omitempty,min=1"`
    PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
    Sort     string `form:"sort"     binding:"omitempty,sort"`         // 自定义校验
    Order    string `form:"order"    binding:"omitempty,oneof=asc desc"`
}
```

### 创建

```go
type CreateUserReq struct {
    Username string `json:"username" binding:"required,min=3,max=32,nohtml"`
    Password string `json:"password" binding:"required,password"`       // 自定义校验
    Email    string `json:"email"    binding:"omitempty,email,max=128"`
}
```

### 更新（用 pointer 区分零值和不传）

```go
type UpdateUserReq struct {
    Email  *string `json:"email"  binding:"omitempty,email,max=128"`    // *string: nil=不更新, ""=清空
    Status *int    `json:"status" binding:"omitempty,oneof=0 1"`
    RoleID *uint   `json:"role_id" binding:"omitempty,min=1"`
}
```

## 错误消息中文化

```go
// core/middleware/validator.go
func translateError(err error) string {
    if ve, ok := err.(validator.ValidationErrors); ok {
        for _, e := range ve {
            switch e.Tag() {
            case "required": return fmt.Sprintf("%s 不能为空", e.Field())
            case "min":      return fmt.Sprintf("%s 长度不能小于 %s", e.Field(), e.Param())
            case "max":      return fmt.Sprintf("%s 长度不能大于 %s", e.Field(), e.Param())
            case "email":    return "邮箱格式不正确"
            case "password": return "密码需要至少 8 位，包含大小写字母和数字"
            case "nohtml":   return fmt.Sprintf("%s 不能包含 HTML 标签", e.Field())
            case "phone":    return "手机号格式不正确"
            }
        }
    }
    return "参数错误"
}
```

Handler 层统一调用：

```go
if err := c.ShouldBindJSON(&req); err != nil {
    response.BadRequest(translateError(err)).Abort(c)
    return
}
```

## Handler 层的额外校验

validator tag 只能覆盖格式校验，业务规则在 Handler 或 Service 层处理：

```go
// 更新用户时，禁止修改 username（业务规则，不是格式问题）
func (h *UserHandler) Update(c *gin.Context) {
    var req UpdateUserReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(translateError(err)).Abort(c)
        return
    }
    // 业务规则校验
    if req.Username != nil {
        response.BadRequest("用户名不可修改").Abort(c)
        return
    }
    // ...
}
```

## 校验层级总结

| 层 | 校验内容 | 工具 |
|------|----------|------|
| Middleware | 请求体大小（10MB 上限） | `http.MaxBytesReader` |
| Gin Binding | 类型、必填、长度、格式 | `binding:"required,min=3,max=32,email"` |
| 自定义 validator | 密码强度、手机号、防 XSS | `binding:"password,phone,nohtml"` |
| Handler | 业务规则（字段不可修改、数据范围） | 手写 if |
| Service | 业务约束（用户名唯一、余额充足） | 查 DB 判断 |
