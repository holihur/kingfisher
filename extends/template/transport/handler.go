package transport

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"kingfisher/core/query"
	"kingfisher/core/response"
	"kingfisher/extends/template/app"
)

type TemplateHandler struct{ svc *app.TemplateService }

func NewTemplateHandler(svc *app.TemplateService) *TemplateHandler { return &TemplateHandler{svc: svc} }

// appErrCode 从 error 中提取 errcode，若非 AppError 则返回 -1
func appErrCode(err error) int {
	var e *app.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return -1
}

// @Summary 模版列表
// @Tags Template
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.PageData} "模版列表"
// @Router /api/v1/templates [get]
// templateQueryDefs 模版可查询字段白名单
var templateQueryDefs = query.Defs{
	"name":          {Name: "name", Type: query.TypeString, Searchable: true, Filterable: true},
	"code":          {Name: "code", Type: query.TypeString, Searchable: true, Filterable: true},
	"template_type": {Name: "template_type", Type: query.TypeString, Filterable: true},
	"status":        {Name: "status", Type: query.TypeInt, Filterable: true},
	"remark":        {Name: "remark", Type: query.TypeString, Searchable: true},
	"created_at":    {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

func (h *TemplateHandler) List(c *gin.Context) {
	pq, err := query.Parse(c, templateQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	templates, total, err := h.svc.List(c.Request.Context(), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, templates, total, pq.Page, pq.PageSize)
}

// GetByID 模版详情
// @Summary 模版详情
// @Tags Template
// @Router /api/v1/templates/:id [get]
func (h *TemplateHandler) GetByID(c *gin.Context) {
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

// TemplateReq 模版请求体
type TemplateReq struct {
	Name         string `json:"name" binding:"required"`
	Code         string `json:"code" binding:"required"`
	TemplateType string `json:"template_type"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Status       *int   `json:"status"`
	Remark       string `json:"remark"`
	Version      string `json:"version"`
}

// Create 创建模版
// @Summary 创建模版
// @Tags Template
// @Router /api/v1/templates [post]
func (h *TemplateHandler) Create(c *gin.Context) {
	var req TemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	t, err := h.svc.Create(c.Request.Context(), req.Name, req.Code, defaultStr(req.TemplateType, "general"),
		req.Title, req.Content, reqOr(req.Status, 1), req.Remark, req.Version)
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

// Update 更新模版
// @Summary 更新模版
// @Tags Template
// @Router /api/v1/templates/:id [put]
func (h *TemplateHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req TemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), uint(id), req.Name, req.Code, defaultStr(req.TemplateType, "general"),
		req.Title, req.Content, reqOr(req.Status, 1), req.Remark, req.Version); err != nil {
		if code := appErrCode(err); code > 0 {
			response.ErrorJSON(c, code)
		} else {
			response.InternalError(c)
		}
		return
	}
	response.OKJSON(c, nil)
}

// Delete 删除模版
// @Summary 删除模版
// @Tags Template
// @Router /api/v1/templates/:id [delete]
func (h *TemplateHandler) Delete(c *gin.Context) {
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
// @Summary 批量删除模版
// @Tags Template
// @Router /api/v1/templates/batch-delete [post]
func (h *TemplateHandler) BatchDelete(c *gin.Context) {
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
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Status *int   `json:"status" binding:"required"`
}

// BatchUpdateStatus 批量启用/禁用
// @Summary 批量启用/禁用模版
// @Tags Template
// @Router /api/v1/templates/batch-status [post]
func (h *TemplateHandler) BatchUpdateStatus(c *gin.Context) {
	var req BatchStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.BatchUpdateStatus(c.Request.Context(), req.IDs, *req.Status); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// ---- helpers ----

func defaultStr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func reqOr(v *int, def int) int {
	if v == nil {
		return def
	}
	return *v
}
