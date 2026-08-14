// Package domain 定义 Agent 聊天模块的领域实体。
package domain

import "time"

// Conversation 一次聊天会话（归属某个用户）。
type Conversation struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message 会话内的一条消息。
// Role: user | assistant | tool。
// assistant 消息的 ToolCalls 存 JSON 数组 [{id,name,input}]；
// tool 消息的 ToolResult 存 JSON 字符串（call_api 执行结果）。
type Message struct {
	ID             uint      `json:"id"`
	ConversationID uint      `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	ToolCalls      string    `json:"tool_calls,omitempty"`
	ToolResult     string    `json:"tool_result,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
