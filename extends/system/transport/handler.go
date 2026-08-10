package transport

import (
	"github.com/gin-gonic/gin"

	"kingfisher/core/response"
	"kingfisher/extends/system/app"
)

type SystemHandler struct{ svc *app.SystemService }

func NewSystemHandler(svc *app.SystemService) *SystemHandler {
	return &SystemHandler{svc: svc}
}

// GetInfo 系统信息
// @Summary 系统信息
// @Tags System
// @Router /system/info [get]
func (h *SystemHandler) GetInfo(c *gin.Context) {
	info, err := h.svc.GetInfo(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, info)
}
