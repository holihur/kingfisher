// Package port 定义 Agent 模块需要的仓储接口。
package port

import (
	"context"

	"kingfisher/extends/agent/domain"
)

// AgentRepository 会话与消息的持久化接口。
type AgentRepository interface {
	// ListConversations 当前用户会话列表（按 updated_at 倒序）。
	ListConversations(ctx context.Context, userID uint) ([]domain.Conversation, error)
	// GetConversation 获取会话（调用方负责校验归属）。
	GetConversation(ctx context.Context, id uint) (*domain.Conversation, error)
	// CreateConversation 创建会话并返回带 ID 的实体。
	CreateConversation(ctx context.Context, userID uint, title string) (*domain.Conversation, error)
	// RenameConversation 更新会话标题（首条消息自动命名）。
	RenameConversation(ctx context.Context, id uint, title string) error
	// DeleteConversation 删除会话及其全部消息。
	DeleteConversation(ctx context.Context, id uint, userID uint) error
	// ListMessages 某会话的全部消息（按 id 升序）。
	ListMessages(ctx context.Context, conversationID uint) ([]domain.Message, error)
	// AddMessage 追加一条消息。
	AddMessage(ctx context.Context, m *domain.Message) error
}
