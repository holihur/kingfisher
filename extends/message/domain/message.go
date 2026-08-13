package domain

import "time"

// Message 站内信消息
type Message struct {
	ID          uint   `json:"id"`
	SenderID    uint   `json:"sender_id"`
	SenderType  string `json:"sender_type"` // admin | system（system 预留）
	RecipientID uint   `json:"recipient_id"`
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
