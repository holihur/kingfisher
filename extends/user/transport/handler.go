package transport

import (
	"context"
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
	auditApp "kingfisher/extends/audit/app"
	"kingfisher/extends/user/app"
	"kingfisher/extends/user/domain"
)

// AuditLogger is injected by main to record auth events (login/logout).
type AuditLogger func(ctx context.Context, userID uint, username, action, resource, ip, userAgent string)

type AuthHandler struct {
	svc      *app.AuthService
	auditLog AuditLogger
}

// PermProvider resolves permission codes for a user. Injected by the RBAC module.
type PermProvider func(ctx context.Context, userID uint) ([]string, error)

type UserHandler struct {
	svc          *app.UserService
	getUserPerms PermProvider   // optional; when nil, falls back to svc.GetUserPermissions
	auditSvc     *auditApp.AuditService // optional; for GetMyLoginLogs
}

func NewAuthHandler(svc *app.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

// SetAuditLogger injects an audit logger from main.
func (h *AuthHandler) SetAuditLogger(fn AuditLogger) { h.auditLog = fn }

func NewUserHandler(svc *app.UserService) *UserHandler  { return &UserHandler{svc: svc} }

// SetAuditService injects the audit service for self-service login log queries.
func (h *UserHandler) SetAuditService(auditSvc *auditApp.AuditService) { h.auditSvc = auditSvc }

type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8,max=64"`
	Email    string `json:"email"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         domain.User `json:"user"`
	// LandingPage 角色落地页（登录后跳转的页面）
	LandingPage string `json:"landing_page"`
}

// @Summary 用户注册
// @Tags Auth
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := h.svc.Register(c.Request.Context(), req.Username, req.Password, req.Email)
	if err != nil {
		switch err.Error() {
		case "registration disabled":
			response.ErrorJSON(c, errcode.ErrRegistrationDisabled)
		default:
			response.ErrorJSON(c, errcode.ErrUserExists)
		}
		return
	}
	response.OKJSON(c, user)
}

// @Summary 用户登录
// @Tags Auth
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	access, refresh, user, landing, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.auditLogin(c, req.Username, "FAILURE", "")
		response.ErrorJSON(c, errcode.ErrPasswordWrong)
		return
	}
	h.auditLogin(c, req.Username, "SUCCESS", "")
	response.OKJSON(c, LoginResp{AccessToken: access, RefreshToken: refresh, User: *user, LandingPage: landing})
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Logout(c *gin.Context) {
	hdr := c.GetHeader("Authorization")
	if hdr == "" || !strings.HasPrefix(hdr, "Bearer ") {
		response.Unauthorized(c)
		return
	}
	_ = h.svc.RevokeToken(c.Request.Context(), hdr[7:])
	username := c.GetString("username")
	if username == "" {
		username = "unknown"
	}
	h.auditLogout(c, username)
	response.OKJSON(c, nil)
}

// @Summary 刷新Token
// @Tags Auth
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	access, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.ErrorJSON(c, errcode.ErrTokenInvalid)
		return
	}
	response.OKJSON(c, gin.H{"access_token": access})
}

// @Summary 创建用户
// @Tags User
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := h.svc.CreateUser(c.Request.Context(), req.Username, req.Password, req.Email)
	if err != nil {
		response.ErrorJSON(c, errcode.ErrUserExists)
		return
	}
	response.OKJSON(c, user)
}

// @Summary 获取用户
// @Tags User
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	user, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c)
		return
	}
	response.OKJSON(c, user)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetUint("user_id")
	user, err := h.svc.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c)
		return
	}
	response.OKJSON(c, user)
}

func (h *UserHandler) GetMyPermissions(c *gin.Context) {
	userID := c.GetUint("user_id")
	if h.getUserPerms != nil {
		perms, err := h.getUserPerms(c.Request.Context(), userID)
		if err == nil {
			response.OKJSON(c, perms)
			return
		}
	}
	perms, err := h.svc.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, perms)
}

// @Summary 用户列表
// @Tags User
// @Router /api/v1/users [get]
// userQueryDefs 用户列表可查询字段白名单
var userQueryDefs = query.Defs{
	"username":   {Name: "username", Type: query.TypeString, Searchable: true, Filterable: true},
	"email":      {Name: "email", Type: query.TypeString, Searchable: true, Filterable: true},
	"status":     {Name: "status", Type: query.TypeInt, Filterable: true},
	"role_id":    {Name: "role_id", Type: query.TypeUint, Filterable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
	"updated_at": {Name: "updated_at", Type: query.TypeTime, Filterable: true},
}

func (h *UserHandler) List(c *gin.Context) {
	pq, err := query.Parse(c, userQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	users, total, err := h.svc.List(c.Request.Context(), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, users, total, pq.Page, pq.PageSize)
}

type UpdateUserReq struct {
	Email  *string `json:"email"`
	Status *int    `json:"status"`
	RoleID *uint   `json:"role_id"`
}

func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if uint(id) == c.GetUint("user_id") {
		response.BadRequest(c, "不能修改自己")
		return
	}
	var req UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updates := map[string]any{}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.RoleID != nil {
		updates["role_id"] = *req.RoleID
	}
	if err := h.svc.Update(c.Request.Context(), uint(id), updates); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if uint(id) == c.GetUint("user_id") {
		response.BadRequest(c, "不能删除自己")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

type ChangePwdReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=64"`
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req ChangePwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		response.ErrorJSON(c, errcode.ErrPasswordWrong)
		return
	}
	response.OKJSON(c, nil)
}

func (h *UserHandler) RevokeSessions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.RevokeSessions(c.Request.Context(), uint(id)); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

func (h *AuthHandler) auditLogin(c *gin.Context, username, result, detail string) {
	if h.auditLog == nil {
		return
	}
	resource := result
	if detail != "" {
		resource = result + ": " + detail
	}
	h.auditLog(c.Request.Context(), 0, username, "LOGIN", resource, c.ClientIP(), c.Request.UserAgent())
}

func (h *AuthHandler) auditLogout(c *gin.Context, username string) {
	if h.auditLog == nil {
		return
	}
	h.auditLog(c.Request.Context(), c.GetUint("user_id"), username, "LOGOUT", "auth", c.ClientIP(), c.Request.UserAgent())
}

// UpdateMeReq 当前用户资料更新请求体
type UpdateMeReq struct {
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// UpdateMe 当前用户更新自己的资料
// @Summary 更新个人资料
// @Tags User
// @Router /api/v1/users/me [put]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req UpdateMeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.UpdateProfile(c.Request.Context(), userID, req.Email, req.Nickname, req.Avatar); err != nil {
		response.InternalError(c)
		return
	}
	// 返回最新用户信息
	user, err := h.svc.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, user)
}

// GetMyLoginLogs 当前用户查询自己的登录日志
// @Summary 我的登录日志
// @Tags User
// @Router /api/v1/users/me/login-logs [get]
func (h *UserHandler) GetMyLoginLogs(c *gin.Context) {
	if h.auditSvc == nil {
		response.OKJSON(c, []any{})
		return
	}
	userID := c.GetUint("user_id")
	pq, err := query.Parse(c, auditLogQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 强制只查当前用户的 LOGIN 记录
	pq.Filters = append(pq.Filters, query.Condition{Field: "user_id", Op: "eq", Value: userID})
	logs, total, err := h.auditSvc.FindAll(c.Request.Context(), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, logs, total, pq.Page, pq.PageSize)
}

// auditLogQueryDefs 登录日志可查询字段白名单
var auditLogQueryDefs = query.Defs{
	"user_id":    {Name: "user_id", Type: query.TypeUint, Filterable: true},
	"username":   {Name: "username", Type: query.TypeString, Searchable: true, Filterable: true},
	"ip":         {Name: "ip", Type: query.TypeString, Searchable: true},
	"user_agent": {Name: "user_agent", Type: query.TypeString, Searchable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

// UploadAvatar 当前用户上传头像
// @Summary 上传头像
// @Tags User
// @Router /api/v1/users/me/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetUint("user_id")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	// 校验文件类型（先检查扩展名，再校验真实内容）
	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
	default:
		response.BadRequest(c, "不支持的文件类型，仅支持 png/jpg/jpeg/gif/webp")
		return
	}

	// 校验大小（最大 2MB）
	if header.Size > 2<<20 {
		response.BadRequest(c, "文件大小不能超过 2MB")
		return
	}

	// 校验真实内容（magic bytes），防止伪造扩展名上传恶意文件
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	if n > 0 {
		detected := http.DetectContentType(buf[:n])
		if !strings.HasPrefix(detected, "image/") {
			response.BadRequest(c, "不支持的文件内容，仅支持图片文件")
			return
		}
	}
	// 确保目录存在
	uploadDir := "uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.InternalError(c)
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(savePath)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer dst.Close()

	// 写入已验证的 magic bytes + 剩余文件内容
	if _, err := dst.Write(buf[:n]); err != nil {
		response.InternalError(c)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		response.InternalError(c)
		return
	}

	// 更新用户 avatar 字段
	avatarURL := "/uploads/avatars/" + filename
	if err := h.svc.UpdateProfile(c.Request.Context(), userID, "", "", avatarURL); err != nil {
		response.InternalError(c)
		return
	}

	response.OKJSON(c, gin.H{"url": avatarURL})
}
