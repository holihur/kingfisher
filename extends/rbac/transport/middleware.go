package transport

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	"kingfisher/core/errcode"
	"kingfisher/core/jwt"
	"kingfisher/core/response"
	adapter "kingfisher/extends/rbac/adapter/mysql"
	"kingfisher/extends/rbac/app"
)

// NewRoleService creates a RoleService for permission lookups.
func NewRoleService(db *gorm.DB, c cache.Cache) *app.RoleService {
	return app.NewRoleService(adapter.NewRoleRepo(db), c)
}

// RBACMiddlewareWith creates RBAC middleware from an existing RoleService.
func RBACMiddlewareWith(roleSvc *app.RoleService) gin.HandlerFunc {
	return RBACMiddleware(roleSvc)
}

func AuthMiddleware(jwtMgr *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		claims, err := jwtMgr.ParseToken(c.Request.Context(), h[7:])
		if err != nil {
			response.ErrorJSON(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
func RBACMiddleware(roleSvc *app.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		perms, _ := roleSvc.GetUserPermissions(c.Request.Context(), userID)
		ps := make(map[string]bool, len(perms))
		for _, p := range perms {
			ps[p] = true
		}
		c.Set("permissions", ps)
		c.Next()
	}
}
func RBACMiddlewareSvc(db *gorm.DB, c cache.Cache) gin.HandlerFunc {
	roleRepo := adapter.NewRoleRepo(db)
	roleSvc := app.NewRoleService(roleRepo, c)
	return RBACMiddleware(roleSvc)
}

func RequirePerm(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ps, _ := c.Get("permissions")
		psMap, _ := ps.(map[string]bool)
		if psMap == nil {
			psMap = map[string]bool{}
		}
		if !psMap[code] {
			response.Forbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
