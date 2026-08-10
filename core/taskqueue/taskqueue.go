// Package taskqueue 基于 asynq 的分布式任务队列基础设施。
//
// 设计原则：每个 extends 模块的 worker 独立实现 WorkerModule 接口，
// 通过注册模式（RegisterWorkers）挂载到同一个 asynq server，
// 不把各模块的任务 handler 混在同一个文件里。
package taskqueue

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"

	"kingfisher/core/config"
)

// Producer 任务入队接口，handler 只依赖该接口，便于注入与 mock。
type Producer interface {
	Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

// ClosableProducer 生命周期接口：asynq.Client 需要在进程退出时关闭连接池。
type ClosableProducer interface {
	Close() error
}

type asynqProducer struct {
	client *asynq.Client
}

// NewProducer 基于 Redis 连接配置创建任务生产者（asynq client）。
func NewProducer(opt asynq.RedisClientOpt) Producer {
	return &asynqProducer{client: asynq.NewClient(opt)}
}

func (p *asynqProducer) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return p.client.EnqueueContext(ctx, task, opts...)
}

func (p *asynqProducer) Close() error { return p.client.Close() }

// RedisClientOpt 从应用 Redis 配置构造 asynq 连接参数，与全局 Redis 共用同一实例/库。
func RedisClientOpt(cfg config.RedisConfig) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}
}
