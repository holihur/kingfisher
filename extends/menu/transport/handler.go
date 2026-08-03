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

// @Summary 菜单树
// @Tags Menu
// @Router /api/v1/menus/tree [get]
func (h *MenuHandler) GetTree(c *gin.Context) {
	tree, err := h.svc.GetTree(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	// Filter by user permissions injected by RBAC middleware
	perms, _ := c.Get("permissions")
	psMap, _ := perms.(map[string]bool)
	if psMap != nil {
		tree = filterTreeByPerms(tree, psMap)
	}
	response.OKJSON(c, tree)
}

// filterTreeByPerms removes menu items the user doesn't have permission for.
// Items with no permission requirement (empty string) are visible to all.
func filterTreeByPerms(menus []domain.Menu, perms map[string]bool) []domain.Menu {
	filtered := make([]domain.Menu, 0, len(menus))
	for _, m := range menus {
		if m.Permission != "" && !perms[m.Permission] {
			continue // user doesn't have this permission, skip
		}
		if len(m.Children) > 0 {
			m.Children = filterTreeByPerms(m.Children, perms)
		}
		filtered = append(filtered, m)
	}
	return filtered
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

func (h *MenuHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var m map[string]any
	if err := c.ShouldBindJSON(&m); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), uint(id), m); err != nil {
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
