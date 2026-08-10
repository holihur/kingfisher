package adapter

import "time"

type scheduledTaskPO struct {
	ID        uint
	Name      string
	TaskType  string
	CronSpec  string
	Payload   string
	Enabled   int
	Remark    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (scheduledTaskPO) TableName() string { return "scheduled_tasks" }
