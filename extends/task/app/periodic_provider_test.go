package app

import (
	"context"
	"testing"
)

func TestPeriodicConfigProviderGetConfigs(t *testing.T) {
	repo := &mockTaskRepo{}
	svc := NewScheduledTaskService(repo)
	ctx := context.Background()

	// 1 个启用 + 1 个禁用
	if _, err := svc.Create(ctx, "启用任务", "message:send", "0 9 * * *", `{"title":"a"}`, 1, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Create(ctx, "禁用任务", "message:send", "0 10 * * *", "", 0, ""); err != nil {
		t.Fatal(err)
	}

	prov := NewPeriodicConfigProvider(svc)
	cfgs, err := prov.GetConfigs()
	if err != nil {
		t.Fatal(err)
	}
	// 只返回启用任务
	if len(cfgs) != 1 {
		t.Fatalf("want 1 config, got %d", len(cfgs))
	}
	c := cfgs[0]
	if c.Cronspec != "0 9 * * *" {
		t.Errorf("cronspec=%s", c.Cronspec)
	}
	if c.Task == nil || c.Task.Type() != "message:send" {
		t.Errorf("task type mismatch: %+v", c.Task)
	}
	// payload 透传
	if string(c.Task.Payload()) != `{"title":"a"}` {
		t.Errorf("payload=%s", c.Task.Payload())
	}
}

func TestPeriodicConfigProviderSkipsEmpty(t *testing.T) {
	repo := &mockTaskRepo{}
	svc := NewScheduledTaskService(repo)
	ctx := context.Background()

	// 空 task_type / 空 cron 的任务应被跳过，不 panic
	if _, err := svc.Create(ctx, "坏任务", "", "", "", 1, ""); err != nil {
		t.Fatal(err)
	}
	prov := NewPeriodicConfigProvider(svc)
	cfgs, err := prov.GetConfigs()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfgs) != 0 {
		t.Errorf("want 0 configs (invalid skipped), got %d", len(cfgs))
	}
}
