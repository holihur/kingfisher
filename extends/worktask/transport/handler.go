package transport

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"kingfisher/core/dataaccess"
	"kingfisher/core/query"
	"kingfisher/core/response"
	"kingfisher/extends/worktask/app"
	"kingfisher/extends/worktask/domain"
)

type ScopeResolver func(ctx context.Context, userID uint, roleIDs []uint, roleCodes []string) (dataaccess.Scope, error)

type Handler struct {
	service       *app.Service
	scopeResolver ScopeResolver
}

func NewHandler(service *app.Service, resolver ScopeResolver) *Handler {
	return &Handler{service: service, scopeResolver: resolver}
}

var taskQueryDefs = query.Defs{
	"title":         {Name: "title", Type: query.TypeString, Searchable: true, Filterable: true},
	"department_id": {Name: "department_id", Type: query.TypeUint, Filterable: true},
	"status":        {Name: "status", Type: query.TypeString, Filterable: true},
}

func (h *Handler) List(c *gin.Context) {
	pq, err := query.Parse(c, taskQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	scope, err := h.resolveScope(c)
	if err != nil {
		response.InternalError(c)
		return
	}
	items, total, err := h.service.List(c.Request.Context(), pq, scope)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, items, total, pq.Page, pq.PageSize)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	scope, err := h.resolveScope(c)
	if err != nil {
		response.InternalError(c)
		return
	}
	item, err := h.service.GetByID(c.Request.Context(), uint(id), scope)
	if err != nil {
		response.NotFound(c)
		return
	}
	response.OKJSON(c, item)
}

type taskRequest struct {
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description"`
	DepartmentID uint   `json:"department_id"`
	Status       string `json:"status" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	var req taskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item := &domain.Task{Title: req.Title, Description: req.Description, OwnerID: c.GetUint("user_id"), DepartmentID: req.DepartmentID, Status: req.Status}
	if err := h.service.Create(c.Request.Context(), item); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, item)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req taskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updates := map[string]any{"title": req.Title, "description": req.Description, "department_id": req.DepartmentID, "status": req.Status}
	scope, err := h.resolveScope(c)
	if err != nil {
		response.InternalError(c)
		return
	}
	if err := h.service.Update(c.Request.Context(), uint(id), updates, scope); err != nil {
		response.NotFound(c)
		return
	}
	response.OKJSON(c, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	scope, err := h.resolveScope(c)
	if err != nil {
		response.InternalError(c)
		return
	}
	if err := h.service.Delete(c.Request.Context(), uint(id), scope); err != nil {
		response.NotFound(c)
		return
	}
	response.OKJSON(c, nil)
}

func (h *Handler) resolveScope(c *gin.Context) (dataaccess.Scope, error) {
	if h.scopeResolver != nil {
		roleIDs := []uint{}
		if value, exists := c.Get("role_ids"); exists {
			if ids, ok := value.([]uint); ok {
				roleIDs = ids
			}
		}
		return h.scopeResolver(c.Request.Context(), c.GetUint("user_id"), roleIDs, c.GetStringSlice("roles"))
	}
	return scopeFor(c), nil
}

func scopeFor(c *gin.Context) dataaccess.Scope {
	for _, role := range c.GetStringSlice("roles") {
		if role == "admin" {
			return dataaccess.All()
		}
	}
	return dataaccess.Self("owner_id", c.GetUint("user_id"))
}
