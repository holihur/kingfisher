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

// ListByRecipient 某收件人的收件箱（固定 recipient_id 防越权；已撤回 revoked 不显示）
func (r *MessageRepo) ListByRecipient(ctx context.Context, recipientID uint, q *query.Query) ([]domain.Message, int64, error) {
	var pos []messagePO
	total, err := q.Find(
		r.db.WithContext(ctx).Model(&messagePO{}).Where("recipient_id = ? AND status <> ?", recipientID, "revoked"),
		&pos,
	)
	if err != nil {
		return nil, 0, err
	}
	return toMessages(pos), total, nil
}

// ListBySender 管理端：按发送者列出已发送站内信（含撤回状态；用于管理列表）
func (r *MessageRepo) ListBySender(ctx context.Context, senderID uint, q *query.Query) ([]domain.Message, int64, error) {
	var pos []messagePO
	total, err := q.Find(
		r.db.WithContext(ctx).Model(&messagePO{}).Where("sender_id = ? AND sender_type = ?", senderID, "admin"),
		&pos,
	)
	if err != nil {
		return nil, 0, err
	}
	msgs := toMessages(pos)
	if err := r.attachRecipientNames(ctx, msgs); err != nil {
		return nil, 0, err
	}
	return msgs, total, nil
}

// attachRecipientNames 关联收件人用户名（管理端列表展示用）
func (r *MessageRepo) attachRecipientNames(ctx context.Context, msgs []domain.Message) error {
	if len(msgs) == 0 {
		return nil
	}
	type row struct {
		ID       uint
		Username string
	}
	var rows []row
	if err := r.db.WithContext(ctx).Table("users").Select("id, username").Scan(&rows).Error; err != nil {
		return err
	}
	names := map[uint]string{}
	for _, rw := range rows {
		names[rw.ID] = rw.Username
	}
	for i := range msgs {
		msgs[i].RecipientName = names[msgs[i].RecipientID]
	}
	return nil
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

// Revoke 撤回（仅发送者本人）：状态改为 revoked，收件箱 ListByRecipient 过滤后不可见
func (r *MessageRepo) Revoke(ctx context.Context, id, senderID uint) error {
	return r.db.WithContext(ctx).Model(&messagePO{}).
		Where("id = ? AND sender_id = ? AND status <> ?", id, senderID, "revoked").
		Update("status", "revoked").Error
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
