package taskqueue

import (
	"context"

	"github.com/hibiken/asynq"
)

// WorkerModule 每个 extends 模块的 worker 独立实现该接口。
// 各模块在自己的 worker 子包中定义任务类型与处理函数，
// 通过 RegisterWorkers 把自己的 handler 注册到统一 ServeMux 上（注册模式）。
type WorkerModule interface {
	// Name 模块名，用于日志与监控标识。
	Name() string
	// RegisterWorkers 把本模块的任务 handler 注册到 mux。
	RegisterWorkers(mux *asynq.ServeMux)
	// Shutdown 在服务优雅退出时调用，用于 worker 资源清理。
	Shutdown(ctx context.Context) error
}

// WorkerProvider 可选接口：extends 模块的 transport 层实现它，
// 返回本模块独立的 worker。主程序通过类型断言收集所有 worker，
// 无需在主程序里手写各模块的 handler 清单。
type WorkerProvider interface {
	Worker() WorkerModule
}
