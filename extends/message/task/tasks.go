// Package task 定义站内信模块的任务类型与载荷（仅属于本站内信模块）。
package task

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

// TypeSendMessage 站内信发送任务。
const TypeSendMessage = "message:send"

// SendMessagePayload 异步发送站内信的任务载荷。
type SendMessagePayload struct {
	SenderID     uint   `json:"sender_id"`
	SenderType   string `json:"sender_type"`
	RecipientIDs []uint `json:"recipient_ids"`
	Title        string `json:"title"`
	Content      string `json:"content"`
}

// defaultRetention 任务完成后在 Redis 中的默认保留时长（7 天），供 asynq 统计/归档查看。
const defaultRetention = 7 * 24 * time.Hour

// NewSendMessageTask 构造发送站内信任务（最多重试 3 次，单次处理超时 30s，结果保留 7 天）。
func NewSendMessageTask(p SendMessagePayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendMessage, payload,
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Second),
		asynq.Retention(defaultRetention),
	), nil
}

// ParseSendMessagePayload 解析发送站内信任务载荷。
func ParseSendMessagePayload(t *asynq.Task) (*SendMessagePayload, error) {
	var p SendMessagePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return nil, err
	}
	return &p, nil
}
