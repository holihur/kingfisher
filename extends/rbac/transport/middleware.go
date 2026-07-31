package transport
import ("github.com/gin-gonic/gin"; "kingfisher/core/errcode"; "kingfisher/core/jwt"; "kingfisher/core/response"; "kingfisher/extends/rbac/app"; "strings")
func AuthMiddleware(jwtMgr *jwt.JWTManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        h := c.GetHeader("Authorization")
        if h == "" || !strings.HasPrefix(h, "Bearer ") { response.Unauthorized(c); c.Abort(); return }
        claims, err := jwtMgr.ParseToken(c.Request.Context(), h[7:])
        if err != nil { response.ErrorJSON(c, errcode.ErrUnauthorized); c.Abort(); return }
        c.Set("user_id", claims.UserID); c.Set("role", claims.Role); c.Next()
    }
}
func RBACMiddleware(roleSvc *app.RoleService) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("user_id"); perms, _ := roleSvc.GetUserPermissions(c.Request.Context(), userID)
        ps := make(map[string]bool, len(perms)); for _, p := range perms { ps[p] = true }; c.Set("permissions", ps); c.Next()
    }
}
func RequirePerm(code string) gin.HandlerFunc {
    return func(c *gin.Context) {
        ps, _ := c.Get("permissions"); psMap, _ := ps.(map[string]bool); if psMap == nil { psMap = map[string]bool{} }
        if !psMap[code] { response.Forbidden(c); c.Abort(); return }; c.Next()
    }
}
