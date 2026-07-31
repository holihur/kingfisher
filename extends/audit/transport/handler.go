package transport

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"kingfisher/core/response"
	"kingfisher/extends/audit/app"
)

type AuditHandler struct{ svc *app.AuditService }

func NewAuditHandler(svc *app.AuditService) *AuditHandler { return &AuditHandler{svc: svc} }
func (h *AuditHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	filters := map[string]any{}
	if v := c.Query("user_id"); v != "" {
		filters["user_id"] = v
	}
	if v := c.Query("resource"); v != "" {
		filters["resource"] = v
	}
	if v := c.Query("action"); v != "" {
		filters["action"] = v
	}
	logs, total, err := h.svc.FindAll(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, logs, total, page, pageSize)
}
