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

func (m *mockConfigRepo) GetPublicAll(ctx context.Context) ([]domain.SystemConfig, error) {
	var out []domain.SystemConfig
	for _, c := range m.configs {
		if c.IsPublic {
			out = append(out, *c)
		}
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

func (m *mockConfigRepo) GetPublicByKey(ctx context.Context, key string) (*domain.SystemConfig, error) {
	c, ok := m.configs[key]
	if !ok || !c.IsPublic {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockConfigRepo) Set(ctx context.Context, key, value string, isPublic bool, version, render, renderOptions string) error {
	if m.configs == nil {
		m.configs = map[string]*domain.SystemConfig{}
	}
	m.configs[key] = &domain.SystemConfig{Key: key, Value: value, IsPublic: isPublic, Version: version, Render: render, RenderOptions: renderOptions}
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
	if err := svc.Set(context.Background(), "key1", "val1", true, "1.1.0", "select", `[{"label":"开启","value":"1"}]`); err != nil {
		t.Fatal("set:", err)
	}
	cfg, _ := svc.Get(context.Background(), "key1")
	if cfg.Value != "val1" {
		t.Error("value mismatch")
	}
	if !cfg.IsPublic || cfg.Version != "1.1.0" {
		t.Errorf("want is_public=true version=1.1.0, got %+v", cfg)
	}
	if cfg.Render != "select" || cfg.RenderOptions == "" {
		t.Errorf("want render=select with options, got render=%q options=%q", cfg.Render, cfg.RenderOptions)
	}
}

func TestConfigSetUpdate(t *testing.T) {
	repo := &mockConfigRepo{
		configs: map[string]*domain.SystemConfig{
			"site_name": {Key: "site_name", Value: "old"},
		},
	}
	svc := NewConfigService(repo, nil)
	if err := svc.Set(context.Background(), "site_name", "new", true, "1.2.0", "text", ""); err != nil {
		t.Fatal("set:", err)
	}
	cfg, _ := svc.Get(context.Background(), "site_name")
	if cfg.Value != "new" {
		t.Error("value should be updated")
	}
	if !cfg.IsPublic || cfg.Version != "1.2.0" {
		t.Errorf("want is_public=true version=1.2.0, got %+v", cfg)
	}
	if cfg.Render != "text" {
		t.Errorf("want render=text, got %q", cfg.Render)
	}
}

func TestConfigGetAllPublic(t *testing.T) {
	repo := &mockConfigRepo{
		configs: map[string]*domain.SystemConfig{
			"site_name":          {Key: "site_name", Value: "Kingfisher Admin", IsPublic: true},
			"site_logo":          {Key: "site_logo", Value: "/logo.png", IsPublic: true},
			"max_login_attempts": {Key: "max_login_attempts", Value: "5"},
		},
	}
	svc := NewConfigService(repo, nil)
	cfgs, err := svc.GetAllPublic(context.Background())
	if err != nil {
		t.Fatal("get all public:", err)
	}
	if len(cfgs) != 2 {
		t.Errorf("want 2 public configs, got %d", len(cfgs))
	}
}

func TestConfigGetPublic(t *testing.T) {
	repo := &mockConfigRepo{
		configs: map[string]*domain.SystemConfig{
			"site_name": {Key: "site_name", Value: "Kingfisher Admin", IsPublic: true},
		},
	}
	svc := NewConfigService(repo, nil)
	cfg, err := svc.GetPublic(context.Background(), "site_name")
	if err != nil {
		t.Fatal("get public:", err)
	}
	if cfg.Value != "Kingfisher Admin" {
		t.Error("value mismatch")
	}
}

func TestConfigGetPublicPrivateReturnsNotFound(t *testing.T) {
	repo := &mockConfigRepo{
		configs: map[string]*domain.SystemConfig{
			"max_login_attempts": {Key: "max_login_attempts", Value: "5"},
		},
	}
	svc := NewConfigService(repo, nil)
	if _, err := svc.GetPublic(context.Background(), "max_login_attempts"); err == nil {
		t.Error("private config should not be readable via public endpoint")
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
