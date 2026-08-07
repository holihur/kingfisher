package transport

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"kingfisher/core/errcode"
	"kingfisher/core/response"
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
	svc        *app.UserService
	getUserPerms PermProvider // optional; when nil, falls back to svc.GetUserPermissions
}

func NewAuthHandler(svc *app.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

// SetAuditLogger injects an audit logger from main.
func (h *AuthHandler) SetAuditLogger(fn AuditLogger) { h.auditLog = fn }

func NewUserHandler(svc *app.UserService) *UserHandler  { return &UserHandler{svc: svc} }

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
		response.ErrorJSON(c, errcode.ErrUserExists)
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
	access, refresh, user, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.auditLogin(c, req.Username, "FAILURE", err.Error())
		switch err.Error() {
		case "too many attempts":
			response.ErrorJSON(c, errcode.ErrLoginFailed)
		case "user disabled":
			response.ErrorJSON(c, errcode.ErrUserDisabled)
		default:
			response.ErrorJSON(c, errcode.ErrPasswordWrong)
		}
		return
	}
	h.auditLogin(c, req.Username, "SUCCESS", "")
	response.OKJSON(c, LoginResp{AccessToken: access, RefreshToken: refresh, User: *user})
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
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	keyword := c.Query("keyword")
	users, total, err := h.svc.List(c.Request.Context(), page, pageSize, keyword)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, users, total, page, pageSize)
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
