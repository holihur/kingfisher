package app

import (
	"context"
	"time"

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

// SendBatch 批量发送站内信（同一标题/内容，逐收件人生成独立消息，共用同一 batch_id）
func (s *MessageService) SendBatch(ctx context.Context, senderID uint, senderType string, recipientIDs []uint, title, content string) (int, error) {
	batchID := time.Now().UnixNano()
	for _, rid := range recipientIDs {
		m := &domain.Message{
			SenderID: senderID, SenderType: senderType, RecipientID: rid, BatchID: batchID,
			Title: title, Content: content, Status: "sent", IsRead: false,
		}
		if err := s.repo.Create(ctx, m); err != nil {
			return 0, err
		}
	}
	return len(recipientIDs), nil
}

func (s *MessageService) MarkRead(ctx context.Context, id, recipientID uint) error {
	return s.repo.MarkRead(ctx, id, recipientID)
}

// ListSent 管理端：当前管理员已发送的站内信（按批次聚合）
func (s *MessageService) ListSent(ctx context.Context, senderID uint, q *query.Query) ([]domain.MessageBatch, int64, error) {
	return s.repo.ListSentBatches(ctx, senderID, q)
}

// RevokeBatch 撤回整批：仅发送者本人可撤回自己的站内信批次
func (s *MessageService) RevokeBatch(ctx context.Context, batchID, senderID uint) error {
	return s.repo.RevokeBatch(ctx, batchID, senderID)
}

// Revoke 撤回单条（batch 详情里的单个收件人）
func (s *MessageService) Revoke(ctx context.Context, id, senderID uint) error {
	return s.repo.Revoke(ctx, id, senderID)
}

// ListBatchMessages 批次详情：该批次下逐收件人的消息（含每条的撤回状态）
func (s *MessageService) ListBatchMessages(ctx context.Context, batchID, senderID uint) ([]domain.Message, error) {
	return s.repo.ListBatchMessages(ctx, batchID, senderID)
}

func (s *MessageService) DeleteBatch(ctx context.Context, ids []uint, recipientID uint) error {
	return s.repo.DeleteBatch(ctx, ids, recipientID)
}

func (s *MessageService) UnreadCount(ctx context.Context, recipientID uint) (int64, error) {
	return s.repo.CountUnread(ctx, recipientID)
}
