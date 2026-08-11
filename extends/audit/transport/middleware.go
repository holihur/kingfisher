package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"kingfisher/extends/audit/app"
	"kingfisher/extends/audit/domain"
)

// methodAction 将 HTTP 方法映射为业务动作（create/update/delete/...）。
var methodAction = map[string]string{
	"POST":   "create",
	"PUT":    "update",
	"PATCH":  "update",
	"DELETE": "delete",
}

// resourceLabels 资源路径段 → 可读中文实体名。
var resourceLabels = map[string]string{
	"users":           "用户",
	"roles":           "角色",
	"permissions":     "权限",
	"menus":           "菜单",
	"configs":         "系统配置",
	"config-groups":   "配置分组",
	"dict-types":      "字典类型",
	"dict-entries":    "字典条目",
	"messages":        "站内信",
	"templates":       "消息模板",
	"scheduled-tasks": "周期任务",
	"audit-logs":      "审计日志",
	"public":          "公开配置",
}

func AuditMiddleware(svc *app.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// 读取并恢复请求体（写操作记录关键字段）
		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		c.Next()

		// 只审计写操作 + 登录/权限相关；GET/HEAD/OPTIONS 跳过
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return
		}
		// 未认证（无 user_id）跳过中间件审计（登录失败等由 handler 单独审计）
		userID := c.GetUint("user_id")
		if userID == 0 {
			return
		}

		status := c.Writer.Status()
		result := "success"
		message := ""
		if status < 200 || status >= 300 {
			result = "failure"
			switch status {
			case 403:
				message = "权限不足"
			case 404:
				message = "资源不存在"
			case 401:
				message = "未认证"
			case 400:
				message = "参数错误"
			default:
				message = fmt.Sprintf("HTTP %d", status)
			}
		}

		resource, resourceID := extractResource(c)
		action := methodAction[method]
		if action == "" {
			action = strings.ToLower(method)
		}
		// 优先使用 handler 提供的变更 Diff（旧值→新值）；否则用请求体详情
		detail := buildDetail(body)
		if diff, ok := c.Get("audit_diff"); ok {
			if s, ok := diff.(string); ok && s != "" {
				detail = s
			}
		}
		svc.Log(c.Request.Context(), &domain.AuditLog{
			UserID: userID, Username: c.GetString("username"),
			Action: action, Resource: resource, ResourceID: resourceID,
			Detail: detail, Result: result, Latency: time.Since(start).Milliseconds(), Message: message,
			IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
		})
	}
}

// buildDetail 从请求体提取关键字段生成 JSON 详情（限制大小，避免记录密码等敏感字段）。
func buildDetail(body []byte) string {
	if len(body) == 0 || len(body) > 8192 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return ""
	}
	// 敏感字段不记录
	for _, k := range []string{"password", "new_password", "old_password", "refresh_token", "access_token", "secret"} {
		delete(m, k)
	}
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}

// extractResource 从 gin.Context 中提取资源名（可读中文）和资源 ID。
func extractResource(c *gin.Context) (string, uint) {
	path := c.FullPath()
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var clean []string
	for _, p := range parts {
		if p == "api" || p == "v1" || p == "" {
			continue
		}
		clean = append(clean, p)
	}

	// 去掉 :param 段，取最后一个非参数段作为资源 key
	resourceKey := "unknown"
	for i := len(clean) - 1; i >= 0; i-- {
		if !strings.HasPrefix(clean[i], ":") {
			resourceKey = clean[i]
			break
		}
	}
	resource := resourceLabels[resourceKey]
	if resource == "" {
		resource = resourceKey // 未映射则用原始路径段
	}

	// 从实际 URL 参数中取资源 ID：优先 entryId，其次 id
	var resourceID uint
	for _, key := range []string{"entryId", "entry_id", "id"} {
		if v := c.Param(key); v != "" {
			if n, err := strconv.ParseUint(v, 10, 64); err == nil {
				resourceID = uint(n)
				break
			}
		}
	}
	return resource, resourceID
}
