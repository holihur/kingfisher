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
	GetByID(ctx context.Context, id, recipientID uint) (*domain.Message, error)
	Create(ctx context.Context, m *domain.Message) error
	MarkRead(ctx context.Context, id, recipientID uint) error
	// Revoke 撤回：状态改为 revoked（收件箱列表过滤 revoked 即不可见）
	Revoke(ctx context.Context, id, senderID uint) error
	DeleteBatch(ctx context.Context, ids []uint, recipientID uint) error
	CountUnread(ctx context.Context, recipientID uint) (int64, error)
}
