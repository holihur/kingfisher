package transport

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"kingfisher/core/response"
	"kingfisher/extends/rbac/app"
	"kingfisher/extends/rbac/domain"
)

type RoleHandler struct{ svc *app.RoleService }
type PermHandler struct{ svc *app.PermService }
type AssignPermReq struct{ PermissionIDs []uint `json:"permission_ids"` }
type AssignMenuReq struct{ MenuIDs []uint `json:"menu_ids"` }

func NewRoleHandler(svc *app.RoleService) *RoleHandler { return &RoleHandler{svc: svc} }
func NewPermHandler(svc *app.PermService) *PermHandler  { return &PermHandler{svc: svc} }

func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.svc.List(c.Request.Context())
	if err != nil { response.InternalError(c); return }
	response.OKJSON(c, roles)
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil { response.NotFound(c); return }
	response.OKJSON(c, role)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var r domain.Role
	if err := c.ShouldBindJSON(&r); err != nil { response.BadRequest(c, err.Error()); return }
	_ = h.svc.Create(c.Request.Context(), &r)
	response.OKJSON(c, r)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var m map[string]any
	if err := c.ShouldBindJSON(&m); err != nil { response.BadRequest(c, err.Error()); return }
	_ = h.svc.Update(c.Request.Context(), uint(id), m)
	response.OKJSON(c, nil)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	_ = h.svc.Delete(c.Request.Context(), uint(id))
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
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	_ = h.svc.AssignPermissions(c.Request.Context(), uint(id), req.PermissionIDs)
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
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, err.Error()); return }
	_ = h.svc.AssignMenus(c.Request.Context(), uint(id), req.MenuIDs)
	response.OKJSON(c, nil)
}

func (h *PermHandler) List(c *gin.Context) {
	perms, err := h.svc.List(c.Request.Context())
	if err != nil { response.InternalError(c); return }
	response.OKJSON(c, perms)
}
