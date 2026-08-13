package transport

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"kingfisher/core/query"
	"kingfisher/core/response"
	"kingfisher/extends/doc/app"
	"kingfisher/extends/doc/domain"
)

type DocHandler struct{ svc *app.DocService }

func NewDocHandler(svc *app.DocService) *DocHandler { return &DocHandler{svc: svc} }

// appErrCode 从 error 中提取 errcode，若非 AppError 则返回 -1
func appErrCode(err error) int {
	var e *app.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return -1
}

// currentUser 从 gin context 提取当前用户可见性上下文（由 AuthMiddleware 注入）。
func currentUser(c *gin.Context) (userID uint, roleIDs []uint, isAdmin bool) {
	userID = c.GetUint("user_id")
	if v, exists := c.Get("role_ids"); exists {
		if ids, ok := v.([]uint); ok {
			roleIDs = ids
		}
	}
	for _, r := range c.GetStringSlice("roles") {
		if r == "admin" {
			isAdmin = true
			break
		}
	}
	return
}

// handleSvcErr 统一映射 service 错误：AppError → errcode，否则 500。
func handleSvcErr(c *gin.Context, err error) {
	if code := appErrCode(err); code > 0 {
		response.ErrorJSON(c, code)
		return
	}
	response.InternalError(c)
}

// ———— 目录 ————

// @Summary 当前用户可见目录树
// @Tags Doc
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.DocDirectory} "目录树"
// @Router /docs/tree [get]
func (h *DocHandler) GetTree(c *gin.Context) {
	userID, roleIDs, isAdmin := currentUser(c)
	tree, err := h.svc.GetTree(c.Request.Context(), userID, roleIDs, isAdmin)
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, tree)
}

// CreateDirReq 创建目录请求体
type CreateDirReq struct {
	ParentID   uint    `json:"parent_id"`
	Name       *string `json:"name" binding:"required"`
	Sort       *int    `json:"sort"`
	Visibility *string `json:"visibility"` // shared | private（默认 shared）
}

// @Summary 创建目录
// @Tags Doc
// @Security BearerAuth
// @Router /docs/dirs [post]
func (h *DocHandler) CreateDir(c *gin.Context) {
	var req CreateDirReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	vis := domain.VisibilityShared
	if req.Visibility != nil && (*req.Visibility == domain.VisibilityPrivate || *req.Visibility == domain.VisibilityShared) {
		vis = *req.Visibility
	}
	d := &domain.DocDirectory{ParentID: req.ParentID, Name: *req.Name, Sort: reqOr(req.Sort, 0), Status: 1, Visibility: vis}
	if err := h.svc.CreateDir(c.Request.Context(), d); err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, d)
}

// UpdateDirReq 目录更新请求体（白名单字段，防 mass assignment）
type UpdateDirReq struct {
	Name       *string `json:"name"`
	ParentID   *uint   `json:"parent_id"`
	Sort       *int    `json:"sort"`
	Status     *int    `json:"status"`
	Visibility *string `json:"visibility"` // shared | private
}

// @Summary 更新目录（改名/移动/排序）
// @Tags Doc
// @Security BearerAuth
// @Router /docs/dirs/:id [put]
func (h *DocHandler) UpdateDir(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req UpdateDirReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Visibility != nil && (*req.Visibility == domain.VisibilityShared || *req.Visibility == domain.VisibilityPrivate) {
		updates["visibility"] = *req.Visibility
	}
	if len(updates) == 0 {
		response.BadRequest(c, "no fields to update")
		return
	}
	if err := h.svc.UpdateDir(c.Request.Context(), uint(id), updates); err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, nil)
}

// @Summary 删除目录（有子目录/文档则拒绝）
// @Tags Doc
// @Security BearerAuth
// @Router /docs/dirs/:id [delete]
func (h *DocHandler) DeleteDir(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.DeleteDir(c.Request.Context(), uint(id)); err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, nil)
}

// ———— 文档 ————

// docQueryDefs 文档列表可查询字段白名单（dir_id 为固定作用域，不在此）
var docQueryDefs = query.Defs{
	"title":      {Name: "title", Type: query.TypeString, Searchable: true, Filterable: true},
	"status":     {Name: "status", Type: query.TypeString, Filterable: true},
	"visibility": {Name: "visibility", Type: query.TypeString, Filterable: true},
	"owner_id":   {Name: "owner_id", Type: query.TypeUint, Filterable: true},
	"sort":       {Name: "sort", Type: query.TypeInt, Filterable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
	"updated_at": {Name: "updated_at", Type: query.TypeTime, Filterable: true},
}

// @Summary 目录内文档分页列表
// @Tags Doc
// @Produce json
// @Security BearerAuth
// @Param dir_id query int true "目录 ID"
// @Success 200 {object} response.Response{data=response.PageData} "文档列表"
// @Router /docs [get]
func (h *DocHandler) ListDocs(c *gin.Context) {
	dirID, err := strconv.ParseUint(c.Query("dir_id"), 10, 64)
	if err != nil || dirID == 0 {
		response.BadRequest(c, "dir_id is required")
		return
	}
	pq, err := query.Parse(c, docQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, roleIDs, isAdmin := currentUser(c)
	docs, total, err := h.svc.ListDocs(c.Request.Context(), uint(dirID), pq, userID, roleIDs, isAdmin)
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.PageJSON(c, docs, total, pq.Page, pq.PageSize)
}

// @Summary 文档详情
// @Tags Doc
// @Produce json
// @Security BearerAuth
// @Router /docs/:id [get]
func (h *DocHandler) GetDoc(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	userID, roleIDs, isAdmin := currentUser(c)
	doc, err := h.svc.GetDoc(c.Request.Context(), uint(id), userID, roleIDs, isAdmin)
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, doc)
}

// @Summary 公开文档（无需登录）
// @Tags Doc
// @Produce json
// @Router /public/docs/:id [get]
func (h *DocHandler) GetPublicDoc(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	doc, err := h.svc.GetPublicDoc(c.Request.Context(), uint(id))
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, doc)
}

// @Summary 公开目录树（匿名，shared 目录 + 公开文档）
// @Tags Doc
// @Router /public/docs/tree [get]
func (h *DocHandler) GetPublicTree(c *gin.Context) {
	tree, err := h.svc.GetPublicTree(c.Request.Context())
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, tree)
}

// UploadImage 上传文档图片（存 uploads/docs/，返回可访问 URL）。
// 独立于 config 的 upload-image：文档图片单独目录，权限用 doc:update。
// @Summary 上传文档图片
// @Tags Doc
// @Security BearerAuth
// @Accept multipart/form-data
// @Param file formData file true "图片文件（png/jpg/jpeg/gif/webp，≤2MB）"
// @Success 200 {object} response.Response{data=map[string]string} "url"
// @Router /docs/upload [post]
func (h *DocHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	defer func() { _ = file.Close() }()

	// 校验扩展名 + 大小 + magic bytes（与 config/upload-image 一致）
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

	uploadDir := "uploads/docs"
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

	response.OKJSON(c, gin.H{"url": "/uploads/docs/" + filename})
}

// CreateDocReq 创建文档请求体
type CreateDocReq struct {
	DirID      uint   `json:"dir_id" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	Visibility string `json:"visibility"`
	Note       string `json:"note"`
}

// @Summary 创建文档（初始 draft，版本 1）
// @Tags Doc
// @Security BearerAuth
// @Router /docs [post]
func (h *DocHandler) CreateDoc(c *gin.Context) {
	var req CreateDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, roleIDs, isAdmin := currentUser(c)
	vis := defaultStr(req.Visibility, domain.VisibilityShared)
	doc, err := h.svc.CreateDoc(c.Request.Context(), req.DirID, req.Title, req.Content, userID, vis, req.Note, roleIDs, isAdmin)
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, doc)
}

// UpdateDocReq 更新文档请求体
type UpdateDocReq struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	Visibility string `json:"visibility"` // shared | private（空 = 不修改）
	Note       string `json:"note"`
}

// @Summary 更新文档（追加新版本）
// @Tags Doc
// @Security BearerAuth
// @Router /docs/:id [put]
func (h *DocHandler) UpdateDoc(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req UpdateDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, _, isAdmin := currentUser(c)
	doc, err := h.svc.UpdateDoc(c.Request.Context(), uint(id), req.Title, req.Content, req.Visibility, req.Note, userID, isAdmin)
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, doc)
}

// PublishReq 发布/撤稿请求体（预留：当前无 body 字段）
type PublishReq struct{}

// @Summary 发布文档 draft → published
// @Tags Doc
// @Security BearerAuth
// @Router /docs/:id/publish [put]
func (h *DocHandler) Publish(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	userID, _, isAdmin := currentUser(c)
	if err := h.svc.Publish(c.Request.Context(), uint(id), userID, isAdmin); err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, nil)
}

// @Summary 撤稿 published → draft
// @Tags Doc
// @Security BearerAuth
// @Router /docs/:id/unpublish [put]
func (h *DocHandler) Unpublish(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	userID, _, isAdmin := currentUser(c)
	if err := h.svc.Unpublish(c.Request.Context(), uint(id), userID, isAdmin); err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, nil)
}

// @Summary 文档版本历史列表
// @Tags Doc
// @Security BearerAuth
// @Router /docs/:id/versions [get]
func (h *DocHandler) ListVersions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	userID, _, isAdmin := currentUser(c)
	vers, err := h.svc.ListVersions(c.Request.Context(), uint(id), userID, isAdmin)
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, vers)
}

// @Summary 查看指定版本内容
// @Tags Doc
// @Security BearerAuth
// @Router /docs/:id/versions/:no [get]
func (h *DocHandler) GetVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	no, err := strconv.Atoi(c.Param("no"))
	if err != nil || no < 1 {
		response.BadRequest(c, "invalid version no")
		return
	}
	userID, _, isAdmin := currentUser(c)
	ver, err := h.svc.GetVersion(c.Request.Context(), uint(id), no, userID, isAdmin)
	if err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, ver)
}

// RestoreReq 还原到指定版本请求体
type RestoreReq struct {
	VersionNo int `json:"version_no" binding:"required"`
}

// @Summary 还原到指定版本（追加新版本）
// @Tags Doc
// @Security BearerAuth
// @Router /docs/:id/restore [post]
func (h *DocHandler) Restore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req RestoreReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, _, isAdmin := currentUser(c)
	if err := h.svc.Restore(c.Request.Context(), uint(id), req.VersionNo, userID, isAdmin); err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, nil)
}

// @Summary 删除文档（级联删版本）
// @Tags Doc
// @Security BearerAuth
// @Router /docs/:id [delete]
func (h *DocHandler) DeleteDoc(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	userID, _, isAdmin := currentUser(c)
	if err := h.svc.DeleteDoc(c.Request.Context(), uint(id), userID, isAdmin); err != nil {
		handleSvcErr(c, err)
		return
	}
	response.OKJSON(c, nil)
}

// —— helpers ——

func defaultStr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func reqOr(v *int, def int) int {
	if v == nil {
		return def
	}
	return *v
}
