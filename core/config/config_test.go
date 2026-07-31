package config

import "testing"

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
}
