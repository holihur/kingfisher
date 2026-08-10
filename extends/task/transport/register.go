package transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kingfisher/core/taskqueue"
	rbacTransport "kingfisher/extends/rbac/transport"
	adapter "kingfisher/extends/task/adapter/mysql"
	"kingfisher/extends/task/app"
)

// TaskModule 周期任务模块，实现 router.Module。
// 提供周期任务 CRUD（后台管理），并暴露 PeriodicConfigProvider 给主程序
// 注册到 asynq PeriodicTaskManager（原始数据从 DB 拉取，周期性同步调度）。
type TaskModule struct {
	handler  *ScheduledTaskHandler
	provider *app.PeriodicConfigProvider
}

func NewTaskModule(db *gorm.DB, producer taskqueue.Producer) *TaskModule {
	repo := adapter.NewScheduledTaskRepo(db)
	svc := app.NewScheduledTaskService(repo)
	return &TaskModule{
		handler:  NewScheduledTaskHandler(svc, producer),
		provider: app.NewPeriodicConfigProvider(svc),
	}
}

// PeriodicProvider 注册模式：主程序通过该可选接口收集本模块的周期任务 provider。
func (m *TaskModule) PeriodicProvider() taskqueue.PeriodicProvider { return m.provider }

// InjectTaskTypes 注入"可用任务类型"收集函数（main 在收集完所有 worker 后调用）。
// 任务管理页通过 /scheduled-tasks/types 动态获取各模块 worker 声明的任务类型。
func (m *TaskModule) InjectTaskTypes(fn func() []taskqueue.TaskTypeInfo) {
	m.handler.taskTypesFn = fn
}

func (m *TaskModule) Name() string                       { return "task" }
func (m *TaskModule) Init(ctx context.Context) error     { return nil }
func (m *TaskModule) Shutdown(ctx context.Context) error { return nil }
func (m *TaskModule) RegisterPublic(r *gin.RouterGroup)  {}

func (m *TaskModule) RegisterProtected(r *gin.RouterGroup) {
	ts := r.Group("/scheduled-tasks")
	ts.GET("", rbacTransport.RequirePerm("task:list"), m.handler.List)
	ts.GET("/types", rbacTransport.RequirePerm("task:list"), m.handler.TaskTypes)
	ts.GET("/:id", rbacTransport.RequirePerm("task:list"), m.handler.GetByID)
	ts.POST("", rbacTransport.RequirePerm("task:create"), m.handler.Create)
	ts.PUT("/:id", rbacTransport.RequirePerm("task:update"), m.handler.Update)
	ts.POST("/:id/run", rbacTransport.RequirePerm("task:update"), m.handler.Run)
	ts.POST("/batch-delete", rbacTransport.RequirePerm("task:delete"), m.handler.BatchDelete)
	ts.POST("/batch-status", rbacTransport.RequirePerm("task:update"), m.handler.BatchUpdateStatus)
	ts.DELETE("/:id", rbacTransport.RequirePerm("task:delete"), m.handler.Delete)
}
