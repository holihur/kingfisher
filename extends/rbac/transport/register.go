// Package rbac implements rbac logic.

package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	adapter "kingfisher/extends/rbac/adapter/mysql"
	"kingfisher/extends/rbac/app"
)

type RBACModule struct {
	roleHandler *RoleHandler
	permHandler *PermHandler
}

func NewRBACModule(db *gorm.DB, c cache.Cache) *RBACModule {
	roleRepo := adapter.NewRoleRepo(db)
	permRepo := adapter.NewPermRepo(db)
	roleSvc := app.NewRoleService(roleRepo, c)
	permSvc := app.NewPermService(permRepo)
	return &RBACModule{roleHandler: NewRoleHandler(roleSvc), permHandler: NewPermHandler(permSvc)}
}
func (m *RBACModule) Name() string                       { return "rbac" }
func (m *RBACModule) Init(ctx context.Context) error     { return nil }
func (m *RBACModule) Shutdown(ctx context.Context) error { return nil }
func (m *RBACModule) RegisterPublic(r *gin.RouterGroup)  {}
func (m *RBACModule) RegisterProtected(r *gin.RouterGroup) {
	roles := r.Group("/roles")
	roles.GET("", RequirePerm("role:list"), m.roleHandler.List)
	roles.GET("/:id", RequirePerm("role:list"), m.roleHandler.GetByID)
	roles.POST("", RequirePerm("role:create"), m.roleHandler.Create)
	roles.PUT("/:id", RequirePerm("role:update"), m.roleHandler.Update)
	roles.DELETE("/:id", RequirePerm("role:delete"), m.roleHandler.Delete)
	roles.POST("/batch-delete", RequirePerm("role:delete"), m.roleHandler.BatchDelete)
	roles.POST("/batch-status", RequirePerm("role:update"), m.roleHandler.BatchUpdateStatus)
	roles.GET("/:id/permissions", RequirePerm("role:list"), m.roleHandler.GetPermissions)
	roles.PUT("/:id/permissions", RequirePerm("role:update"), m.roleHandler.AssignPerms)
	roles.GET("/:id/menus", RequirePerm("role:list"), m.roleHandler.GetMenus)
	roles.PUT("/:id/menus", RequirePerm("role:update"), m.roleHandler.AssignMenus)
	roles.GET("/:id/data-scope", RequirePerm("role:list"), m.roleHandler.GetDataScope)
	roles.PUT("/:id/data-scope", RequirePerm("role:update"), m.roleHandler.SetDataScope)
	perms := r.Group("/permissions")
	perms.GET("", RequirePerm("role:list"), m.permHandler.List)
}
