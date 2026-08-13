package domain

import "time"

// MessageBatch 管理端按批次聚合的一条已发送记录
type MessageBatch struct {
	BatchID        int64  `json:"batch_id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	RecipientCount int    `json:"recipient_count"`
	RecipientNames string `json:"recipient_names"` // 逗号拼接
	Status         string `json:"status"`          // 批次状态：sent | partial | revoked（部分撤回）
	ReadCount      int    `json:"read_count"`
	UnreadCount    int    `json:"unread_count"`
	CreatedAt      string `json:"created_at"`
	LastUpdatedAt  string `json:"updated_at"`
}

// Message 站内信消息
type Message struct {
	ID          uint   `json:"id"`
	SenderID    uint   `json:"sender_id"`
	SenderType  string `json:"sender_type"` // admin | system（system 预留）
	RecipientID uint   `json:"recipient_id"`
	// BatchID 批次号（同一次发送给多个收件人共用；管理端按批次聚合）
	BatchID int64 `json:"batch_id"`
	// RecipientName 收件人用户名（管理端列表关联 users 表填充）
	RecipientName string     `json:"recipient_name,omitempty"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	Status        string     `json:"status"` // draft | sent | revoked（draft 预留）
	IsRead        bool       `json:"is_read"`
	ReadAt        *time.Time `json:"read_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
