package port

import (
	"context"

	"kingfisher/core/query"
	"kingfisher/extends/message/domain"
)

type MessageRepository interface {
	ListByRecipient(ctx context.Context, recipientID uint, q *query.Query) ([]domain.Message, int64, error)
	// ListBySender 管理端：按发送者列出已发送站内信（含撤回状态；撤回的收件箱不显示）
	ListBySender(ctx context.Context, senderID uint, q *query.Query) ([]domain.Message, int64, error)
	// ListSentBatches 管理端：按批次聚合的已发送列表（同一批收件人一条记录 + 收件人数量）
	ListSentBatches(ctx context.Context, senderID uint, q *query.Query) ([]domain.MessageBatch, int64, error)
	GetByID(ctx context.Context, id, recipientID uint) (*domain.Message, error)
	Create(ctx context.Context, m *domain.Message) error
	MarkRead(ctx context.Context, id, recipientID uint) error
	// RevokeBatch 撤回整批：同一 batch_id 全部标记 revoked（收件箱列表过滤后不可见）
	RevokeBatch(ctx context.Context, batchID, senderID uint) error
	// Revoke 撤回单条：仅发送者本人可撤回（batch 详情里的单个收件人）
	Revoke(ctx context.Context, id, senderID uint) error
	// ListBatchMessages 批次详情：该批次下逐收件人的消息（含收件人姓名/状态）
	ListBatchMessages(ctx context.Context, batchID, senderID uint) ([]domain.Message, error)
	DeleteBatch(ctx context.Context, ids []uint, recipientID uint) error
	CountUnread(ctx context.Context, recipientID uint) (int64, error)
}
