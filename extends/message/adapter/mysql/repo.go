package adapter

import (
	"context"
	"time"

	"gorm.io/gorm"

	"kingfisher/core/query"
	"kingfisher/extends/message/domain"
)

type MessageRepo struct{ db *gorm.DB }

func NewMessageRepo(db *gorm.DB) *MessageRepo { return &MessageRepo{db: db} }

// ListByRecipient 某收件人的收件箱（固定 recipient_id 防越权，支持分页/过滤/排序）
func (r *MessageRepo) ListByRecipient(ctx context.Context, recipientID uint, q *query.Query) ([]domain.Message, int64, error) {
	var pos []messagePO
	total, err := q.Find(r.db.WithContext(ctx).Model(&messagePO{}).Where("recipient_id = ?", recipientID), &pos)
	if err != nil {
		return nil, 0, err
	}
	return toMessages(pos), total, nil
}

// GetByID 查询单条消息（带 recipient_id 防越权）
func (r *MessageRepo) GetByID(ctx context.Context, id, recipientID uint) (*domain.Message, error) {
	var po messagePO
	if err := r.db.WithContext(ctx).Where("id = ? AND recipient_id = ?", id, recipientID).First(&po).Error; err != nil {
		return nil, err
	}
	return toMessage(&po), nil
}

func (r *MessageRepo) Create(ctx context.Context, m *domain.Message) error {
	po := messagePO{
		SenderID: m.SenderID, SenderType: m.SenderType, RecipientID: m.RecipientID,
		Title: m.Title, Content: m.Content, Status: m.Status, IsRead: m.IsRead,
	}
	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return err
	}
	m.ID = po.ID
	m.CreatedAt = po.CreatedAt
	m.UpdatedAt = po.UpdatedAt
	return nil
}

// MarkRead 标记已读（仅限自己的消息）
func (r *MessageRepo) MarkRead(ctx context.Context, id, recipientID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&messagePO{}).
		Where("id = ? AND recipient_id = ?", id, recipientID).
		Updates(map[string]any{"is_read": true, "read_at": &now}).Error
}

// DeleteBatch 批量删除（仅限自己的消息）
func (r *MessageRepo) DeleteBatch(ctx context.Context, ids []uint, recipientID uint) error {
	return r.db.WithContext(ctx).Where("id IN ? AND recipient_id = ?", ids, recipientID).Delete(&messagePO{}).Error
}

func (r *MessageRepo) CountUnread(ctx context.Context, recipientID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&messagePO{}).
		Where("recipient_id = ? AND is_read = ?", recipientID, false).
		Count(&count).Error
	return count, err
}

// ---- helpers ----

func toMessage(p *messagePO) *domain.Message {
	return &domain.Message{
		ID: p.ID, SenderID: p.SenderID, SenderType: p.SenderType, RecipientID: p.RecipientID,
		Title: p.Title, Content: p.Content, Status: p.Status, IsRead: p.IsRead, ReadAt: p.ReadAt,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func toMessages(pos []messagePO) []domain.Message {
	out := make([]domain.Message, len(pos))
	for i, p := range pos {
		out[i] = *toMessage(&p)
	}
	return out
}
