package taskqueue

import (
	"go.uber.org/zap"
)

// asynqZapLogger 把 zap.Logger 适配为 asynq.Logger 接口。
type asynqZapLogger struct {
	log *zap.Logger
}

func (l *asynqZapLogger) Debug(args ...interface{}) { l.log.Sugar().Debug(args...) }
func (l *asynqZapLogger) Info(args ...interface{})  { l.log.Sugar().Info(args...) }
func (l *asynqZapLogger) Warn(args ...interface{})  { l.log.Sugar().Warn(args...) }
func (l *asynqZapLogger) Error(args ...interface{}) { l.log.Sugar().Error(args...) }
func (l *asynqZapLogger) Fatal(args ...interface{}) { l.log.Sugar().Fatal(args...) }
