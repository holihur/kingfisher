package transport

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"kingfisher/extends/audit/app"
	"kingfisher/extends/audit/domain"
)

func AuditMiddleware(svc *app.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		svc.Log(c.Request.Context(), &domain.AuditLog{
			UserID: userID, Username: c.GetString("username"),
			Action: c.Request.Method, Resource: resource, ResourceID: resourceID,
			IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
		})
	}
}

// extractResource 从 gin.Context 中提取资源名和资源 ID。
// 用 c.Param 获取路由参数值（如 :id、:entryId），优先取最后一个数字参数。
func extractResource(c *gin.Context) (string, uint) {
	// 从 FullPath 提取资源名：/api/v1/dict-types/:id/entries/:entryId → dict-entries
	path := c.FullPath()
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var clean []string
	for _, p := range parts {
		if p == "api" || p == "v1" || p == "" {
			continue
		}
		clean = append(clean, p)
	}

	resource := "unknown"
	if len(clean) > 0 {
		// 去掉 :param 段，取最后一个非参数段作为资源名
		for i := len(clean) - 1; i >= 0; i-- {
			if !strings.HasPrefix(clean[i], ":") {
				resource = clean[i]
				break
			}
		}
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
