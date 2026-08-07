package port

import (
	"context"

	"kingfisher/core/query"
	"kingfisher/extends/message/domain"
)

type MessageRepository interface {
	ListByRecipient(ctx context.Context, recipientID uint, q *query.Query) ([]domain.Message, int64, error)
	GetByID(ctx context.Context, id, recipientID uint) (*domain.Message, error)
	Create(ctx context.Context, m *domain.Message) error
	MarkRead(ctx context.Context, id, recipientID uint) error
	DeleteBatch(ctx context.Context, ids []uint, recipientID uint) error
	CountUnread(ctx context.Context, recipientID uint) (int64, error)
}
