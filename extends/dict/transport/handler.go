package transport

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"kingfisher/core/errcode"
	"kingfisher/core/query"
	"kingfisher/core/response"
	"kingfisher/extends/dict/app"
)

// appErrCode 从 error 中提取 errcode，若非 AppError 则返回 -1
func appErrCode(err error) int {
	var e *app.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return -1
}

// ---- DictTypeHandler ----

type DictTypeHandler struct{ svc *app.DictTypeService }

func NewDictTypeHandler(svc *app.DictTypeService) *DictTypeHandler { return &DictTypeHandler{svc: svc} }

// @Summary 字典类型列表
// @Tags Dict
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.PageData} "字典类型列表"
// @Router /dict-types [get]
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
// @Produce json
// @Security BearerAuth
// @Param id path int true "类型ID"
// @Success 200 {object} response.Response{object} "字典类型详情"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 10501 {object} response.Response "字典类型不存在"
// @Router /dict-types/:id [get]
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
	Version  string `json:"version"`
}

// @Summary 创建字典类型
// @Tags Dict
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body DictTypeReq true "创建请求"
// @Success 200 {object} response.Response{object} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 10502 {object} response.Response "编码已存在"
// @Router /dict-types [post]
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
	t, err := h.svc.Create(c.Request.Context(), req.Code, req.Name, isPublic, status, req.Remark, req.Version)
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
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "类型ID"
// @Param body body DictTypeReq true "更新请求"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 10502 {object} response.Response "编码已存在"
// @Router /dict-types/:id [put]
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
	if err := h.svc.Update(c.Request.Context(), uint(id), req.Code, req.Name, isPublic, status, req.Remark, req.Version); err != nil {
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
// @Produce json
// @Security BearerAuth
// @Param id path int true "类型ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 10504 {object} response.Response "存在条目不可删除"
// @Router /dict-types/:id [delete]
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

// batchIDsReq 批量操作请求体（按 id）
type batchIDsReq struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// batchStatusReq 批量启用/禁用请求体
type batchStatusReq struct {
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Status *int   `json:"status" binding:"required"`
}

// @Summary 批量删除字典类型
// @Tags Dict
// @Router /dict-types/batch-delete [post]
func (h *DictTypeHandler) BatchDelete(c *gin.Context) {
	var req batchIDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.BatchDelete(c.Request.Context(), req.IDs); err != nil {
		if code := appErrCode(err); code > 0 {
			response.ErrorJSON(c, code)
		} else {
			response.InternalError(c)
		}
		return
	}
	response.OKJSON(c, nil)
}

// @Summary 批量启用/禁用字典类型
// @Tags Dict
// @Router /dict-types/batch-status [post]
func (h *DictTypeHandler) BatchUpdateStatus(c *gin.Context) {
	var req batchStatusReq
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

// ---- DictEntryHandler ----

type DictEntryHandler struct{ svc *app.DictEntryService }

func NewDictEntryHandler(svc *app.DictEntryService) *DictEntryHandler {
	return &DictEntryHandler{svc: svc}
}

// @Summary 字典条目列表（按类型）
// @Tags Dict
// @Produce json
// @Security BearerAuth
// @Param id path int true "类型ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.PageData} "条目列表"
// @Router /dict-types/:id/entries [get]
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
// @Produce json
// @Param code path string true "类型编码"
// @Success 200 {object} response.Response{object} "条目列表"
// @Failure 10501 {object} response.Response "类型不存在"
// @Failure 10505 {object} response.Response "类型未公开"
// @Router /public/dicts/:code/entries [get]
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
	Label   string `json:"label" binding:"required"`
	Value   string `json:"value" binding:"required"`
	Sort    int    `json:"sort"`
	Status  *int   `json:"status"`
	Remark  string `json:"remark"`
	Version string `json:"version"`
}

// @Summary 创建字典条目
// @Tags Dict
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "类型ID"
// @Param body body DictEntryReq true "创建请求"
// @Success 200 {object} response.Response{object} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /dict-types/:id/entries [post]
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
	e, err := h.svc.Create(c.Request.Context(), uint(typeID), req.Label, req.Value, req.Sort, status, req.Remark, req.Version)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, e)
}

// @Summary 更新字典条目
// @Tags Dict
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "类型ID"
// @Param entryId path int true "条目ID"
// @Param body body DictEntryReq true "更新请求"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /dict-types/:id/entries/:entryId [put]
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
	if err := h.svc.Update(c.Request.Context(), uint(entryID), uint(typeID), req.Label, req.Value, req.Sort, status, req.Remark, req.Version); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// @Summary 删除字典条目
// @Tags Dict
// @Produce json
// @Security BearerAuth
// @Param id path int true "类型ID"
// @Param entryId path int true "条目ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /dict-types/:id/entries/:entryId [delete]
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

// @Summary 批量删除字典条目
// @Tags Dict
// @Router /dict-types/:id/entries/batch-delete [post]
func (h *DictEntryHandler) BatchDelete(c *gin.Context) {
	typeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid type id")
		return
	}
	var req batchIDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.BatchDelete(c.Request.Context(), req.IDs); err != nil {
		response.InternalError(c)
		return
	}
	_ = typeID // 条目按全局 id 删除，type_id 仅用于路由定位
	response.OKJSON(c, nil)
}

// @Summary 批量启用/禁用字典条目
// @Tags Dict
// @Router /dict-types/:id/entries/batch-status [post]
func (h *DictEntryHandler) BatchUpdateStatus(c *gin.Context) {
	typeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid type id")
		return
	}
	var req batchStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.BatchUpdateStatus(c.Request.Context(), req.IDs, *req.Status); err != nil {
		response.InternalError(c)
		return
	}
	_ = typeID
	response.OKJSON(c, nil)
}
