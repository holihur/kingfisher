package app

import (
	"context"

	"github.com/hibiken/asynq"

	"kingfisher/core/errcode"
	"kingfisher/core/query"
	"kingfisher/extends/task/domain"
	"kingfisher/extends/task/port"
)

// Error 携带 errcode 的错误类型，handler 层据此映射到 HTTP 错误码
type Error struct{ Code int }

func (e *Error) Error() string { return errcode.Msg(e.Code) }

// ScheduledTaskService 周期任务服务
type ScheduledTaskService struct {
	repo port.ScheduledTaskRepository
}

func NewScheduledTaskService(repo port.ScheduledTaskRepository) *ScheduledTaskService {
	return &ScheduledTaskService{repo: repo}
}

func (s *ScheduledTaskService) List(ctx context.Context, q *query.Query) ([]domain.ScheduledTask, int64, error) {
	return s.repo.List(ctx, q)
}

func (s *ScheduledTaskService) GetByID(ctx context.Context, id uint) (*domain.ScheduledTask, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, &Error{Code: errcode.ErrTaskNotFound}
	}
	return t, nil
}

func (s *ScheduledTaskService) Create(ctx context.Context, name, taskType, cronSpec, payload string, enabled int, remark string) (*domain.ScheduledTask, error) {
	return s.repo.Create(ctx, name, taskType, cronSpec, payload, enabled, remark)
}

func (s *ScheduledTaskService) Update(ctx context.Context, id uint, name, taskType, cronSpec, payload string, enabled int, remark string) error {
	return s.repo.Update(ctx, id, name, taskType, cronSpec, payload, enabled, remark)
}

func (s *ScheduledTaskService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *ScheduledTaskService) BatchDelete(ctx context.Context, ids []uint) error {
	return s.repo.DeleteBatch(ctx, ids)
}

func (s *ScheduledTaskService) BatchUpdateStatus(ctx context.Context, ids []uint, enabled int) error {
	return s.repo.UpdateStatusBatch(ctx, ids, enabled)
}

// ListEnabled 返回所有启用的任务（供 provider 拉取）
func (s *ScheduledTaskService) ListEnabled(ctx context.Context) ([]domain.ScheduledTask, error) {
	return s.repo.ListEnabled(ctx)
}

// BuildTask 由任务配置构造一个 asynq 任务（用于手动执行：直接入队一次）。
// 任务类型与载荷与周期性调度一致，入队后由对应 task_type 的 worker 消费。
func (s *ScheduledTaskService) BuildTask(ctx context.Context, id uint) (*asynq.Task, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, &Error{Code: errcode.ErrTaskNotFound}
	}
	if t.TaskType == "" {
		return nil, &Error{Code: errcode.ErrTaskNotFound}
	}
	var payload []byte
	if t.Payload != "" {
		payload = []byte(t.Payload)
	}
	return asynq.NewTask(t.TaskType, payload), nil
}
