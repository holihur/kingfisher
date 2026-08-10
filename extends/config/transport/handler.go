package transport

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"kingfisher/core/errcode"
	"kingfisher/core/query"
	"kingfisher/core/response"
	"kingfisher/extends/config/app"
)

type ConfigHandler struct{ svc *app.ConfigService }

func NewConfigHandler(svc *app.ConfigService) *ConfigHandler { return &ConfigHandler{svc: svc} }

// configQueryDefs 配置列表可查询字段白名单
var configQueryDefs = query.Defs{
	"key":        {Name: "key", Type: query.TypeString, Searchable: true, Filterable: true},
	"value":      {Name: "value", Type: query.TypeString, Searchable: true},
	"remark":     {Name: "remark", Type: query.TypeString, Searchable: true},
	"is_public":  {Name: "is_public", Type: query.TypeBool, Filterable: true},
	"version":    {Name: "version", Type: query.TypeString, Filterable: true},
	"render":     {Name: "render", Type: query.TypeString, Filterable: true},
	"group_id":   {Name: "group_id", Type: query.TypeUint, Filterable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
	"updated_at": {Name: "updated_at", Type: query.TypeTime, Filterable: true},
}

// @Summary 配置列表
// @Tags Config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{object} "配置列表"
// @Router /api/v1/configs [get]
func (h *ConfigHandler) GetAll(c *gin.Context) {
	pq, err := query.Parse(c, configQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	configs, total, err := h.svc.List(c.Request.Context(), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, configs, total, pq.Page, pq.PageSize)
}

// GetPublicAll 公开配置列表（无需登录）
// @Summary 公开配置列表
// @Tags Config
// @Produce json
// @Success 200 {object} response.Response{object} "公开配置列表"
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
// @Produce json
// @Param key path string true "配置键"
// @Success 200 {object} response.Response{object} "配置详情"
// @Failure 10401 {object} response.Response "不存在"
// @Router /api/v1/public/configs/:key [get]
func (h *ConfigHandler) GetPublic(c *gin.Context) {
	v, err := h.svc.GetPublic(c.Request.Context(), c.Param("key"))
	if err != nil {
		response.ErrorJSON(c, errcode.ErrConfigNotFound)
		return
	}
	response.OKJSON(c, v)
}

func (h * // @Produce json
// @Security BearerAuth
// @Param key path string true "配置键"
// @Success 200 {object} response.Response{object} "配置详情"
// @Failure 10401 {object} response.Response "不存在"
ConfigHandler) Get(c *gin.Context) {
	v, err := h.svc.Get(c.Request.Context(), c.Param("key"))
	if err != nil {
		response.ErrorJSON(c, errcode.ErrConfigNotFound)
		return
	}
	response.OKJSON(c, v)
}

type SetConfigReq struct {
	Value         string `json:"value" binding:"required"`
	IsPublic      *bool  `json:"is_public"`
	Version       string `json:"version"`
	Render        string `json:"render"`
	RenderOptions string `json:"render_options"`
	GroupID       uint   `json:"group_id"`
}

func (h * // @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键"
// @Param body body SetConfigReq true "配置值"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "参数错误"
ConfigHandler) Set(c *gin.Context) {
	var req SetConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	isPublic := false
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	} else if existing, err := h.svc.Get(c.Request.Context(), c.Param("key")); err == nil {
		// 未传 is_public 时保留原值，避免只更新 value 时误重置公开状态
		isPublic = existing.IsPublic
	}
	_ = h.svc.Set(c.Request.Context(), c.Param("key"), req.Value, isPublic, req.Version, req.Render, req.RenderOptions, req.GroupID)
	response.OKJSON(c, nil)
}

func (h * // @Produce json
// @Security BearerAuth
// @Param key path string true "配置键"
// @Success 200 {object} response.Response "删除成功"
ConfigHandler) Delete(c *gin.Context) {
	_ = h.svc.Delete(c.Request.Context(), c.Param("key"))
	response.OKJSON(c, nil)
}

// batchKeysReq 配置批量删除请求体（按 key）
type batchKeysReq struct {
	Keys []string `json:"keys" binding:"required,min=1"`
}

func (h *ConfigHandler) BatchDelete(c *gin.Context) {
	var req batchKeysReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.BatchDelete(c.Request.Context(), req.Keys); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// UploadImage 上传图片配置（如登录页封面 site_login_cover）
// @Summary 上传配置图片
// @Tags Config
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "图片文件（png/jpg/jpeg/gif/webp，≤2MB）"
// @Success 200 {object} response.Response{data=object{url=string}} "上传成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/v1/configs/upload-image [post]
func (h *ConfigHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	defer func() { _ = file.Close() }()

	// 校验文件类型（扩展名 + 真实内容 magic bytes），与头像上传一致
	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
	default:
		response.BadRequest(c, "不支持的文件类型，仅支持 png/jpg/jpeg/gif/webp")
		return
	}
	if header.Size > 2<<20 {
		response.BadRequest(c, "文件大小不能超过 2MB")
		return
	}
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	if n > 0 {
		detected := http.DetectContentType(buf[:n])
		if !strings.HasPrefix(detected, "image/") {
			response.BadRequest(c, "不支持的文件内容，仅支持图片文件")
			return
		}
	}

	uploadDir := "uploads/configs"
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		response.InternalError(c)
		return
	}
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)
	//nolint:gosec // G304: filename 由时间戳生成，非用户输入，无路径注入风险
	dst, err := os.Create(savePath)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer func() { _ = dst.Close() }()

	if _, err := dst.Write(buf[:n]); err != nil {
		response.InternalError(c)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		response.InternalError(c)
		return
	}

	response.OKJSON(c, gin.H{"url": "/uploads/configs/" + filename})
}

// ConfigGroupHandler 配置分组 CRUD
type ConfigGroupHandler struct{ svc *app.ConfigGroupService }

func NewConfigGroupHandler(svc *app.ConfigGroupService) *ConfigGroupHandler {
	return &ConfigGroupHandler{svc: svc}
}

// @Summary 配置分组列表
// @Tags Config
// @Router /api/v1/config-groups [get]
func (h * // @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.ConfigGroup} "分组列表"
ConfigGroupHandler) List(c *gin.Context) {
	groups, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, groups)
}

type ConfigGroupReq struct {
	Name string `json:"name" binding:"required"`
	Sort int    `json:"sort"`
}

func (h * // @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ConfigGroupReq true "创建请求"
// @Success 200 {object} response.Response{object} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
ConfigGroupHandler) Create(c *gin.Context) {
	var req ConfigGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	g, err := h.svc.Create(c.Request.Context(), req.Name, req.Sort)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, g)
}

func (h * // @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "分组ID"
// @Param body body ConfigGroupReq true "更新请求"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "参数错误"
ConfigGroupHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req ConfigGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), uint(id), req.Name, req.Sort); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

func (h * // @Produce json
// @Security BearerAuth
// @Param id path int true "分组ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "参数错误"
ConfigGroupHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}
