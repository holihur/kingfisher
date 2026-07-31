# Swagger Annotation Checklist — 每个 Handler 必须的注解

## 最小注解集

每个公开的 Handler 方法必须带以下注解，缺一不可。缺少任一项会导致生成的 TS 类型不完整或 API 文档不可用。

## 检查清单

```
✅ @Summary         — 一句话描述
✅ @Tags            — 所属分组（用于 Swagger UI 分组折叠）
✅ @Accept          — json / multipart/form-data
✅ @Produce         — json
✅ @Param           — 每个参数一个，包含：来源 名称 类型 是否必填 描述
✅ @Success         — 200 响应，指明 {object} 和 data 的 model
✅ @Failure         — 至少 400 和 500
✅ @Router          — method + path
✅ @Security        — (如需要认证) BearerAuth
```

## 完整模板

### GET 列表（带分页 + 搜索）

```go
// @Summary 获取用户列表
// @Description 分页查询用户列表，支持关键词搜索
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1) minimum(1)
// @Param page_size query int false "每页条数" default(20) minimum(1) maximum(100)
// @Param keyword query string false "搜索关键词（用户名/邮箱）"
// @Param sort query string false "排序字段" Enums(created_at, username)
// @Param order query string false "排序方向" Enums(asc, desc) default(desc)
// @Success 200 {object} response.Response{data=PaginatedUser} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未认证"
// @Failure 403 {object} response.Response "无权限"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context)
```

### GET 单条

```go
// @Summary 获取用户详情
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "用户 ID" minimum(1)
// @Success 200 {object} response.Response{data=domain.User} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context)
```

### POST 创建

```go
// @Summary 创建用户
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateUserReq true "用户信息"
// @Success 200 {object} response.Response{data=domain.User} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 409 {object} response.Response "用户名已存在"
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context)
```

### PUT 更新

```go
// @Summary 更新用户
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Param body body UpdateUserReq true "更新内容"
// @Success 200 {object} response.Response{data=domain.User} "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context)
```

### DELETE

```go
// @Summary 删除用户
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 404 {object} response.Response "用户不存在"
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context)
```

### 文件上传

```go
// @Summary 上传头像
// @Tags User
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Success 200 {object} response.Response{data=UploadResp} "上传成功"
// @Failure 400 {object} response.Response "文件格式不支持"
// @Router /api/v1/users/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context)
```

## 必须定义 Swagger 可引用的类型

在 `main.go` 或 handler 文件中定义响应 data 的匿名 struct 为命名类型：

```go
// 在 handler 文件顶部
type PaginatedUser struct {
    Items    []domain.User `json:"items"`
    Total    int64         `json:"total"`
    Page     int           `json:"page"`
    PageSize int           `json:"page_size"`
}
```

Gin 模式的特殊处理——因为使用 `response.Response{data=...}` 无法直接描述泛型：

```go
// @Success 200 {object} response.Response{data=PaginatedUser}
```

`swaggo` 能正确解析此格式并展开 data 字段。

## 主文件中的通用注解

```go
// cmd/server/main.go

// @title           Kingfisher Admin API
// @version         1.0
// @description     后台管理系统 API 文档
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @schemes http https
func main() { ... }
```

## CI 自动检查

```bash
#!/bin/bash
# scripts/check-swagger.sh
# 1. 生成 swagger
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# 2. 检查每个 routes.go 文件中的 handler 是否有 @Summary
MISSING=$(grep -rL '@Summary' extends/*/transport/ | grep -v '_test.go' || true)
if [ -n "$MISSING" ]; then
    echo "ERROR: 以下 handler 文件缺少 Swagger 注解:"
    echo "$MISSING"
    exit 1
fi

# 3. 检查 @Success 是否都有 {object} 类型
MISSING_TYPE=$(grep -r '@Success' extends/*/transport/ | grep -v '{object}' || true)
if [ -n "$MISSING_TYPE" ]; then
    echo "WARNING: 以下 @Success 未指定返回类型:"
    echo "$MISSING_TYPE"
fi
```

## 常见遗漏 & 后果

| 遗漏 | 后果 |
|------|------|
| 缺少 `@Success {object}` | TS 类型生成为 `unknown` |
| 缺少 `@Param type` | Swagger UI 无法渲染输入框 |
| 缺少 `@Failure` | 前端不知道可能的错误码 |
| 缺少 `Enums()` | 前端只能用 `string` 而非字面量联合类型 |
| 缺少 `minimum/maximum` | 前端无法做 range 校验 |
| `@Router` 路径含 Go template `:id` 而非 `{id}` | Swagger 路径参数无法识别——必须用 `{id}` 格式 |
