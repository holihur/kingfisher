package app

import (
	"context"
	"fmt"
	"testing"

	"kingfisher/extends/config/domain"
)

type mockConfigRepo struct {
	configs map[string]*domain.SystemConfig
}

func (m *mockConfigRepo) GetAll(ctx context.Context) ([]domain.SystemConfig, error) {
	var out []domain.SystemConfig
	for _, c := range m.configs {
		out = append(out, *c)
	}
	return out, nil
}

func (m *mockConfigRepo) GetByKey(ctx context.Context, key string) (*domain.SystemConfig, error) {
	c, ok := m.configs[key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockConfigRepo) Set(ctx context.Context, key, value string) error {
	if m.configs == nil {
		m.configs = map[string]*domain.SystemConfig{}
	}
	m.configs[key] = &domain.SystemConfig{Key: key, Value: value}
	return nil
}

func (m *mockConfigRepo) Delete(ctx context.Context, key string) error {
	delete(m.configs, key)
	return nil
}

func TestConfigGetAll(t *testing.T) {
	repo := &mockConfigRepo{
		configs: map[string]*domain.SystemConfig{
			"site_name":          {Key: "site_name", Value: "Kingfisher Admin"},
			"max_login_attempts": {Key: "max_login_attempts", Value: "5"},
		},
	}
	svc := NewConfigService(repo, nil)
	cfgs, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatal("get all:", err)
	}
	if len(cfgs) != 2 {
		t.Errorf("want 2 configs, got %d", len(cfgs))
	}
}

func TestConfigGet(t *testing.T) {
	repo := &mockConfigRepo{
		configs: map[string]*domain.SystemConfig{
			"site_name": {Key: "site_name", Value: "Kingfisher Admin"},
		},
	}
	svc := NewConfigService(repo, nil)
	cfg, err := svc.Get(context.Background(), "site_name")
	if err != nil {
		t.Fatal("get:", err)
	}
	if cfg.Value != "Kingfisher Admin" {
		t.Error("value mismatch")
	}
}

func TestConfigGetNotFound(t *testing.T) {
	repo := &mockConfigRepo{configs: map[string]*domain.SystemConfig{}}
	svc := NewConfigService(repo, nil)
	_, err := svc.Get(context.Background(), "notexist")
	if err == nil {
		t.Error("should return error")
	}
}

func TestConfigSet(t *testing.T) {
	repo := &mockConfigRepo{configs: map[string]*domain.SystemConfig{}}
	svc := NewConfigService(repo, nil)
	if err := svc.Set(context.Background(), "key1", "val1"); err != nil {
		t.Fatal("set:", err)
	}
	cfg, _ := svc.Get(context.Background(), "key1")
	if cfg.Value != "val1" {
		t.Error("value mismatch")
	}
}

func TestConfigSetUpdate(t *testing.T) {
	repo := &mockConfigRepo{
		configs: map[string]*domain.SystemConfig{
			"site_name": {Key: "site_name", Value: "old"},
		},
	}
	svc := NewConfigService(repo, nil)
	if err := svc.Set(context.Background(), "site_name", "new"); err != nil {
		t.Fatal("set:", err)
	}
	cfg, _ := svc.Get(context.Background(), "site_name")
	if cfg.Value != "new" {
		t.Error("value should be updated")
	}
}

func TestConfigDelete(t *testing.T) {
	repo := &mockConfigRepo{
		configs: map[string]*domain.SystemConfig{
			"site_name": {Key: "site_name", Value: "x"},
		},
	}
	svc := NewConfigService(repo, nil)
	if err := svc.Delete(context.Background(), "site_name"); err != nil {
		t.Fatal("delete:", err)
	}
	if len(repo.configs) != 0 {
		t.Error("should be empty")
	}
}
