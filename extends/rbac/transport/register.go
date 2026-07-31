package transport
import ("context"; "github.com/gin-gonic/gin"; "gorm.io/gorm"; "kingfisher/core/cache"; "kingfisher/extends/rbac/adapter/mysql"; "kingfisher/extends/rbac/app")
type RBACModule struct{ roleHandler *RoleHandler; permHandler *PermHandler }
func NewRBACModule(db *gorm.DB, c cache.Cache) *RBACModule {
    roleRepo := adapter.NewRoleRepo(db); permRepo := adapter.NewPermRepo(db)
    roleSvc := app.NewRoleService(roleRepo, c); permSvc := app.NewPermService(permRepo)
    return &RBACModule{roleHandler: NewRoleHandler(roleSvc), permHandler: NewPermHandler(permSvc)}
}
func (m *RBACModule) Name() string { return "rbac" }
func (m *RBACModule) Init(ctx context.Context) error { return nil }
func (m *RBACModule) Shutdown(ctx context.Context) error { return nil }
func (m *RBACModule) RegisterPublic(r *gin.RouterGroup) {}
func (m *RBACModule) RegisterProtected(r *gin.RouterGroup) {
    roles := r.Group("/roles"); roles.GET("", m.roleHandler.List); roles.GET("/:id", m.roleHandler.GetByID)
    roles.POST("", RequirePerm("role:create"), m.roleHandler.Create)
    roles.PUT("/:id", RequirePerm("role:update"), m.roleHandler.Update)
    roles.DELETE("/:id", RequirePerm("role:delete"), m.roleHandler.Delete)
    roles.GET("/:id/permissions", m.roleHandler.GetPermissions)
    roles.PUT("/:id/permissions", RequirePerm("role:update"), m.roleHandler.AssignPerms)
    roles.GET("/:id/menus", m.roleHandler.GetMenus)
    roles.PUT("/:id/menus", RequirePerm("role:update"), m.roleHandler.AssignMenus)
    perms := r.Group("/permissions"); perms.GET("", m.permHandler.List)
}
