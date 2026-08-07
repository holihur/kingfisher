package transport

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"kingfisher/core/errcode"
	"kingfisher/core/query"
	"kingfisher/core/response"
	"kingfisher/extends/dict/app"
)

// appErrCode 从 error 中提取 errcode，若非 AppError 则返回 -1
func appErrCode(err error) int {
	if e, ok := err.(*app.Error); ok {
		return e.Code
	}
	return -1
}

// ---- DictTypeHandler ----

type DictTypeHandler struct{ svc *app.DictTypeService }

func NewDictTypeHandler(svc *app.DictTypeService) *DictTypeHandler { return &DictTypeHandler{svc: svc} }

// @Summary 字典类型列表
// @Tags Dict
// @Router /api/v1/dict-types [get]
// dictTypeQueryDefs 字典类型可查询字段白名单
var dictTypeQueryDefs = query.Defs{
	"code":       {Name: "code", Type: query.TypeString, Searchable: true, Filterable: true},
	"name":       {Name: "name", Type: query.TypeString, Searchable: true, Filterable: true},
	"remark":     {Name: "remark", Type: query.TypeString, Searchable: true},
	"is_public":  {Name: "is_public", Type: query.TypeBool, Filterable: true},
	"status":     {Name: "status", Type: query.TypeInt, Filterable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

func (h *DictTypeHandler) List(c *gin.Context) {
	pq, err := query.Parse(c, dictTypeQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	types, total, err := h.svc.List(c.Request.Context(), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, types, total, pq.Page, pq.PageSize)
}

// @Summary 字典类型详情
// @Tags Dict
// @Router /api/v1/dict-types/:id [get]
func (h *DictTypeHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	t, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.ErrorJSON(c, errcode.ErrDictTypeNotFound)
		return
	}
	response.OKJSON(c, t)
}

// DictTypeReq 字典类型请求体
type DictTypeReq struct {
	Code     string `json:"code" binding:"required"`
	Name     string `json:"name" binding:"required"`
	IsPublic *bool  `json:"is_public"`
	Status   *int   `json:"status"`
	Remark   string `json:"remark"`
}

// @Summary 创建字典类型
// @Tags Dict
// @Router /api/v1/dict-types [post]
func (h *DictTypeHandler) Create(c *gin.Context) {
	var req DictTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	isPublic := false
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	t, err := h.svc.Create(c.Request.Context(), req.Code, req.Name, isPublic, status, req.Remark)
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

// @Summary 更新字典类型
// @Tags Dict
// @Router /api/v1/dict-types/:id [put]
func (h *DictTypeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req DictTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	isPublic := false
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	if err := h.svc.Update(c.Request.Context(), uint(id), req.Code, req.Name, isPublic, status, req.Remark); err != nil {
		if code := appErrCode(err); code > 0 {
			response.ErrorJSON(c, code)
		} else {
			response.InternalError(c)
		}
		return
	}
	response.OKJSON(c, nil)
}

// @Summary 删除字典类型
// @Tags Dict
// @Router /api/v1/dict-types/:id [delete]
func (h *DictTypeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		if code := appErrCode(err); code > 0 {
			response.ErrorJSON(c, code)
		} else {
			response.InternalError(c)
		}
		return
	}
	response.OKJSON(c, nil)
}

// ---- DictEntryHandler ----

type DictEntryHandler struct{ svc *app.DictEntryService }

func NewDictEntryHandler(svc *app.DictEntryService) *DictEntryHandler {
	return &DictEntryHandler{svc: svc}
}

// @Summary 字典条目列表（按类型）
// @Tags Dict
// @Router /api/v1/dict-types/:id/entries [get]
// dictEntryQueryDefs 字典条目可查询字段白名单（type_id 由路径参数固定）
var dictEntryQueryDefs = query.Defs{
	"label":      {Name: "label", Type: query.TypeString, Searchable: true, Filterable: true},
	"value":      {Name: "value", Type: query.TypeString, Searchable: true, Filterable: true},
	"remark":     {Name: "remark", Type: query.TypeString, Searchable: true},
	"status":     {Name: "status", Type: query.TypeInt, Filterable: true},
	"sort":       {Name: "sort", Type: query.TypeInt, Filterable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

func (h *DictEntryHandler) ListByTypeID(c *gin.Context) {
	typeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid type id")
		return
	}
	pq, err := query.Parse(c, dictEntryQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	entries, total, err := h.svc.ListByTypeID(c.Request.Context(), uint(typeID), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, entries, total, pq.Page, pq.PageSize)
}

// @Summary 公开字典条目（无需登录）
// @Tags Dict
// @Router /api/v1/public/dicts/:code/entries [get]
func (h *DictEntryHandler) GetPublicEntries(c *gin.Context) {
	code := c.Param("code")
	entries, err := h.svc.ListPublicByCode(c.Request.Context(), code)
	if err != nil {
		if code := appErrCode(err); code > 0 {
			response.ErrorJSON(c, code)
		} else {
			response.InternalError(c)
		}
		return
	}
	response.OKJSON(c, entries)
}

// DictEntryReq 字典条目请求体
type DictEntryReq struct {
	Label  string `json:"label" binding:"required"`
	Value  string `json:"value" binding:"required"`
	Sort   int    `json:"sort"`
	Status *int   `json:"status"`
	Remark string `json:"remark"`
}

// @Summary 创建字典条目
// @Tags Dict
// @Router /api/v1/dict-types/:id/entries [post]
func (h *DictEntryHandler) Create(c *gin.Context) {
	typeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid type id")
		return
	}
	var req DictEntryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	e, err := h.svc.Create(c.Request.Context(), uint(typeID), req.Label, req.Value, req.Sort, status, req.Remark)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, e)
}

// @Summary 更新字典条目
// @Tags Dict
// @Router /api/v1/dict-types/:id/entries/:entryId [put]
func (h *DictEntryHandler) Update(c *gin.Context) {
	typeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid type id")
		return
	}
	entryID, err := strconv.ParseUint(c.Param("entryId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid entry id")
		return
	}
	var req DictEntryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	if err := h.svc.Update(c.Request.Context(), uint(entryID), uint(typeID), req.Label, req.Value, req.Sort, status, req.Remark); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// @Summary 删除字典条目
// @Tags Dict
// @Router /api/v1/dict-types/:id/entries/:entryId [delete]
func (h *DictEntryHandler) Delete(c *gin.Context) {
	entryID, err := strconv.ParseUint(c.Param("entryId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid entry id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uint(entryID)); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}
