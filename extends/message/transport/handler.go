package transport

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"kingfisher/core/query"
	"kingfisher/core/response"
	"kingfisher/core/taskqueue"
	"kingfisher/extends/message/app"
	messageTask "kingfisher/extends/message/task"
)

type MessageHandler struct {
	svc      *app.MessageService
	producer taskqueue.Producer
}

func NewMessageHandler(svc *app.MessageService, producer taskqueue.Producer) *MessageHandler {
	return &MessageHandler{svc: svc, producer: producer}
}

// messageQueryDefs 收件箱可查询字段白名单
var messageQueryDefs = query.Defs{
	"title":      {Name: "title", Type: query.TypeString, Searchable: true, Filterable: true},
	"is_read":    {Name: "is_read", Type: query.TypeBool, Filterable: true},
	"status":     {Name: "status", Type: query.TypeString, Filterable: true},
	"created_at": {Name: "created_at", Type: query.TypeTime, Filterable: true},
}

// List 当前用户收件箱
// @Summary 我的收件箱
// @Tags Message
// @Router /me/messages [get]
func (h *MessageHandler) List(c *gin.Context) {
	pq, err := query.Parse(c, messageQueryDefs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	messages, total, err := h.svc.List(c.Request.Context(), c.GetUint("user_id"), pq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.PageJSON(c, messages, total, pq.Page, pq.PageSize)
}

// GetByID 消息详情（仅限自己的）
// @Summary 消息详情
// @Tags Message
// @Router /me/messages/:id [get]
func (h *MessageHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	m, err := h.svc.GetByID(c.Request.Context(), uint(id), c.GetUint("user_id"))
	if err != nil {
		response.NotFound(c)
		return
	}
	response.OKJSON(c, m)
}

// sendReq 管理员发送站内信请求体（recipient_ids 多选；兼容旧的 recipient_id 单发）
type sendReq struct {
	RecipientIDs []uint `json:"recipient_ids"`
	RecipientID  uint   `json:"recipient_id"`
	Title        string `json:"title" binding:"required"`
	Content      string `json:"content"`
}

// Send 管理员发送站内信（支持单个/多个收件人，异步投递）
// @Summary 发送站内信
// @Tags Message
// @Router /messages [post]
func (h *MessageHandler) Send(c *gin.Context) {
	var req sendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 发件人为当前登录的管理员；发送者为"人"（sender_type=admin，与 system 预留区分）
	ids := req.RecipientIDs
	if len(ids) == 0 && req.RecipientID != 0 {
		ids = []uint{req.RecipientID}
	}
	if len(ids) == 0 {
		response.BadRequest(c, "please provide recipient_ids or recipient_id")
		return
	}
	// 异步发送：入队后立即返回，由站内信 worker 消费并批量落库。
	// Redis 是必选项（main 启动时强依赖），producer 始终有效，不做同步降级。
	task, err := messageTask.NewSendMessageTask(messageTask.SendMessagePayload{
		SenderID:     c.GetUint("user_id"),
		SenderType:   "admin",
		RecipientIDs: ids,
		Title:        req.Title,
		Content:      req.Content,
	})
	if err != nil {
		response.InternalError(c)
		return
	}
	if _, err := h.producer.Enqueue(c.Request.Context(), task); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, gin.H{"enqueued": true, "recipients": len(ids)})
}

// MarkRead 标记已读
// @Summary 标记已读
// @Tags Message
// @Router /me/messages/:id/read [put]
func (h *MessageHandler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), uint(id), c.GetUint("user_id")); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// batchIDsReq 批量删除请求体
type batchIDsReq struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// BatchDelete 批量删除
// @Summary 批量删除消息
// @Tags Message
// @Router /me/messages/batch-delete [post]
func (h *MessageHandler) BatchDelete(c *gin.Context) {
	var req batchIDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.DeleteBatch(c.Request.Context(), req.IDs, c.GetUint("user_id")); err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, nil)
}

// UnreadCount 未读消息数
// @Summary 未读消息数
// @Tags Message
// @Router /me/messages/unread-count [get]
func (h *MessageHandler) UnreadCount(c *gin.Context) {
	count, err := h.svc.UnreadCount(c.Request.Context(), c.GetUint("user_id"))
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKJSON(c, gin.H{"unread_count": count})
}
