package domain

import "time"

// ScheduledTask 周期任务配置（后台管理 → asynq PeriodicTaskManager 周期性同步调度）
type ScheduledTask struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`      // 任务名称
	TaskType  string    `json:"task_type"` // 任务类型（对应 worker 注册的类型，如 message:send）
	CronSpec  string    `json:"cron_spec"` // cron 表达式（5 段，如 0 9 * * *）
	Payload   string    `json:"payload"`   // 任务载荷 JSON
	Enabled   int       `json:"enabled"`   // 1=启用, 0=禁用
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
