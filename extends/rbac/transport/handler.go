package transport

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"kingfisher/core/query"
	"kingfisher/core/response"
	"kingfisher/extends/rbac/app"
	"kingfisher/extends/rbac/domain"
)

type RoleHandler struct{ svc *app.RoleService }
type PermHandler struct{ svc *app.PermService }
type AssignPermReq struct {
	PermissionIDs []uint `json:"permission_ids"`
}
type AssignMenuReq struct {
	MenuIDs []uint `json:"menu_ids"`
}
type batchIDsReq struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}
type batchStatusReq struct {
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Status *int   `json:"status" binding:"required"`
}

func NewRoleHandler(svc *app.RoleService) *RoleHandler { return &RoleHandler{svc: svc} }
func NewPermHandler(svc *app.PermService) *PermHandler { return &PermHandler{svc: svc} }

// @Summary 角色列表
// @Tags Role
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.PageData} "角色列表"
// @Router /roles [get]
// roleQueryDefs 角色列表可查询字段白名单
var roleQueryDefs = query.Defs{
	"name":        {Name: "name", Type: query.TypeString, Searchable: true, Filterable: true},
	"code":        {Name: "code", Type: query.TypeString, Searchable: true, Filterable: true},
	"description": {Name: "description", Type: query.TypeString, Searchable: true},
	"status":      {Name: "status", Type: query.TypeInt, Filterable: true},
	"level":       {Name: "level", Type: query.TypeInt, Filterable: true},
	"created_at":  {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

func (h *RoleHandler) List(c *gin.Context) {
	pq, err := query.Parse(c, roleQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	roles, total, err := h.svc.List(c.Request.Context(), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, roles, total, pq.Page, pq.PageSize)
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c)
		return
	}
	response.OKJSON(c, role)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var r domain.Role
	if err := c.ShouldBindJSON(&r); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.Create(c.Request.Context(), &r); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, r)
}

// updateRoleReq 角色更新请求体（白名单字段，防止 mass assignment）
type updateRoleReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *int    `json:"status"`
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req updateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if err := h.svc.Update(c.Request.Context(), uint(id), updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, nil)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, nil)
}

func (h *RoleHandler) BatchDelete(c *gin.Context) {
	var req batchIDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.BatchDelete(c.Request.Context(), req.IDs); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, nil)
}

func (h *RoleHandler) BatchUpdateStatus(c *gin.Context) {
	var req batchStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.BatchUpdateStatus(c.Request.Context(), req.IDs, *req.Status); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, nil)
}

func (h *RoleHandler) GetPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	perms, _ := h.svc.GetRolePermissions(c.Request.Context(), uint(id))
	response.OKJSON(c, perms)
}

func (h *RoleHandler) AssignPerms(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req AssignPermReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.AssignPermissions(c.Request.Context(), uint(id), req.PermissionIDs); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, nil)
}

func (h *RoleHandler) GetMenus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	menus, _ := h.svc.GetRoleMenus(c.Request.Context(), uint(id))
	response.OKJSON(c, menus)
}

func (h *RoleHandler) AssignMenus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req AssignMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.AssignMenus(c.Request.Context(), uint(id), req.MenuIDs); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, nil)
}

// @Summary 权限列表
// @Tags Role
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "权限列表"
// @Router /permissions [get]
func (h *PermHandler) List(c *gin.Context) {
	perms, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, perms)
}
