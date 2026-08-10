// Package task 定义邮件模块的任务类型与载荷。
package task

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

// TypeSendEmail 发送邮件任务。
const TypeSendEmail = "email:send"

// SendEmailPayload 异步发送邮件的任务载荷。
type SendEmailPayload struct {
	To      string `json:"to"`      // 收件人
	Subject string `json:"subject"` // 主题
	Body    string `json:"body"`    // HTML 正文
}

// defaultRetention 任务完成后在 Redis 中的默认保留时长（7 天）。
const defaultRetention = 7 * 24 * time.Hour

// NewSendEmailTask 构造发送邮件任务（最多重试 3 次，单次超时 30s）。
func NewSendEmailTask(p SendEmailPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendEmail, payload,
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Second),
		asynq.Retention(defaultRetention),
	), nil
}

// ParseSendEmailPayload 解析发送邮件任务载荷。
func ParseSendEmailPayload(t *asynq.Task) (*SendEmailPayload, error) {
	var p SendEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return nil, err
	}
	return &p, nil
}
