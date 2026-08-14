// Package transport Agent 模块的 HTTP 层：会话管理接口 + SSE 流式对话。
package transport

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"kingfisher/core/response"
	"kingfisher/extends/agent/app"
)

// AgentHandler Agent HTTP 处理器。
type AgentHandler struct{ svc *app.AgentService }

// NewAgentHandler 创建 AgentHandler。
func NewAgentHandler(svc *app.AgentService) *AgentHandler { return &AgentHandler{svc: svc} }

// ListConversations 当前用户会话列表。
// @Summary 会话列表
// @Description 当前登录用户的 Agent 会话列表（按最近活跃倒序）
// @Tags Agent
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "会话列表"
// @Failure 11001 {object} response.Response "未启用"
// @Router /agent/conversations [get]
func (h *AgentHandler) ListConversations(c *gin.Context) {
	if err := h.svc.CheckEnabled(); err != nil {
		response.ErrorJSON(c, app.ErrorCode(err))
		return
	}
	items, err := h.svc.ListConversations(c.Request.Context(), c.GetUint("user_id"))
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, items)
}

// CreateConversation 创建会话。
// @Summary 创建会话
// @Description 新建一个 Agent 会话（title 可选，留空用首条消息自动命名）
// @Tags Agent
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateConversationReq true "创建请求"
// @Success 200 {object} response.Response "会话"
// @Failure 11001 {object} response.Response "未启用"
// @Router /agent/conversations [post]
func (h *AgentHandler) CreateConversation(c *gin.Context) {
	if err := h.svc.CheckEnabled(); err != nil {
		response.ErrorJSON(c, app.ErrorCode(err))
		return
	}
	var req CreateConversationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	conv, err := h.svc.CreateConversation(c.Request.Context(), c.GetUint("user_id"), req.Title)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, conv)
}

// ListMessages 会话消息历史。
// @Summary 消息历史
// @Description 某会话的消息记录（按时间正序）
// @Tags Agent
// @Produce json
// @Security BearerAuth
// @Param id path int true "会话ID"
// @Success 200 {object} response.Response "消息列表"
// @Failure 11002 {object} response.Response "会话不存在"
// @Router /agent/conversations/{id}/messages [get]
func (h *AgentHandler) ListMessages(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "会话ID非法")
		return
	}
	msgs, err := h.svc.ListMessages(c.Request.Context(), uint(id), c.GetUint("user_id"))
	if err != nil {
		response.ErrorJSON(c, app.ErrorCode(err))
		return
	}
	response.OKJSON(c, msgs)
}

// DeleteConversation 删除会话。
// @Summary 删除会话
// @Description 删除会话及其全部消息
// @Tags Agent
// @Security BearerAuth
// @Param id path int true "会话ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 11002 {object} response.Response "会话不存在"
// @Router /agent/conversations/{id} [delete]
func (h *AgentHandler) DeleteConversation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "会话ID非法")
		return
	}
	if err := h.svc.DeleteConversation(c.Request.Context(), uint(id), c.GetUint("user_id")); err != nil {
		response.ErrorJSON(c, app.ErrorCode(err))
		return
	}
	response.OKJSON(c, nil)
}

// ChatStream 流式对话（SSE）。
// @Summary 流式对话
// @Description 发送一条消息并流式接收回复（text/event-stream）。事件：start/text_delta/tool_use/tool_result/done/error
// @Tags Agent
// @Accept json
// @Produce text/event-stream
// @Security BearerAuth
// @Param body body ChatStreamReq true "对话请求"
// @Router /agent/chat/stream [post]
func (h *AgentHandler) ChatStream(c *gin.Context) {
	if err := h.svc.CheckEnabled(); err != nil {
		response.ErrorJSON(c, app.ErrorCode(err))
		return
	}
	var req ChatStreamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.ConversationID == 0 {
		response.BadRequest(c, "缺少 conversation_id")
		return
	}
	if req.Content == "" {
		response.BadRequest(c, "消息内容不能为空")
		return
	}

	// SSE 头：绕过 gzip（已在 engine 层排除 /api/v1/agent），直接写流。
	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)

	flusher, _ := c.Writer.(http.Flusher)
	emit := func(evt app.SSEEvent) {
		_, _ = c.Writer.WriteString(evt.String())
		if flusher != nil {
			flusher.Flush()
		}
	}

	userToken := c.GetHeader("Authorization")
	err := h.svc.ChatStream(c.Request.Context(), req.ConversationID, c.GetUint("user_id"), req.Content, userToken, emit)
	if err != nil {
		code := app.ErrorCode(err)
		msg := "对话失败"
		var e *app.Error
		if errors.As(err, &e) {
			msg = e.Error()
		}
		if code <= 0 {
			code = 10006
			msg = "服务器内部错误"
		}
		emit(app.SSEEvent{Type: "error", Code: code, Message: msg})
	}
}

// CreateConversationReq 创建会话请求。
type CreateConversationReq struct {
	Title string `json:"title"`
}

// ChatStreamReq 流式对话请求。
type ChatStreamReq struct {
	ConversationID uint   `json:"conversation_id"`
	Content        string `json:"content"`
}
