package taskqueue

import (
	"context"

	"github.com/hibiken/asynq"
)

// TaskTypeInfo 一个可被周期任务调用的任务类型（由各模块 worker 声明）。
type TaskTypeInfo struct {
	// Type 任务类型标识，如 message:send（对应 worker 注册的类型）
	Type string `json:"type"`
	// Label 任务类型中文名，用于前端展示
	Label string `json:"label"`
	// PayloadExample 载荷 JSON 示例，帮助前端配置 payload
	PayloadExample string `json:"payload_example,omitempty"`
}

// WorkerModule 每个 extends 模块的 worker 独立实现该接口。
// 各模块在自己的 worker 子包中定义任务类型与处理函数，
// 通过 RegisterWorkers 把自己的 handler 注册到统一 ServeMux 上（注册模式）。
type WorkerModule interface {
	// Name 模块名，用于日志与监控标识。
	Name() string
	// RegisterWorkers 把本模块的任务 handler 注册到 mux。
	RegisterWorkers(mux *asynq.ServeMux)
	// TaskTypes 声明本模块可被周期任务调用的任务类型（用于任务管理页动态加载）。
	TaskTypes() []TaskTypeInfo
	// Shutdown 在服务优雅退出时调用，用于 worker 资源清理。
	Shutdown(ctx context.Context) error
}

// WorkerProvider 可选接口：extends 模块的 transport 层实现它，
// 返回本模块独立的 worker。主程序通过类型断言收集所有 worker，
// 无需在主程序里手写各模块的 handler 清单。
type WorkerProvider interface {
	Worker() WorkerModule
}
