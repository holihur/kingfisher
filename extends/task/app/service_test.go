package app

import (
	"context"
	"errors"
	"testing"

	"kingfisher/core/query"
	"kingfisher/extends/task/domain"
)

// mockTaskRepo 内存实现 port.ScheduledTaskRepository
type mockTaskRepo struct {
	tasks map[uint]*domain.ScheduledTask
	seq   uint
}

func (m *mockTaskRepo) List(ctx context.Context, q *query.Query) ([]domain.ScheduledTask, int64, error) {
	var out []domain.ScheduledTask
	for _, v := range m.tasks {
		out = append(out, *v)
	}
	return out, int64(len(out)), nil
}

func (m *mockTaskRepo) GetByID(ctx context.Context, id uint) (*domain.ScheduledTask, error) {
	v, ok := m.tasks[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

func (m *mockTaskRepo) ListEnabled(ctx context.Context) ([]domain.ScheduledTask, error) {
	var out []domain.ScheduledTask
	for _, v := range m.tasks {
		if v.Enabled == 1 {
			out = append(out, *v)
		}
	}
	return out, nil
}

func (m *mockTaskRepo) Create(ctx context.Context, name, taskType, cronSpec, payload string, enabled int, remark string) (*domain.ScheduledTask, error) {
	m.seq++
	t := &domain.ScheduledTask{
		ID: m.seq, Name: name, TaskType: taskType, CronSpec: cronSpec,
		Payload: payload, Enabled: enabled, Remark: remark,
	}
	if m.tasks == nil {
		m.tasks = map[uint]*domain.ScheduledTask{}
	}
	m.tasks[t.ID] = t
	return t, nil
}

func (m *mockTaskRepo) Update(ctx context.Context, id uint, name, taskType, cronSpec, payload string, enabled int, remark string) error {
	v, ok := m.tasks[id]
	if !ok {
		return errors.New("not found")
	}
	v.Name, v.TaskType, v.CronSpec, v.Payload, v.Enabled, v.Remark = name, taskType, cronSpec, payload, enabled, remark
	return nil
}

func (m *mockTaskRepo) Delete(ctx context.Context, id uint) error {
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		delete(m.tasks, id)
	}
	return nil
}

func (m *mockTaskRepo) UpdateStatusBatch(ctx context.Context, ids []uint, enabled int) error {
	for _, id := range ids {
		if v, ok := m.tasks[id]; ok {
			v.Enabled = enabled
		}
	}
	return nil
}

func TestScheduledTaskCRUD(t *testing.T) {
	svc := NewScheduledTaskService(&mockTaskRepo{})
	ctx := context.Background()

	t1, err := svc.Create(ctx, "每日通知", "message:send", "0 9 * * *", `{"title":"x"}`, 1, "")
	if err != nil || t1.ID == 0 || t1.CronSpec != "0 9 * * *" || t1.Enabled != 1 {
		t.Fatalf("create: err=%v %+v", err, t1)
	}

	// GetByID + 不存在
	if _, err := svc.GetByID(ctx, 999); err == nil {
		t.Error("want error for missing id")
	}

	// Update
	if err := svc.Update(ctx, t1.ID, "每日通知V2", "message:send", "0 10 * * *", "", 0, "r"); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.GetByID(ctx, t1.ID)
	if got.CronSpec != "0 10 * * *" || got.Enabled != 0 {
		t.Errorf("update not applied: %+v", got)
	}

	// BatchUpdateStatus
	if err := svc.BatchUpdateStatus(ctx, []uint{t1.ID}, 1); err != nil {
		t.Fatal(err)
	}
	got, _ = svc.GetByID(ctx, t1.ID)
	if got.Enabled != 1 {
		t.Error("batch status not applied")
	}

	// ListEnabled
	enabled, err := svc.ListEnabled(ctx)
	if err != nil || len(enabled) != 1 {
		t.Errorf("list enabled: err=%v n=%d", err, len(enabled))
	}

	// Delete
	if err := svc.Delete(ctx, t1.ID); err != nil {
		t.Fatal(err)
	}
	if _, total, _ := svc.List(ctx, &query.Query{Page: 1, PageSize: 20}); total != 0 {
		t.Errorf("want 0 after delete, got %d", total)
	}
}

func TestScheduledTaskBuildTask(t *testing.T) {
	svc := NewScheduledTaskService(&mockTaskRepo{})
	ctx := context.Background()

	t1, err := svc.Create(ctx, "每日通知", "message:send", "0 9 * * *", `{"title":"x"}`, 1, "")
	if err != nil {
		t.Fatal(err)
	}

	// 正常构造：task 类型 + payload 透传
	task, err := svc.BuildTask(ctx, t1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if task.Type() != "message:send" {
		t.Errorf("type=%s", task.Type())
	}
	if string(task.Payload()) != `{"title":"x"}` {
		t.Errorf("payload=%s", task.Payload())
	}

	// 不存在的 id → ErrTaskNotFound
	if _, err := svc.BuildTask(ctx, 999); err == nil {
		t.Error("want error for missing task")
	}
}
