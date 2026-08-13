// Package transport implements HTTP transport for the department module.
package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	adapter "kingfisher/extends/department/adapter/mysql"
	"kingfisher/extends/department/app"
	rbacTransport "kingfisher/extends/rbac/transport"
)

type DepartmentModule struct{ handler *DepartmentHandler }

func NewDepartmentModule(db *gorm.DB, c cache.Cache) *DepartmentModule {
	repo := adapter.NewDepartmentRepo(db)
	svc := app.NewDepartmentService(repo, c)
	return &DepartmentModule{handler: NewDepartmentHandler(svc)}
}

func (m *DepartmentModule) Name() string                       { return "department" }
func (m *DepartmentModule) Init(ctx context.Context) error     { return nil }
func (m *DepartmentModule) Shutdown(ctx context.Context) error { return nil }
func (m *DepartmentModule) RegisterPublic(r *gin.RouterGroup)  {}

func (m *DepartmentModule) RegisterProtected(r *gin.RouterGroup) {
	depts := r.Group("/departments")
	depts.GET("/tree", rbacTransport.RequirePerm("department:list"), m.handler.Tree)
	depts.GET("", rbacTransport.RequirePerm("department:list"), m.handler.List)
	depts.GET("/:id", rbacTransport.RequirePerm("department:list"), m.handler.GetByID)
	depts.POST("", rbacTransport.RequirePerm("department:create"), m.handler.Create)
	depts.PUT("/:id", rbacTransport.RequirePerm("department:update"), m.handler.Update)
	depts.PUT("/:id/roles", rbacTransport.RequirePerm("department:update"), m.handler.AssignRoles)
	depts.DELETE("/:id", rbacTransport.RequirePerm("department:delete"), m.handler.Delete)
}
