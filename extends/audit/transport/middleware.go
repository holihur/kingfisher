package transport

import (
	"github.com/gin-gonic/gin"
	"kingfisher/extends/audit/domain"
	"kingfisher/extends/audit/app"
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
		resource := extractResource(c.FullPath())
		userID := c.GetUint("user_id")
		if userID == 0 { return }
		svc.Log(c.Request.Context(), &domain.AuditLog{
			UserID: userID, Username: c.GetString("username"),
			Action: c.Request.Method, Resource: resource,
			IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
		})
	}
}

func extractResource(path string) string {
	parts := []string{}
	for _, p := range splitPath(path) {
		if p != "" && p != "api" && p != "v1" { parts = append(parts, p) }
	}
	if len(parts) > 0 { return parts[0] }
	return "unknown"
}

func splitPath(p string) []string {
	var result []string
	start := 1
	for i := 1; i <= len(p); i++ {
		if i == len(p) || p[i] == '/' {
			result = append(result, p[start:i])
			start = i + 1
		}
	}
	return result
}
