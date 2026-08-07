package transport

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"kingfisher/core/response"
	"kingfisher/extends/menu/app"
	"kingfisher/extends/menu/domain"
)

type MenuHandler struct{ svc *app.MenuService }

func NewMenuHandler(svc *app.MenuService) *MenuHandler { return &MenuHandler{svc: svc} }

// @Summary 菜单树（导航结构，权限校验在业务接口层）
// @Tags Menu
// @Router /api/v1/menus/tree [get]
func (h *MenuHandler) GetTree(c *gin.Context) {
	tree, err := h.svc.GetTree(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, tree)
}

// GetMyTree returns the menu tree filtered by the current user's role.
func (h *MenuHandler) GetMyTree(c *gin.Context) {
	roleID := c.GetUint("role_id")
	tree, err := h.svc.GetTreeForRole(c.Request.Context(), roleID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, tree)
}

func (h *MenuHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	m, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c)
		return
	}
	response.OKJSON(c, m)
}

// @Summary 创建菜单
// @Tags Menu
// @Router /api/v1/menus [post]
func (h *MenuHandler) Create(c *gin.Context) {
	var m domain.Menu
	if err := c.ShouldBindJSON(&m); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.Create(c.Request.Context(), &m); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, m)
}

// updateMenuReq 菜单更新请求体（白名单字段，防止 mass assignment）
type updateMenuReq struct {
	Name      *string `json:"name"`
	Icon      *string `json:"icon"`
	Path      *string `json:"path"`
	Component *string `json:"component"`
	Sort      *int    `json:"sort"`
	ParentID  *uint   `json:"parent_id"`
}

func (h *MenuHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req updateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.Path != nil {
		updates["path"] = *req.Path
	}
	if req.Component != nil {
		updates["component"] = *req.Component
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if err := h.svc.Update(c.Request.Context(), uint(id), updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, nil)
}

func (h *MenuHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, nil)
}
