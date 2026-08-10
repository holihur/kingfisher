package taskqueue

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// NopTaskType nop 任务类型：什么都不做，仅打印一条日志。
// 用途：周期任务的测试/占位示例，验证调度器 → 入队 → worker 消费整条链路。
const NopTaskType = "nop:run"

// NopWorker 内置的 nop worker：消费 nop:run 任务，仅记录日志。
type NopWorker struct {
	log *zap.Logger
}

// NewNopWorker 创建 nop worker。
func NewNopWorker(log *zap.Logger) *NopWorker {
	return &NopWorker{log: log}
}

// Name 模块名。
func (w *NopWorker) Name() string { return "nop" }

// RegisterWorkers 注册 nop 任务 handler。
func (w *NopWorker) RegisterWorkers(mux *asynq.ServeMux) {
	mux.HandleFunc(NopTaskType, w.HandleNop)
}

// TaskTypes 声明 nop 任务类型（供任务管理页动态加载）。
func (w *NopWorker) TaskTypes() []TaskTypeInfo {
	return []TaskTypeInfo{
		{Type: NopTaskType, Label: "空转任务（测试）", PayloadExample: `{"note":"可携带任意备注，worker 仅记录"}`},
	}
}

// Shutdown worker 无独立资源。
func (w *NopWorker) Shutdown(ctx context.Context) error { return nil }

// HandleNop 处理 nop 任务：不执行业务逻辑，仅打印日志。
func (w *NopWorker) HandleNop(ctx context.Context, t *asynq.Task) error {
	w.log.Info("nop task executed",
		zap.String("type", t.Type()),
		zap.String("payload", string(t.Payload())),
		zap.String("task_id", t.Type()))
	return nil
}
