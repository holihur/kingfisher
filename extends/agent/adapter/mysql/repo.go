package adapter

import (
	"context"

	"gorm.io/gorm"

	"kingfisher/extends/agent/domain"
)

// AgentRepo 会话与消息的 GORM 实现。
type AgentRepo struct{ db *gorm.DB }

// NewAgentRepo 创建 AgentRepo。
func NewAgentRepo(db *gorm.DB) *AgentRepo { return &AgentRepo{db: db} }

func (r *AgentRepo) ListConversations(ctx context.Context, userID uint) ([]domain.Conversation, error) {
	var pos []conversationPO
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("updated_at DESC").Find(&pos).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Conversation, len(pos))
	for i, p := range pos {
		out[i] = *toConversation(&p)
	}
	return out, nil
}

func (r *AgentRepo) GetConversation(ctx context.Context, id uint) (*domain.Conversation, error) {
	var po conversationPO
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error; err != nil {
		return nil, err
	}
	return toConversation(&po), nil
}

func (r *AgentRepo) CreateConversation(ctx context.Context, userID uint, title string) (*domain.Conversation, error) {
	po := conversationPO{UserID: userID, Title: title}
	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return nil, err
	}
	return toConversation(&po), nil
}

func (r *AgentRepo) RenameConversation(ctx context.Context, id uint, title string) error {
	return r.db.WithContext(ctx).Model(&conversationPO{}).Where("id = ?", id).
		Update("title", title).Error
}

func (r *AgentRepo) DeleteConversation(ctx context.Context, id uint, userID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("conversation_id = ?", id).Delete(&messagePO{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND user_id = ?", id, userID).Delete(&conversationPO{}).Error
	})
}

func (r *AgentRepo) ListMessages(ctx context.Context, conversationID uint) ([]domain.Message, error) {
	var pos []messagePO
	if err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).
		Order("id ASC").Find(&pos).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Message, len(pos))
	for i, p := range pos {
		out[i] = *toMessage(&p)
	}
	return out, nil
}

func (r *AgentRepo) AddMessage(ctx context.Context, m *domain.Message) error {
	po := messagePO{
		ConversationID: m.ConversationID,
		Role:           m.Role,
		Content:        m.Content,
		ToolCalls:      m.ToolCalls,
		ToolResult:     m.ToolResult,
	}
	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return err
	}
	m.ID = po.ID
	m.CreatedAt = po.CreatedAt
	return nil
}

func toConversation(p *conversationPO) *domain.Conversation {
	return &domain.Conversation{
		ID: p.ID, UserID: p.UserID, Title: p.Title,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func toMessage(p *messagePO) *domain.Message {
	return &domain.Message{
		ID: p.ID, ConversationID: p.ConversationID, Role: p.Role,
		Content: p.Content, ToolCalls: p.ToolCalls, ToolResult: p.ToolResult,
		CreatedAt: p.CreatedAt,
	}
}
