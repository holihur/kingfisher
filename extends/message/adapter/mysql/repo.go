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

// ListSentBatches 管理端按批次聚合：同一 batch_id 的多条消息合并为一条记录。
// batch_id=0（历史未分组数据）也按单条聚合，保证兼容。
func (r *MessageRepo) ListSentBatches(ctx context.Context, senderID uint, q *query.Query) ([]domain.MessageBatch, int64, error) {
	type batchRow struct {
		BatchID        int64
		Title          string
		Content        string
		RecipientCount int
		RecipientNames string
		Status         string
		RevokedCount   int
		ReadCount      int
		UnreadCount    int
		CreatedAt      string // MIN/MAX 聚合返回字符串，scan 为 time.Time 会失败
		UpdatedAt      string
	}
	var rows []batchRow
	base := r.db.WithContext(ctx).
		Model(&messagePO{}).
		Where("sender_id = ? AND sender_type = ?", senderID, "admin")

	// 先统计总数（聚合前按 sender 过滤）
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 聚合：按 batch_id 分组；batch_id=0 的历史数据不分组（每条自成一批）
	err := base.
		Select(`CASE WHEN batch_id = 0 THEN -id ELSE batch_id END AS batch_id,
			MAX(title) AS title,
			MAX(content) AS content,
			COUNT(*) AS recipient_count,
			GROUP_CONCAT((SELECT username FROM users u WHERE u.id = messages.recipient_id)) AS recipient_names,
			SUM(CASE WHEN status = 'revoked' THEN 1 ELSE 0 END) AS revoked_count,
			SUM(CASE WHEN is_read = 1 THEN 1 ELSE 0 END) AS read_count,
			SUM(CASE WHEN is_read = 0 THEN 1 ELSE 0 END) AS unread_count,
			MIN(created_at) AS created_at,
			MAX(updated_at) AS updated_at`).
		Group("CASE WHEN batch_id = 0 THEN -id ELSE batch_id END").
		Order("created_at DESC").
		Limit(q.PageSize).Offset((q.Page - 1) * q.PageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	batches := make([]domain.MessageBatch, len(rows))
	for i, rw := range rows {
		status := "sent"
		if rw.RevokedCount == rw.RecipientCount {
			status = "revoked"
		} else if rw.RevokedCount > 0 {
			status = "partial" // 部分撤回
		}
		batches[i] = domain.MessageBatch{
			BatchID:        rw.BatchID,
			Title:          rw.Title,
			Content:        rw.Content,
			RecipientCount: rw.RecipientCount,
			RecipientNames: rw.RecipientNames,
			Status:         status,
			ReadCount:      rw.ReadCount,
			UnreadCount:    rw.UnreadCount,
			CreatedAt:      rw.CreatedAt,
			LastUpdatedAt:  rw.UpdatedAt,
		}
	}
	return batches, total, nil
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

// GetByID 查询单条消息（带 recipient_id 防越权；已撤回的对收件人不可见）
func (r *MessageRepo) GetByID(ctx context.Context, id, recipientID uint) (*domain.Message, error) {
	var po messagePO
	if err := r.db.WithContext(ctx).
		Where("id = ? AND recipient_id = ? AND status <> ?", id, recipientID, "revoked").
		First(&po).Error; err != nil {
		return nil, err
	}
	return toMessage(&po), nil
}

func (r *MessageRepo) Create(ctx context.Context, m *domain.Message) error {
	po := messagePO{
		SenderID: m.SenderID, SenderType: m.SenderType, RecipientID: m.RecipientID,
		BatchID: m.BatchID, Title: m.Title, Content: m.Content, Status: m.Status, IsRead: m.IsRead,
	}
	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return err
	}
	m.ID = po.ID
	m.BatchID = po.BatchID
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

// RevokeBatch 撤回整批（仅发送者本人）：同一 batch_id 全部标记 revoked，
// 收件箱 ListByRecipient 过滤后不可见
func (r *MessageRepo) RevokeBatch(ctx context.Context, batchID, senderID uint) error {
	return r.db.WithContext(ctx).Model(&messagePO{}).
		Where("batch_id = ? AND sender_id = ? AND status <> ?", batchID, senderID, "revoked").
		Update("status", "revoked").Error
}

// Revoke 撤回单条（仅发送者本人）：batch 详情里对单个收件人撤回
func (r *MessageRepo) Revoke(ctx context.Context, id, senderID uint) error {
	return r.db.WithContext(ctx).Model(&messagePO{}).
		Where("id = ? AND sender_id = ? AND status <> ?", id, senderID, "revoked").
		Update("status", "revoked").Error
}

// ListBatchMessages 批次详情：该批次下逐收件人的消息（附收件人姓名）
func (r *MessageRepo) ListBatchMessages(ctx context.Context, batchID, senderID uint) ([]domain.Message, error) {
	var pos []messagePO
	if err := r.db.WithContext(ctx).Model(&messagePO{}).
		Where("batch_id = ? AND sender_id = ?", batchID, senderID).
		Order("id ASC").
		Find(&pos).Error; err != nil {
		return nil, err
	}
	msgs := toMessages(pos)
	if err := r.attachRecipientNames(ctx, msgs); err != nil {
		return nil, err
	}
	return msgs, nil
}

// DeleteBatch 批量删除（仅限自己的消息）
func (r *MessageRepo) DeleteBatch(ctx context.Context, ids []uint, recipientID uint) error {
	return r.db.WithContext(ctx).Where("id IN ? AND recipient_id = ?", ids, recipientID).Delete(&messagePO{}).Error
}

func (r *MessageRepo) CountUnread(ctx context.Context, recipientID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&messagePO{}).
		// 已撤回的消息不计入未读数（与收件箱列表 ListByRecipient 过滤一致）
		Where("recipient_id = ? AND is_read = ? AND status <> ?", recipientID, false, "revoked").
		Count(&count).Error
	return count, err
}

// ---- helpers ----

func toMessage(p *messagePO) *domain.Message {
	return &domain.Message{
		ID: p.ID, SenderID: p.SenderID, SenderType: p.SenderType, RecipientID: p.RecipientID,
		BatchID: p.BatchID,
		Title:   p.Title, Content: p.Content, Status: p.Status, IsRead: p.IsRead, ReadAt: p.ReadAt,
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
