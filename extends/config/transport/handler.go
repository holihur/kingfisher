package transport

import (
	"github.com/gin-gonic/gin"

	"kingfisher/core/errcode"
	"kingfisher/core/response"
	"kingfisher/extends/config/app"
)

type ConfigHandler struct{ svc *app.ConfigService }

func NewConfigHandler(svc *app.ConfigService) *ConfigHandler { return &ConfigHandler{svc: svc} }

// @Summary 配置列表
// @Tags Config
// @Router /api/v1/configs [get]
func (h *ConfigHandler) GetAll(c *gin.Context) {
	configs, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, configs)
}

// GetPublicAll 公开配置列表（无需登录）
// @Summary 公开配置列表
// @Tags Config
// @Router /api/v1/public/configs [get]
func (h *ConfigHandler) GetPublicAll(c *gin.Context) {
	configs, err := h.svc.GetAllPublic(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, configs)
}

// GetPublic 公开配置单条（无需登录；非公开项视为不存在）
// @Summary 公开配置单条
// @Tags Config
// @Router /api/v1/public/configs/:key [get]
func (h *ConfigHandler) GetPublic(c *gin.Context) {
	v, err := h.svc.GetPublic(c.Request.Context(), c.Param("key"))
	if err != nil {
		response.ErrorJSON(c, errcode.ErrConfigNotFound)
		return
	}
	response.OKJSON(c, v)
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
	Value    string `json:"value" binding:"required"`
	IsPublic *bool  `json:"is_public"`
	Version  string `json:"version"`
}

func (h *ConfigHandler) Set(c *gin.Context) {
	var req SetConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	isPublic := false
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}
	_ = h.svc.Set(c.Request.Context(), c.Param("key"), req.Value, isPublic, req.Version)
	response.OKJSON(c, nil)
}

func (h *ConfigHandler) Delete(c *gin.Context) {
	_ = h.svc.Delete(c.Request.Context(), c.Param("key"))
	response.OKJSON(c, nil)
}
