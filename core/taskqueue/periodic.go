// Package taskqueue 的周期性调度支持：包装 asynq.PeriodicTaskManager。
// 每个 extends 模块可独立实现 PeriodicProvider 接口提供自己的周期任务配置，
// 由 PeriodicManager 统一注册到 asynq 调度器（注册模式，与 WorkerModule 一致）。
package taskqueue

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"kingfisher/core/config"
)

// PeriodicProvider 周期任务配置提供者：每个模块独立实现。
// 返回的配置会被 PeriodicManager 周期性同步到 asynq 调度器（原始数据从 DB 拉取）。
type PeriodicProvider interface {
	GetConfigs() ([]*asynq.PeriodicTaskConfig, error)
}

// PeriodicProviderProvider 可选接口：模块 transport 层实现它，
// 返回本模块独立的周期任务 provider。主程序通过类型断言收集。
type PeriodicProviderProvider interface {
	PeriodicProvider() PeriodicProvider
}

// PeriodicManager 包装 asynq.PeriodicTaskManager：汇聚各模块周期 provider。
type PeriodicManager struct {
	mgr *asynq.PeriodicTaskManager
	log *zap.Logger
}

// NewPeriodicManager 创建周期性调度器。
// providers 为各模块周期任务配置提供者的集合；cfg.PeriodicSyncInterval 为同步间隔（秒，<=0 用 asynq 默认 3m）。
func NewPeriodicManager(opt asynq.RedisClientOpt, cfg config.TaskQueueConfig, providers []PeriodicProvider, log *zap.Logger) (*PeriodicManager, error) {
	opts := asynq.PeriodicTaskManagerOpts{
		RedisConnOpt:               opt,
		PeriodicTaskConfigProvider: &multiProvider{providers: providers},
	}
	if cfg.PeriodicSyncInterval > 0 {
		opts.SyncInterval = time.Duration(cfg.PeriodicSyncInterval) * time.Second
	}
	mgr, err := asynq.NewPeriodicTaskManager(opts)
	if err != nil {
		return nil, err
	}
	return &PeriodicManager{mgr: mgr, log: log}, nil
}

// Start 启动周期性调度：首次从各 provider 拉取配置并注册到 asynq 调度器，
// 之后按 SyncInterval 周期性同步（DB 配置变更在下一个同步周期自动生效）。
func (m *PeriodicManager) Start() error {
	if err := m.mgr.Start(); err != nil {
		return err
	}
	m.log.Info("periodic task manager started")
	return nil
}

// Shutdown 停止周期性调度（asynq 的 Shutdown 无返回值，阻塞至调度器停止）。
func (m *PeriodicManager) Shutdown(ctx context.Context) error {
	m.mgr.Shutdown()
	return nil
}

// multiProvider 聚合多个模块 provider 的配置（注册模式：不混在一个文件里）。
type multiProvider struct {
	providers []PeriodicProvider
}

func (p *multiProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	var all []*asynq.PeriodicTaskConfig
	for _, prov := range p.providers {
		cfgs, err := prov.GetConfigs()
		if err != nil {
			return nil, err
		}
		all = append(all, cfgs...)
	}
	return all, nil
}
