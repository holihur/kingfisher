package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"

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
	"users":            "用户",
	"roles":            "角色",
	"permissions":      "权限",
	"menus":            "菜单",
	"configs":          "系统配置",
	"config-groups":    "配置分组",
	"dict-types":       "字典类型",
	"dict-entries":     "字典条目",
	"messages":         "站内信",
	"templates":        "消息模板",
	"scheduled-tasks":  "周期任务",
	"audit-logs":       "审计日志",
	"public":           "公开配置",
}

func AuditMiddleware(svc *app.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 读取并恢复请求体（写操作记录关键字段）
		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		c.Next()
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			return
		}
		if c.Writer.Status() < 200 || c.Writer.Status() >= 300 {
			return
		}
		resource, resourceID := extractResource(c)
		userID := c.GetUint("user_id")
		if userID == 0 {
			return
		}
		action := methodAction[c.Request.Method]
		if action == "" {
			action = strings.ToLower(c.Request.Method)
		}
		detail := buildDetail(body)
		svc.Log(c.Request.Context(), &domain.AuditLog{
			UserID: userID, Username: c.GetString("username"),
			Action: action, Resource: resource, ResourceID: resourceID,
			Detail: detail, IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
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
