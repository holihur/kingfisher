package transport

import (
	"github.com/gin-gonic/gin"

	"kingfisher/core/query"
	"kingfisher/core/response"
	"kingfisher/extends/audit/app"
)

type AuditHandler struct{ svc *app.AuditService }

func NewAuditHandler(svc *app.AuditService) *AuditHandler { return &AuditHandler{svc: svc} }
// auditQueryDefs 审计日志可查询字段白名单
var auditQueryDefs = query.Defs{
	"username":   {Name: "username", Type: query.TypeString, Searchable: true, Filterable: true},
	"ip":         {Name: "ip", Type: query.TypeString, Searchable: true},
	"user_agent": {Name: "user_agent", Type: query.TypeString, Searchable: true},
	"user_id":    {Name: "user_id", Type: query.TypeUint, Filterable: true},
	"resource":   {Name: "resource", Type: query.TypeString, Filterable: true},
	"action":     {Name: "action", Type: query.TypeString, Filterable: true},
	"resource_id": {Name: "resource_id", Type: query.TypeUint, Filterable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

func (h *AuditHandler) List(c *gin.Context) {
	pq, err := query.Parse(c, auditQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	logs, total, err := h.svc.FindAll(c.Request.Context(), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, logs, total, pq.Page, pq.PageSize)
}
