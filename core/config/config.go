// Package config implements config logic.

package config

import (
	"fmt"
	"os"
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
}

type ServerConfig struct {
	Port           int      `mapstructure:"port"`
	Mode           string   `mapstructure:"mode"`
	ReadTimeout    string   `mapstructure:"read_timeout"`
	WriteTimeout   string   `mapstructure:"write_timeout"`
	MaxRequestBody string   `mapstructure:"max_request_body"`
	TrustedProxies []string `mapstructure:"trusted_proxies"`
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
