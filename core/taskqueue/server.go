package taskqueue

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"kingfisher/core/config"
)

// Server 包装 asynq.Server：收集各模块 worker 并注册到同一 ServeMux。
type Server struct {
	srv     *asynq.Server
	workers []WorkerModule
	log     *zap.Logger
}

// NewServer 创建任务队列服务端。workers 为各模块独立 worker 的集合。
func NewServer(opt asynq.RedisClientOpt, cfg config.TaskQueueConfig, workers []WorkerModule, log *zap.Logger) *Server {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 10
	}
	srv := asynq.NewServer(opt, asynq.Config{
		Concurrency: cfg.Concurrency,
		Logger:      &asynqZapLogger{log},
	})
	return &Server{srv: srv, workers: workers, log: log}
}

// Start 让各模块 worker 注册 handler，并启动异步消费 goroutine。
func (s *Server) Start() error {
	mux := asynq.NewServeMux()
	for _, w := range s.workers {
		w.RegisterWorkers(mux)
		s.log.Info("worker registered", zap.String("module", w.Name()))
	}
	go func() {
		if err := s.srv.Run(mux); err != nil {
			s.log.Error("taskqueue server stopped", zap.Error(err))
		}
	}()
	return nil
}

// Shutdown 停止消费并等待正在处理的任务完成。
func (s *Server) Shutdown(ctx context.Context) error {
	s.srv.Shutdown() // asynq 的 Shutdown 无返回值，阻塞至 ShutdownTimeout
	for _, w := range s.workers {
		if err := w.Shutdown(ctx); err != nil {
			s.log.Error("worker shutdown error", zap.String("module", w.Name()), zap.Error(err))
		}
	}
	return nil
}
