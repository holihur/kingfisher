package adapter

import (
	"context"

	"gorm.io/gorm"

	"kingfisher/core/query"
	"kingfisher/extends/task/domain"
)

type ScheduledTaskRepo struct{ db *gorm.DB }

func NewScheduledTaskRepo(db *gorm.DB) *ScheduledTaskRepo { return &ScheduledTaskRepo{db: db} }

func (r *ScheduledTaskRepo) List(ctx context.Context, q *query.Query) ([]domain.ScheduledTask, int64, error) {
	var pos []scheduledTaskPO
	total, err := q.Find(r.db.WithContext(ctx).Model(&scheduledTaskPO{}), &pos)
	if err != nil {
		return nil, 0, err
	}
	return toScheduledTaskList(pos), total, nil
}

func (r *ScheduledTaskRepo) GetByID(ctx context.Context, id uint) (*domain.ScheduledTask, error) {
	var po scheduledTaskPO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
	if err != nil {
		return nil, err
	}
	return toScheduledTask(&po), nil
}

// ListEnabled 返回所有启用任务（供 PeriodicTaskConfigProvider 周期性拉取）
func (r *ScheduledTaskRepo) ListEnabled(ctx context.Context) ([]domain.ScheduledTask, error) {
	var pos []scheduledTaskPO
	err := r.db.WithContext(ctx).Where("enabled = 1").Find(&pos).Error
	if err != nil {
		return nil, err
	}
	return toScheduledTaskList(pos), nil
}

func (r *ScheduledTaskRepo) Create(ctx context.Context, name, taskType, cronSpec, payload string, enabled int, remark string) (*domain.ScheduledTask, error) {
	po := scheduledTaskPO{
		Name: name, TaskType: taskType, CronSpec: cronSpec,
		Payload: payload, Enabled: enabled, Remark: remark,
	}
	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return nil, err
	}
	return toScheduledTask(&po), nil
}

func (r *ScheduledTaskRepo) Update(ctx context.Context, id uint, name, taskType, cronSpec, payload string, enabled int, remark string) error {
	return r.db.WithContext(ctx).Model(&scheduledTaskPO{}).Where("id = ?", id).Updates(map[string]any{
		"name":      name,
		"task_type": taskType,
		"cron_spec": cronSpec,
		"payload":   payload,
		"enabled":   enabled,
		"remark":    remark,
	}).Error
}

func (r *ScheduledTaskRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&scheduledTaskPO{}).Error
}

func (r *ScheduledTaskRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&scheduledTaskPO{}).Error
}

func (r *ScheduledTaskRepo) UpdateStatusBatch(ctx context.Context, ids []uint, enabled int) error {
	return r.db.WithContext(ctx).Model(&scheduledTaskPO{}).Where("id IN ?", ids).Update("enabled", enabled).Error
}

func toScheduledTask(p *scheduledTaskPO) *domain.ScheduledTask {
	return &domain.ScheduledTask{
		ID:        p.ID,
		Name:      p.Name,
		TaskType:  p.TaskType,
		CronSpec:  p.CronSpec,
		Payload:   p.Payload,
		Enabled:   p.Enabled,
		Remark:    p.Remark,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toScheduledTaskList(pos []scheduledTaskPO) []domain.ScheduledTask {
	out := make([]domain.ScheduledTask, len(pos))
	for i, p := range pos {
		out[i] = *toScheduledTask(&p)
	}
	return out
}
