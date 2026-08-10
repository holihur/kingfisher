package transport

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"

	"kingfisher/core/query"
	"kingfisher/core/response"
	"kingfisher/core/taskqueue"
	"kingfisher/extends/task/app"
)

type ScheduledTaskHandler struct {
	svc         *app.ScheduledTaskService
	producer    taskqueue.Producer
	taskTypesFn func() []taskqueue.TaskTypeInfo
}

func NewScheduledTaskHandler(svc *app.ScheduledTaskService, producer taskqueue.Producer) *ScheduledTaskHandler {
	return &ScheduledTaskHandler{svc: svc, producer: producer}
}

// TaskTypes 可用任务类型列表（各模块 worker 声明），供前端下拉动态加载。
// @Summary 可用任务类型
// @Tags ScheduledTask
// @Router /scheduled-tasks/types [get]
func (h *ScheduledTaskHandler) TaskTypes(c *gin.Context) {
	if h.taskTypesFn == nil {
		response.OKJSON(c, []taskqueue.TaskTypeInfo{})
		return
	}
	response.OKJSON(c, h.taskTypesFn())
}

// appErrCode 从 error 中提取 errcode，若非 AppError 则返回 -1
func appErrCode(err error) int {
	var e *app.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return -1
}

// scheduledTaskQueryDefs 周期任务可查询字段白名单
var scheduledTaskQueryDefs = query.Defs{
	"name":       {Name: "name", Type: query.TypeString, Searchable: true, Filterable: true},
	"task_type":  {Name: "task_type", Type: query.TypeString, Filterable: true},
	"enabled":    {Name: "enabled", Type: query.TypeInt, Filterable: true},
	"remark":     {Name: "remark", Type: query.TypeString, Searchable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

// List 周期任务列表
// @Summary 周期任务列表
// @Tags ScheduledTask
// @Router /scheduled-tasks [get]
func (h *ScheduledTaskHandler) List(c *gin.Context) {
	pq, err := query.Parse(c, scheduledTaskQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	tasks, total, err := h.svc.List(c.Request.Context(), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, tasks, total, pq.Page, pq.PageSize)
}

// GetByID 周期任务详情
// @Summary 周期任务详情
// @Tags ScheduledTask
// @Router /scheduled-tasks/:id [get]
func (h *ScheduledTaskHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	t, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if code := appErrCode(err); code > 0 {
			response.ErrorJSON(c, code)
		} else {
			response.InternalError(c)
		}
		return
	}
	response.OKJSON(c, t)
}

// ScheduledTaskReq 周期任务请求体
type ScheduledTaskReq struct {
	Name     string `json:"name" binding:"required"`
	TaskType string `json:"task_type" binding:"required"`
	CronSpec string `json:"cron_spec" binding:"required"`
	Payload  string `json:"payload"`
	Enabled  *int   `json:"enabled"`
	Remark   string `json:"remark"`
}

// validateCronSpec 校验 cron 表达式（robfig/cron v3 标准 5 段格式）
func validateCronSpec(spec string) bool {
	if _, err := cron.ParseStandard(spec); err != nil {
		return false
	}
	return true
}

// Create 创建周期任务
// @Summary 创建周期任务
// @Tags ScheduledTask
// @Router /scheduled-tasks [post]
func (h *ScheduledTaskHandler) Create(c *gin.Context) {
	var req ScheduledTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if !validateCronSpec(req.CronSpec) {
		response.BadRequest(c, "cron_spec 格式错误（5 段，如 0 9 * * *）")
		return
	}
	t, err := h.svc.Create(c.Request.Context(), req.Name, req.TaskType, req.CronSpec,
		req.Payload, reqOr(req.Enabled, 1), req.Remark)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, t)
}

// Update 更新周期任务
// @Summary 更新周期任务
// @Tags ScheduledTask
// @Router /scheduled-tasks/:id [put]
func (h *ScheduledTaskHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req ScheduledTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if !validateCronSpec(req.CronSpec) {
		response.BadRequest(c, "cron_spec 格式错误（5 段，如 0 9 * * *）")
		return
	}
	if err := h.svc.Update(c.Request.Context(), uint(id), req.Name, req.TaskType, req.CronSpec,
		req.Payload, reqOr(req.Enabled, 1), req.Remark); err != nil {
		if code := appErrCode(err); code > 0 {
			response.ErrorJSON(c, code)
		} else {
			response.InternalError(c)
		}
		return
	}
	response.OKJSON(c, nil)
}

// Delete 删除周期任务
// @Summary 删除周期任务
// @Tags ScheduledTask
// @Router /scheduled-tasks/:id [delete]
func (h *ScheduledTaskHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// batchIDsReq 批量删除请求体
type batchIDsReq struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// BatchDelete 批量删除
// @Summary 批量删除周期任务
// @Tags ScheduledTask
// @Router /scheduled-tasks/batch-delete [post]
func (h *ScheduledTaskHandler) BatchDelete(c *gin.Context) {
	var req batchIDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.BatchDelete(c.Request.Context(), req.IDs); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// BatchStatusReq 批量启用/禁用请求体
type BatchStatusReq struct {
	IDs     []uint `json:"ids" binding:"required,min=1"`
	Enabled *int   `json:"enabled" binding:"required"`
}

// BatchUpdateStatus 批量启用/禁用
// @Summary 批量启用/禁用周期任务
// @Tags ScheduledTask
// @Router /scheduled-tasks/batch-status [post]
func (h *ScheduledTaskHandler) BatchUpdateStatus(c *gin.Context) {
	var req BatchStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.BatchUpdateStatus(c.Request.Context(), req.IDs, *req.Enabled); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// Run 手动执行周期任务：立即入队一次（不受 cron 调度影响），由对应 worker 消费。
// @Summary 手动执行周期任务
// @Tags ScheduledTask
// @Router /scheduled-tasks/:id/run [post]
func (h *ScheduledTaskHandler) Run(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	task, err := h.svc.BuildTask(c.Request.Context(), uint(id))
	if err != nil {
		if code := appErrCode(err); code > 0 {
			response.ErrorJSON(c, code)
		} else {
			response.InternalError(c)
		}
		return
	}
	if _, err := h.producer.Enqueue(c.Request.Context(), task); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, gin.H{"enqueued": true})
}

// ---- helpers ----

func reqOr(v *int, def int) int {
	if v == nil {
		return def
	}
	return *v
}
