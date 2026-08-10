package domain

import "time"

type AuditLog struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID uint      `json:"resource_id"`
	// Detail 操作详情（JSON 字符串，含关键字段/变更信息）
	Detail string `json:"detail"`
	// Result 操作结果：success | failure
	Result string `json:"result"`
	// Latency 操作耗时（毫秒）
	Latency int64 `json:"latency"`
	// Message 结果说明/失败原因
	Message   string    `json:"message"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}
