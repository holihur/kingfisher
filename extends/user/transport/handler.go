package transport

import (
	"context"
	"encoding/json"
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
// result: success | failure（登录成功/失败）
type AuditLogger func(ctx context.Context, userID uint, username, action, resource, result string, ip, userAgent string)

type AuthHandler struct {
	svc      *app.AuthService
	auditLog AuditLogger
}

// PermProvider resolves permission codes for a user. Injected by the RBAC module.
type PermProvider func(ctx context.Context, userID uint) ([]string, error)

type UserHandler struct {
	svc          *app.UserService
	getUserPerms PermProvider           // optional; when nil, falls back to svc.GetUserPermissions
	auditSvc     *auditApp.AuditService // optional; for GetMyLoginLogs
}

func NewAuthHandler(svc *app.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

func (h *AuthHandler) SetAuditLogger(fn AuditLogger) { h.auditLog = fn }

func NewUserHandler(svc *app.UserService) *UserHandler { return &UserHandler{svc: svc} }

func (h *UserHandler) SetAuditService(auditSvc *auditApp.AuditService) { h.auditSvc = auditSvc }

// ---------- request / response types ----------

type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=32" example:"newuser"`
	Password string `json:"password" binding:"required,min=8,max=64,password" example:"Abcd1234"`
	Email    string `json:"email" example:"user@example.com"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"Abcd1234"`
}

type LoginResp struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         domain.User `json:"user"`
	LandingPage  string      `json:"landing_page"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ForgotPasswordReq struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

type ResetPasswordReq struct {
	Token       string `json:"token" binding:"required" example:"abc123"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=64,password" example:"NewPass123"`
}

type UpdateUserReq struct {
	Email   *string `json:"email"`
	Status  *int    `json:"status"`
	RoleIDs []uint  `json:"role_ids"`
	DeptIDs []uint  `json:"dept_ids"`
}

// CreateUserReq 管理员创建用户（可指定多个角色与部门）
type CreateUserReq struct {
	Username string `json:"username" binding:"required,min=3,max=32" example:"newuser"`
	Password string `json:"password" binding:"required,min=8,max=64,password" example:"Abcd1234"`
	Email    string `json:"email" example:"user@example.com"`
	RoleIDs  []uint `json:"role_ids"`
	DeptIDs  []uint `json:"dept_ids"`
}

type CreateSubAccountReq struct {
	Username string `json:"username" binding:"required,min=3,max=32" example:"subuser"`
	Password string `json:"password" binding:"required,min=8,max=64,password" example:"Abcd1234"`
	Email    string `json:"email" example:"sub@example.com"`
	RoleIDs  []uint `json:"role_ids" binding:"required" example:"3"`
}

type UpdateSubAccountReq struct {
	RoleIDs []uint `json:"role_ids" binding:"required" example:"3"`
}

type ChangePwdReq struct {
	OldPassword string `json:"old_password" binding:"required" example:"Abcd1234"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=64,password" example:"NewPass123"`
}

type UpdateMeReq struct {
	Email    string `json:"email" example:"admin@example.com"`
	Nickname string `json:"nickname" example:"管理员"`
	Avatar   string `json:"avatar"`
}

// BatchUserOp is the request body for batch-delete and batch-status endpoints.
type BatchUserOp struct {
	IDs []uint `json:"ids" binding:"required"`
}

// BatchStatusReq is the request body for batch-update-status.
type BatchStatusReq struct {
	IDs    []uint `json:"ids" binding:"required"`
	Status *int   `json:"status" binding:"required" example:"1"`
}

// ---------- Auth handlers ----------

// Register 注册新用户
// @Summary 用户注册
// @Description 使用用户名、密码、邮箱注册新账号；受注册开关控制
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body RegisterReq true "注册请求"
// @Success 200 {object} response.Response{data=domain.User} "注册成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 10101 {object} response.Response "用户已存在"
// @Router /auth/register [post]
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

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名和密码登录，成功后返回 access/refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginReq true "登录请求"
// @Success 200 {object} response.Response{data=LoginResp} "登录成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 10103 {object} response.Response "用户名或密码错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	access, refresh, user, landing, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.auditLogin(c, 0, req.Username, "failure", "")
		response.ErrorJSON(c, errcode.ErrPasswordWrong)
		return
	}
	h.auditLogin(c, user.ID, req.Username, "success", "")
	response.OKJSON(c, LoginResp{AccessToken: access, RefreshToken: refresh, User: *user, LandingPage: landing})
}

// Logout 退出登录
// @Summary 退出登录
// @Description 撤销当前 Bearer token
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "退出成功"
// @Failure 401 {object} response.Response "未认证"
// @Router /auth/logout [post]
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

// Refresh 刷新 token
// @Summary 刷新Token
// @Description 使用 refresh_token 换取新的 access_token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body RefreshReq true "刷新请求"
// @Success 200 {object} response.Response{data=object{access_token=string}} "刷新成功"
// @Failure 10105 {object} response.Response "Token 无效"
// @Router /auth/refresh [post]
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

// ForgotPassword 找回密码
// @Summary 找回密码
// @Description 按邮箱发送密码重置邮件（含一次性 token 链接）。防枚举：邮箱不存在也返回成功
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body ForgotPasswordReq true "找回密码请求"
// @Success 200 {object} response.Response "邮件已发送"
// @Failure 400 {object} response.Response "参数错误"
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		if err.Error() == "password reset disabled" {
			response.BadRequest(c, "找回密码功能未开启")
			return
		}
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 用邮件中的一次性 token 设置新密码，并使旧 session 失效
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body ResetPasswordReq true "重置密码请求"
// @Success 200 {object} response.Response "密码已重置"
// @Failure 400 {object} response.Response "参数错误或 token 无效"
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, nil)
}

// ---------- User CRUD handlers ----------

// Create 创建用户（管理员）
// @Summary 创建用户
// @Description 管理员创建新用户，可指定多个角色（role_ids）
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateUserReq true "创建请求"
// @Success 200 {object} response.Response{data=domain.User} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 10101 {object} response.Response "用户已存在"
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := h.svc.CreateUser(c.Request.Context(), req.Username, req.Password, req.Email, req.RoleIDs, req.DeptIDs)
	if err != nil {
		response.ErrorJSON(c, errcode.ErrUserExists)
		return
	}
	response.OKJSON(c, user)
}

// GetByID 获取用户详情
// @Summary 获取用户
// @Description 根据 ID 获取用户信息
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=domain.User} "用户信息"
// @Failure 404 {object} response.Response "用户不存在"
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	user, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c)
		return
	}
	response.OKJSON(c, user)
}

// GetMe 获取当前用户
// @Summary 获取当前用户信息
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=domain.User} "用户信息"
// @Failure 404 {object} response.Response "用户不存在"
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetUint("user_id")
	user, err := h.svc.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c)
		return
	}
	response.OKJSON(c, user)
}

// UpdateMe 更新个人资料
// @Summary 更新个人资料
// @Description 当前用户更新自己的邮箱、昵称、头像
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body UpdateMeReq true "资料更新请求"
// @Success 200 {object} response.Response{data=domain.User} "更新后的用户信息"
// @Failure 400 {object} response.Response "参数错误"
// @Router /users/me [put]
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
	user, err := h.svc.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, user)
}

// GetMyPermissions 获取当前用户权限
// @Summary 获取当前用户权限列表
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]string} "权限代码列表"
// @Router /users/me/permissions [get]
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

// GetMyLoginLogs 获取当前用户登录日志
// @Summary 我的登录日志
// @Description 查询当前用户的历史登录记录（分页）
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param sort query string false "排序" default(-created_at)
// @Success 200 {object} response.Response{data=response.PageData} "登录日志列表"
// @Router /users/me/login-logs [get]
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
	pq.Filters = append(pq.Filters, query.Condition{Field: "user_id", Op: "eq", Value: userID})
	logs, total, err := h.auditSvc.FindAll(c.Request.Context(), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, logs, total, pq.Page, pq.PageSize)
}

// UploadAvatar 上传头像
// @Summary 上传头像
// @Description 上传当前用户的头像图片（png/jpg/jpeg/gif/webp，最大 2MB）
// @Tags User
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param file formData file true "头像图片"
// @Success 200 {object} response.Response{data=object{url=string}} "头像URL"
// @Failure 400 {object} response.Response "文件类型或大小不符合要求"
// @Router /users/me/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetUint("user_id")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	defer func() { _ = file.Close() }()

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

	// 校验真实内容（magic bytes）
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	if n > 0 {
		detected := http.DetectContentType(buf[:n])
		if !strings.HasPrefix(detected, "image/") {
			response.BadRequest(c, "不支持的文件内容，仅支持图片文件")
			return
		}
	}

	uploadDir := "uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		response.InternalError(c)
		return
	}

	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
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

	avatarURL := "/uploads/avatars/" + filename
	if err := h.svc.UpdateProfile(c.Request.Context(), userID, "", "", avatarURL); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, gin.H{"url": avatarURL})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 当前用户修改自己的登录密码
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ChangePwdReq true "修改密码请求"
// @Success 200 {object} response.Response "修改成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 10103 {object} response.Response "旧密码错误"
// @Router /users/me/password [put]
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

// List 用户列表
// @Summary 用户列表（分页）
// @Description 查询用户列表，支持关键词搜索和字段过滤
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param q query string false "关键词（搜索 username/email）"
// @Param sort query string false "排序" default(-created_at)
// @Success 200 {object} response.Response{data=response.PageData} "用户列表"
// @Router /users [get]
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

// Update 管理员更新用户
// @Summary 更新用户
// @Description 管理员更新指定用户的邮箱、状态、角色（role_ids 全量替换）
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param body body UpdateUserReq true "更新请求"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /users/{id} [put]
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
	if req.RoleIDs != nil {
		// 角色可清空（依赖部门继承角色），但「无直接角色且无部门」才是非法状态
		if len(req.RoleIDs) == 0 && len(req.DeptIDs) == 0 {
			response.BadRequest(c, "至少需要分配一个角色或部门")
			return
		}
		updates["role_ids"] = req.RoleIDs
	}
	if req.DeptIDs != nil {
		updates["dept_ids"] = req.DeptIDs
	}

	// 变更 Diff：更新前读取旧值，构造 旧值→新值 供审计记录
	if len(updates) > 0 {
		if old, err := h.svc.GetByID(c.Request.Context(), uint(id)); err == nil && old != nil {
			diff := map[string]any{}
			if req.Email != nil && old.Email != *req.Email {
				diff["email"] = map[string]any{"old": old.Email, "new": *req.Email}
			}
			if req.Status != nil && old.Status != *req.Status {
				diff["status"] = map[string]any{"old": old.Status, "new": *req.Status}
			}
			if req.RoleIDs != nil && !sameUintSlice(old.DirectRoleIDs, req.RoleIDs) {
				diff["role_ids"] = map[string]any{"old": old.DirectRoleIDs, "new": req.RoleIDs}
			}
			if req.DeptIDs != nil && !sameUintSlice(old.DeptIDs, req.DeptIDs) {
				diff["dept_ids"] = map[string]any{"old": old.DeptIDs, "new": req.DeptIDs}
			}
			if len(diff) > 0 {
				if b, err := json.Marshal(diff); err == nil {
					c.Set("audit_diff", string(b))
				}
			}
		}
	}

	if err := h.svc.Update(c.Request.Context(), uint(id), updates); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// sameUintSlice 判断两个 uint 切片是否相同（忽略顺序用于角色比较）。
func sameUintSlice(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	am := make(map[uint]bool, len(a))
	for _, v := range a {
		am[v] = true
	}
	for _, v := range b {
		if !am[v] {
			return false
		}
	}
	return true
}

// Delete 删除用户
// @Summary 删除用户
// @Description 软删除指定用户（不可删除自己）
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /users/{id} [delete]
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

// BatchDelete 批量删除用户
// @Summary 批量删除用户
// @Description 软删除多个用户
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body BatchUserOp true "批量删除请求"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /users/batch-delete [post]
func (h *UserHandler) BatchDelete(c *gin.Context) {
	var req BatchUserOp
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.BatchDelete(c.Request.Context(), req.IDs); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// BatchUpdateStatus 批量更新用户状态
// @Summary 批量更新用户状态
// @Description 批量启用/禁用用户
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body BatchStatusReq true "批量更新请求"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /users/batch-status [post]
func (h *UserHandler) BatchUpdateStatus(c *gin.Context) {
	var req BatchStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.BatchUpdateStatus(c.Request.Context(), req.IDs, *req.Status); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// RevokeSessions 踢下线
// @Summary 踢下线
// @Description 使指定用户的全部 session 失效
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /users/{id}/sessions [delete]
func (h *UserHandler) RevokeSessions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.RevokeSessions(c.Request.Context(), uint(id)); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

func (h *UserHandler) CreateSubAccount(c *gin.Context) {
	parentID := c.GetUint("user_id")
	var req CreateSubAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := h.svc.CreateSubAccount(c.Request.Context(), parentID, req.Username, req.Password, req.Email, req.RoleIDs)
	if err != nil {
		switch err.Error() {
		case "sub account cannot create sub account":
			response.ErrorJSON(c, errcode.ErrSubAccountIsSub)
		case "sub account limit reached":
			response.ErrorJSON(c, errcode.ErrSubAccountLimit)
		case "user exists":
			response.ErrorJSON(c, errcode.ErrUserExists)
		default:
			if strings.Contains(err.Error(), "not in parent permissions") {
				response.ErrorJSON(c, errcode.ErrSubAccountNoPerm)
			} else {
				response.BadRequest(c, err.Error())
			}
		}
		return
	}
	response.OKJSON(c, user)
}

func (h *UserHandler) ListSubAccounts(c *gin.Context) {
	parentID := c.GetUint("user_id")
	users, err := h.svc.ListSubAccounts(c.Request.Context(), parentID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKJSON(c, users)
}

func (h *UserHandler) UpdateSubAccount(c *gin.Context) {
	parentID := c.GetUint("user_id")
	subID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req UpdateSubAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.UpdateSubAccount(c.Request.Context(), parentID, uint(subID), req.RoleIDs); err != nil {
		if strings.Contains(err.Error(), "not in parent permissions") {
			response.ErrorJSON(c, errcode.ErrSubAccountNoPerm)
		} else if strings.Contains(err.Error(), "not your sub account") {
			response.ErrorJSON(c, errcode.ErrSubAccountNotFound)
		} else {
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.OKJSON(c, nil)
}

func (h *UserHandler) DeleteSubAccount(c *gin.Context) {
	parentID := c.GetUint("user_id")
	subID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteSubAccount(c.Request.Context(), parentID, uint(subID)); err != nil {
		response.ErrorJSON(c, errcode.ErrSubAccountNotFound)
		return
	}
	response.OKJSON(c, nil)
}

func (h *UserHandler) AdminListSubAccounts(c *gin.Context) {
	pq, err := query.Parse(c, userQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	users, total, err := h.svc.AdminListSubAccounts(c.Request.Context(), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, users, total, pq.Page, pq.PageSize)
}

// ---------- helpers ----------

var userQueryDefs = query.Defs{
	"username":       {Name: "username", Type: query.TypeString, Searchable: true, Filterable: true},
	"email":          {Name: "email", Type: query.TypeString, Searchable: true, Filterable: true},
	"status":         {Name: "status", Type: query.TypeInt, Filterable: true},
	"role_id":        {Name: "role_id", Type: query.TypeUint, Filterable: true},
	"department_id":  {Name: "department_id", Type: query.TypeUint, Filterable: true},
	"parent_id":      {Name: "parent_id", Type: query.TypeUint, Filterable: true},
	"is_sub_account": {Name: "is_sub_account", Type: query.TypeBool, Filterable: true},
	"created_at":     {Name: "created_at", Type: query.TypeTime, Filterable: true},
	"updated_at":     {Name: "updated_at", Type: query.TypeTime, Filterable: true},
}

var auditLogQueryDefs = query.Defs{
	"user_id":    {Name: "user_id", Type: query.TypeUint, Filterable: true},
	"username":   {Name: "username", Type: query.TypeString, Searchable: true, Filterable: true},
	"ip":         {Name: "ip", Type: query.TypeString, Searchable: true},
	"user_agent": {Name: "user_agent", Type: query.TypeString, Searchable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

func (h *AuthHandler) auditLogin(c *gin.Context, userID uint, username, result, detail string) {
	if h.auditLog == nil {
		return
	}
	resource := "auth"
	message := result
	if detail != "" {
		message = result + ": " + detail
	}
	h.auditLog(c.Request.Context(), userID, username, "login", resource, result, c.ClientIP(), c.Request.UserAgent())
	_ = message
}

func (h *AuthHandler) auditLogout(c *gin.Context, username string) {
	if h.auditLog == nil {
		return
	}
	h.auditLog(c.Request.Context(), c.GetUint("user_id"), username, "logout", "auth", "success", c.ClientIP(), c.Request.UserAgent())
}
