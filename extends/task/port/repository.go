package port

import (
	"context"

	"kingfisher/core/query"
	"kingfisher/extends/task/domain"
)

// ScheduledTaskRepository 周期任务仓库接口
type ScheduledTaskRepository interface {
	List(ctx context.Context, q *query.Query) ([]domain.ScheduledTask, int64, error)
	GetByID(ctx context.Context, id uint) (*domain.ScheduledTask, error)
	// ListEnabled 返回所有启用的任务（供 PeriodicTaskConfigProvider 拉取）
	ListEnabled(ctx context.Context) ([]domain.ScheduledTask, error)
	Create(ctx context.Context, name, taskType, cronSpec, payload string, enabled int, remark string) (*domain.ScheduledTask, error)
	Update(ctx context.Context, id uint, name, taskType, cronSpec, payload string, enabled int, remark string) error
	Delete(ctx context.Context, id uint) error
	DeleteBatch(ctx context.Context, ids []uint) error
	UpdateStatusBatch(ctx context.Context, ids []uint, enabled int) error
}
