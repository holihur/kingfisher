// Package config implements config logic.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	TaskQueue TaskQueueConfig `mapstructure:"taskqueue"`
	Telemetry TelemetryConfig `mapstructure:"telemetry"`
	CORS      CORSConfig      `mapstructure:"cors"`
	SMTP      SMTPConfig      `mapstructure:"smtp"`
	Agent     AgentConfig     `mapstructure:"agent"`
}

// AgentConfig Agent 聊天模块配置（LLM 走 anthropic 兼容格式，默认 DeepSeek 端点）
type AgentConfig struct {
	Enabled     bool   `mapstructure:"enabled"`  // 是否启用 agent 聊天
	BaseURL     string `mapstructure:"base_url"` // anthropic 兼容 API 基础地址（如 https://api.deepseek.com/anthropic）
	Model       string `mapstructure:"model"`    // LLM 模型名（如 deepseek-chat）
	MaxTokens   int    `mapstructure:"max_tokens"`
	APIKeyEnv   string `mapstructure:"api_key_env"`   // API key 环境变量名（fallback 到 system_configs.llm_api_key）
	APIKey      string `mapstructure:"api_key"`       // 可选：直接配置 API key（优先级最低）
	SelfBaseURL string `mapstructure:"self_base_url"` // 本服务自身地址（call_api 工具内部请求用，默认 http://127.0.0.1:port）
}

// SMTPConfig 邮件发送配置（找回密码等邮件通知）
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	FromName string `mapstructure:"from_name"`
	// 是否启用邮件发送；未配置时邮件仅记录日志
	Enabled bool `mapstructure:"enabled"`
}

type ServerConfig struct {
	Port           int      `mapstructure:"port"`
	Mode           string   `mapstructure:"mode"`
	ReadTimeout    string   `mapstructure:"read_timeout"`
	WriteTimeout   string   `mapstructure:"write_timeout"`
	MaxRequestBody string   `mapstructure:"max_request_body"`
	TrustedProxies []string `mapstructure:"trusted_proxies"`
	// StaticDir 前端构建产物目录（含 index.html）。非空则单进程同时提供前端 SPA 与 API。
	StaticDir string `mapstructure:"static_dir"`
}

type DatabaseConfig struct {
	Driver   string         `mapstructure:"driver"`
	SQLite   SQLiteConfig   `mapstructure:"sqlite"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type SQLiteConfig struct {
	Path string `mapstructure:"path"`
}

type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
}

type PostgresConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	Issuer        string `mapstructure:"issuer"`
	AccessTTLRaw  string `mapstructure:"access_ttl"`
	RefreshTTLRaw string `mapstructure:"refresh_ttl"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

type TelemetryConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
	LoginPerMinute    int  `mapstructure:"login_per_minute"`
}

// TaskQueueConfig asynq 任务队列配置。Redis 必须可用（main 启动时强依赖）。
type TaskQueueConfig struct {
	Enabled     bool `mapstructure:"enabled"`
	Concurrency int  `mapstructure:"concurrency"`
	// PeriodicSyncInterval 周期任务同步间隔（秒），PeriodicTaskManager 从 DB 拉取配置并同步到调度器的周期。
	PeriodicSyncInterval int `mapstructure:"periodic_sync_interval"`
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           string   `mapstructure:"max_age"`
}

func Load(configPath string) (*Config, error) {
	// 加载项目根 .env（自研解析，零依赖；已存在的环境变量不覆盖）。
	loadDotEnv(".env")
	if dir := filepath.Dir(configPath); dir != "." && dir != "" {
		loadDotEnv(filepath.Join(dir, "..", ".env"))
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	// Load env-specific override (config.{APP_ENV}.yaml)
	if env := os.Getenv("APP_ENV"); env != "" {
		v.SetConfigName("config." + env)
		_ = v.MergeInConfig() // Optional: dev/prod override file may not exist
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Parse duration strings
	if cfg.JWT.AccessTTLRaw != "" {
		d, err := time.ParseDuration(cfg.JWT.AccessTTLRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid jwt.access_ttl: %w", err)
		}
		cfg.JWT.AccessTTL = d
	}
	if cfg.JWT.RefreshTTLRaw != "" {
		d, err := time.ParseDuration(cfg.JWT.RefreshTTLRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid jwt.refresh_ttl: %w", err)
		}
		cfg.JWT.RefreshTTL = d
	}

	// Override JWT secret from env if set
	if sec := v.GetString("JWT_SECRET"); sec != "" {
		cfg.JWT.Secret = sec
	}
	// Override MySQL password from env if set
	if pw := v.GetString("MYSQL_PASSWORD"); pw != "" && cfg.Database.MySQL.Password == "" {
		cfg.Database.MySQL.Password = pw
	}
	// Override PG password
	if pw := v.GetString("PG_PASSWORD"); pw != "" && cfg.Database.Postgres.Password == "" {
		cfg.Database.Postgres.Password = pw
	}
	// Override Redis password
	if pw := v.GetString("REDIS_PASSWORD"); pw != "" && cfg.Redis.Password == "" {
		cfg.Redis.Password = pw
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.Server.Mode != "debug" && c.Server.Mode != "release" && c.Server.Mode != "test" {
		return fmt.Errorf("invalid server mode: %s (must be debug|release|test)", c.Server.Mode)
	}
	// Only check JWT secret for production; dev accepts the default placeholder
	if c.JWT.Secret == "change-me-in-production" && c.Server.Mode == "release" {
		return fmt.Errorf("JWT secret must be set via environment variable JWT_SECRET in release mode")
	}
	// For non-sqlite drivers, require password
	if c.Database.Driver != "sqlite" {
		if c.Database.MySQL.Password == "" && c.Database.Postgres.Password == "" {
			return fmt.Errorf("database password must be set via environment variable")
		}
	}
	return nil
}

// loadDotEnv 解析并加载 .env 文件（每行 KEY=VALUE，# 开头为注释，支持引号）。
// 已存在的环境变量优先，不覆盖。文件不存在时静默忽略。
func loadDotEnv(path string) {
	data, err := os.ReadFile(path) // #nosec G304 -- path 来自固定的 .env 配置路径（启动时由 Load 显式传入）
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(strings.Trim(val, `"'`))
		if key == "" {
			continue
		}
		if os.Getenv(key) != "" {
			continue // 环境变量优先
		}
		_ = os.Setenv(key, val)
	}
}
