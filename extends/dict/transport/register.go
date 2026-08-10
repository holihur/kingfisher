package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/cache"
	adapter "kingfisher/extends/dict/adapter/mysql"
	"kingfisher/extends/dict/app"
	rbacTransport "kingfisher/extends/rbac/transport"
)

// DictModule 字典模块，实现 router.Module 接口
type DictModule struct {
	typeHandler  *DictTypeHandler
	entryHandler *DictEntryHandler
}

func NewDictModule(db *gorm.DB, c cache.Cache) *DictModule {
	typeRepo := adapter.NewDictTypeRepo(db)
	entryRepo := adapter.NewDictEntryRepo(db)
	typeSvc := app.NewDictTypeService(typeRepo, entryRepo)
	entrySvc := app.NewDictEntryService(entryRepo, typeRepo)

	return &DictModule{
		typeHandler:  NewDictTypeHandler(typeSvc),
		entryHandler: NewDictEntryHandler(entrySvc),
	}
}

func (m *DictModule) Name() string                       { return "dict" }
func (m *DictModule) Init(ctx context.Context) error     { return nil }
func (m *DictModule) Shutdown(ctx context.Context) error { return nil }

func (m *DictModule) RegisterPublic(r *gin.RouterGroup) {
	pub := r.Group("/public/dicts")
	pub.GET("/:code/entries", m.entryHandler.GetPublicEntries)
}

func (m *DictModule) RegisterProtected(r *gin.RouterGroup) {
	types := r.Group("/dict-types")
	types.GET("", rbacTransport.RequirePerm("dict:list"), m.typeHandler.List)
	types.GET("/:id", rbacTransport.RequirePerm("dict:list"), m.typeHandler.GetByID)
	types.POST("", rbacTransport.RequirePerm("dict:create"), m.typeHandler.Create)
	types.PUT("/:id", rbacTransport.RequirePerm("dict:update"), m.typeHandler.Update)
	types.POST("/batch-delete", rbacTransport.RequirePerm("dict:delete"), m.typeHandler.BatchDelete)
	types.POST("/batch-status", rbacTransport.RequirePerm("dict:update"), m.typeHandler.BatchUpdateStatus)
	types.DELETE("/:id", rbacTransport.RequirePerm("dict:delete"), m.typeHandler.Delete)

	entries := r.Group("/dict-types/:id/entries")
	entries.GET("", rbacTransport.RequirePerm("dict:list"), m.entryHandler.ListByTypeID)
	entries.POST("", rbacTransport.RequirePerm("dict:create"), m.entryHandler.Create)
	entries.POST("/batch-delete", rbacTransport.RequirePerm("dict:delete"), m.entryHandler.BatchDelete)
	entries.POST("/batch-status", rbacTransport.RequirePerm("dict:update"), m.entryHandler.BatchUpdateStatus)
	entries.PUT("/:entryId", rbacTransport.RequirePerm("dict:update"), m.entryHandler.Update)
	entries.DELETE("/:entryId", rbacTransport.RequirePerm("dict:delete"), m.entryHandler.Delete)
}
