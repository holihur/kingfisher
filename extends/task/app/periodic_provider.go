package app

import (
	"context"

	"github.com/hibiken/asynq"
)

// PeriodicConfigProvider 周期任务配置提供者：从 DB 拉取启用的任务配置，
// 由 asynq PeriodicTaskManager 周期性同步到调度器（原始数据从 DB 拉取）。
// 实现 core/taskqueue.PeriodicProvider 接口。
type PeriodicConfigProvider struct {
	svc *ScheduledTaskService
}

func NewPeriodicConfigProvider(svc *ScheduledTaskService) *PeriodicConfigProvider {
	return &PeriodicConfigProvider{svc: svc}
}

// GetConfigs 返回所有启用任务的 asynq 周期配置。
// 周期任务按 task_type + payload 构造，入队后由对应 task_type 的 worker 消费。
func (p *PeriodicConfigProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	tasks, err := p.svc.ListEnabled(context.Background())
	if err != nil {
		return nil, err
	}
	cfgs := make([]*asynq.PeriodicTaskConfig, 0, len(tasks))
	for _, t := range tasks {
		if t.TaskType == "" || t.CronSpec == "" {
			continue // 空类型/表达式无法调度，跳过
		}
		var payload []byte
		if t.Payload != "" {
			payload = []byte(t.Payload)
		}
		cfgs = append(cfgs, &asynq.PeriodicTaskConfig{
			Cronspec: t.CronSpec,
			Task:     asynq.NewTask(t.TaskType, payload),
		})
	}
	return cfgs, nil
}
