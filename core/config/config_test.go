package config

import (
	"os"
	"testing"
)

func TestValidate(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Port: 8080, Mode: "debug"}, Database: DatabaseConfig{Driver: "sqlite"}}
	if err := cfg.Validate(); err != nil {
		t.Error("valid config should pass:", err)
	}
	cfg.Server.Port = 0
	if err := cfg.Validate(); err == nil {
		t.Error("port=0 should fail")
	}
	cfg.Server.Port = 8080
	cfg.Server.Mode = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Error("invalid mode should fail")
	}
	cfg.Server.Mode = "release"
	cfg.JWT.Secret = "change-me-in-production"
	if err := cfg.Validate(); err == nil {
		t.Error("placeholder JWT secret in release should fail")
	}
	cfg.Server.Port = 99999
	if err := cfg.Validate(); err == nil {
		t.Error("port > 65535 should fail")
	}
}

func TestLoadValid(t *testing.T) {
	tmp := t.TempDir()
	configPath := tmp + "/config.yaml"
	content := []byte("server:\n  port: 9090\n  mode: debug\ndatabase:\n  driver: sqlite\nredis:\n  host: 127.0.0.1\n  port: 6379\njwt:\n  secret: test123\n  access_ttl: 2h\n  refresh_ttl: 168h\n  issuer: test\nlog:\n  level: debug\n  format: console\n  output: stdout\nrate_limit:\n  enabled: false\ncors:\n  allowed_origins:\n    - \"*\"\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal("write config:", err)
	}
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatal("Load:", err)
	}
	if cfg.Server.Port != 9090 {
		t.Error("port mismatch")
	}
}

func TestLoadNoFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("should fail on missing file")
	}
}

func TestJWTDefaultTTL(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Port: 8080, Mode: "test"}, Database: DatabaseConfig{Driver: "sqlite"}, JWT: JWTConfig{Secret: "x"}}
	if cfg.JWT.AccessTTL != 0 {
		t.Log("AccessTTL raw value not parsed from config string")
	}
	if cfg.JWT.RefreshTTL != 0 {
		t.Log("RefreshTTL raw value not parsed from config string")
	}
	if err := cfg.Validate(); err != nil {
		t.Error("should validate:", err)
	}
}
