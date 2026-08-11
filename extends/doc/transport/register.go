// Package transport implements HTTP transport for the doc module.
package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	adapter "kingfisher/extends/doc/adapter/mysql"
	"kingfisher/extends/doc/app"
	rbacTransport "kingfisher/extends/rbac/transport"
)

type DocModule struct{ handler *DocHandler }

func NewDocModule(db *gorm.DB, c cache.Cache) *DocModule {
	repo := adapter.NewDocRepo(db)
	svc := app.NewDocService(repo, c)
	return &DocModule{handler: NewDocHandler(svc)}
}

func (m *DocModule) Name() string                       { return "doc" }
func (m *DocModule) Init(ctx context.Context) error     { return nil }
func (m *DocModule) Shutdown(ctx context.Context) error { return nil }
func (m *DocModule) RegisterPublic(r *gin.RouterGroup)  {}

func (m *DocModule) RegisterProtected(r *gin.RouterGroup) {
	docs := r.Group("/docs")
	// 目录
	docs.GET("/tree", rbacTransport.RequirePerm("doc:list"), m.handler.GetTree)
	docs.POST("/dirs", rbacTransport.RequirePerm("doc:create"), m.handler.CreateDir)
	docs.PUT("/dirs/:id", rbacTransport.RequirePerm("doc:update"), m.handler.UpdateDir)
	docs.DELETE("/dirs/:id", rbacTransport.RequirePerm("doc:delete"), m.handler.DeleteDir)
	docs.GET("/dirs/:id/roles", rbacTransport.RequirePerm("doc:list"), m.handler.GetDirRoles)
	docs.PUT("/dirs/:id/roles", rbacTransport.RequirePerm("doc:update"), m.handler.SetDirRoles)
	// 文档
	docs.GET("", rbacTransport.RequirePerm("doc:list"), m.handler.ListDocs)
	docs.POST("", rbacTransport.RequirePerm("doc:create"), m.handler.CreateDoc)
	docs.GET("/:id", rbacTransport.RequirePerm("doc:list"), m.handler.GetDoc)
	docs.PUT("/:id", rbacTransport.RequirePerm("doc:update"), m.handler.UpdateDoc)
	docs.PUT("/:id/publish", rbacTransport.RequirePerm("doc:update"), m.handler.Publish)
	docs.PUT("/:id/unpublish", rbacTransport.RequirePerm("doc:update"), m.handler.Unpublish)
	docs.GET("/:id/versions", rbacTransport.RequirePerm("doc:list"), m.handler.ListVersions)
	docs.GET("/:id/versions/:no", rbacTransport.RequirePerm("doc:list"), m.handler.GetVersion)
	docs.POST("/:id/restore", rbacTransport.RequirePerm("doc:update"), m.handler.Restore)
	docs.DELETE("/:id", rbacTransport.RequirePerm("doc:delete"), m.handler.DeleteDoc)
}
