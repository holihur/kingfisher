package app

import (
	"context"

	"kingfisher/core/query"
	"kingfisher/extends/message/domain"
	"kingfisher/extends/message/port"
)

type MessageService struct {
	repo port.MessageRepository
}

func NewMessageService(repo port.MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

func (s *MessageService) List(ctx context.Context, recipientID uint, q *query.Query) ([]domain.Message, int64, error) {
	return s.repo.ListByRecipient(ctx, recipientID, q)
}

// GetByID 查询单条消息（仅限自己的）
func (s *MessageService) GetByID(ctx context.Context, id, recipientID uint) (*domain.Message, error) {
	return s.repo.GetByID(ctx, id, recipientID)
}

func (s *MessageService) Create(ctx context.Context, senderID uint, senderType string, recipientID uint, title, content string) (*domain.Message, error) {
	m := &domain.Message{
		SenderID: senderID, SenderType: senderType, RecipientID: recipientID,
		Title: title, Content: content, Status: "sent", IsRead: false,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MessageService) MarkRead(ctx context.Context, id, recipientID uint) error {
	return s.repo.MarkRead(ctx, id, recipientID)
}

func (s *MessageService) DeleteBatch(ctx context.Context, ids []uint, recipientID uint) error {
	return s.repo.DeleteBatch(ctx, ids, recipientID)
}

func (s *MessageService) UnreadCount(ctx context.Context, recipientID uint) (int64, error) {
	return s.repo.CountUnread(ctx, recipientID)
}
