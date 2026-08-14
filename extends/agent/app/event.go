// Package app Agent 聊天业务逻辑。
package app

import "encoding/json"

// SSEEvent 聊天流式事件（transport 层序列化为 text/event-stream）。
// 事件类型：start / text_delta / tool_use / tool_result / done / error。
type SSEEvent struct {
	Type string `json:"type"`
	// 各类型负载（按需填充）。
	Delta   string         `json:"delta,omitempty"`   // text_delta 增量文本
	Tool    string         `json:"tool,omitempty"`    // tool_use 工具名
	Input   map[string]any `json:"input,omitempty"`   // tool_use 入参
	Message string         `json:"message,omitempty"` // error 错误信息 / done 摘要
	Code    int            `json:"code,omitempty"`    // error 错误码
	Role    string         `json:"role,omitempty"`    // done 中 assistant 角色
	Content string         `json:"content,omitempty"` // done 中完整回复
}

// String 序列化为 SSE 单行 data。
func (e SSEEvent) String() string {
	b, _ := json.Marshal(e)
	return "data: " + string(b) + "\n\n"
}
