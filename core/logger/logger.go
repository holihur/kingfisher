// Package logger provides structured logging with Zap.

package logger

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config holds logger configuration.
type Config struct {
	Level      string
	Format     string
	Output     string
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

// New creates a Zap logger from config.
func New(cfg Config) (*zap.Logger, error) {
	level := parseLevel(cfg.Level)
	encoder := buildEncoder(cfg.Format)
	writer := buildWriter(cfg.Output, cfg.FilePath, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	core := zapcore.NewCore(encoder, writer, level)
	return zap.New(withMask(core), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

func parseLevel(l string) zapcore.Level {
	switch strings.ToLower(l) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func buildEncoder(format string) zapcore.Encoder {
	ec := zap.NewProductionEncoderConfig()
	ec.TimeKey = "time"
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	if format == "console" {
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return zapcore.NewConsoleEncoder(ec)
	}
	ec.EncodeLevel = zapcore.LowercaseLevelEncoder
	return zapcore.NewJSONEncoder(ec)
}

func buildWriter(output, filePath string, maxSize, maxBackups, maxAge int) zapcore.WriteSyncer {
	if output == "file" && filePath != "" {
		if dir := filepath.Dir(filePath); dir != "." {
			_ = os.MkdirAll(dir, 0755)
		}
		lumber := &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   true,
		}
		return zapcore.AddSync(lumber)
	}
	return zapcore.AddSync(os.Stdout)
}

// maskCore wraps a zapcore.Core and masks sensitive fields.
type maskCore struct{ zapcore.Core }

var sensitiveKeys = map[string]bool{
	"password": true, "token": true, "secret": true, "access_token": true, "refresh_token": true,
}

func (c *maskCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	for i, f := range fields {
		if sensitiveKeys[f.Key] {
			fields[i] = zap.String(f.Key, "***")
		}
	}
	return c.Core.Write(entry, fields)
}

func (c *maskCore) With(fields []zapcore.Field) zapcore.Core {
	for i, f := range fields {
		if sensitiveKeys[f.Key] {
			fields[i] = zap.String(f.Key, "***")
		}
	}
	return &maskCore{Core: c.Core.With(fields)}
}

func (c *maskCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

func withMask(core zapcore.Core) zapcore.Core { return &maskCore{Core: core} }

var globalLogger *zap.Logger

// WithContext extracts trace_id from context for log correlation.
func WithContext(ctx context.Context) *zap.Logger {
	l := Get()
	if l == nil {
		return zap.NewNop()
	}
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return l.With(zap.String("trace_id", traceID))
	}
	return l
}

// ReplaceGlobals sets the global zap logger.
func ReplaceGlobals(logger *zap.Logger) {
	globalLogger = logger
	zap.ReplaceGlobals(logger)
}

// Get returns the global logger.
func Get() *zap.Logger { return globalLogger }
