// Package adapter Agent 模块的持久化实现（SQLite 兼容驱动）。
package adapter

import "time"

// conversationPO 会话表（与 core/database.AgentConversationPO 同构，表名一致）。
type conversationPO struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index;not null"`
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (conversationPO) TableName() string { return "agent_conversations" }

// messagePO 消息表（与 core/database.AgentMessagePO 同构，表名一致）。
type messagePO struct {
	ID             uint `gorm:"primaryKey"`
	ConversationID uint `gorm:"index;not null"`
	Role           string
	Content        string `gorm:"type:text"`
	ToolCalls      string `gorm:"type:text"`
	ToolResult     string `gorm:"type:text"`
	CreatedAt      time.Time
}

func (messagePO) TableName() string { return "agent_messages" }
