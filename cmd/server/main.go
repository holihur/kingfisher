// @title           Kingfisher Admin API
// @version         1.0
// @description     后台管理系统 API 文档
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"kingfisher/cmd/server/app"
	"kingfisher/core/config"
	"kingfisher/core/logger"
	"kingfisher/core/middleware"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	configPath := "config/config.yaml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		configPath = p
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load failed: %v\n", err)
		os.Exit(1)
	}

	middleware.InitValidator()
	zapLog, err := logger.New(logger.Config{
		Level: cfg.Log.Level, Format: cfg.Log.Format, Output: cfg.Log.Output,
		FilePath: cfg.Log.FilePath, MaxSize: cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups, MaxAge: cfg.Log.MaxAge,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger init failed: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = zapLog.Sync() }()
	logger.ReplaceGlobals(zapLog)
	zapLog.Info("config loaded", zap.String("mode", cfg.Server.Mode))

	vi := app.VersionInfo{Version: version, Commit: commit, BuildTime: buildTime}
	if err := app.Run(cfg, zapLog, vi); err != nil {
		zapLog.Fatal("app run failed", zap.Error(err))
	}
}
