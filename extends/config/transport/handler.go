package transport

import (
	"github.com/gin-gonic/gin"

	"kingfisher/core/errcode"
	"kingfisher/core/response"
	"kingfisher/extends/config/app"
)

type ConfigHandler struct{ svc *app.ConfigService }

func NewConfigHandler(svc *app.ConfigService) *ConfigHandler { return &ConfigHandler{svc: svc} }

func (h *ConfigHandler) GetAll(c *gin.Context) {
	configs, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, configs)
}

func (h *ConfigHandler) Get(c *gin.Context) {
	v, err := h.svc.Get(c.Request.Context(), c.Param("key"))
	if err != nil {
		response.ErrorJSON(c, errcode.ErrConfigNotFound)
		return
	}
	response.OKJSON(c, v)
}

type SetConfigReq struct {
	Value string `json:"value" binding:"required"`
}

func (h *ConfigHandler) Set(c *gin.Context) {
	var req SetConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	_ = h.svc.Set(c.Request.Context(), c.Param("key"), req.Value)
	response.OKJSON(c, nil)
}

func (h *ConfigHandler) Delete(c *gin.Context) {
	_ = h.svc.Delete(c.Request.Context(), c.Param("key"))
	response.OKJSON(c, nil)
}
