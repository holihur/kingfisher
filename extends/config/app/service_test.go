package app

import (
	"context"
	"fmt"
	"testing"

	"kingfisher/core/query"
	"kingfisher/extends/config/domain"
)

type mockConfigRepo struct {
	configs map[string]*domain.SystemConfig
}

func (m *mockConfigRepo) List(ctx context.Context, q *query.Query) ([]domain.SystemConfig, int64, error) {
	var out []domain.SystemConfig
	for _, c := range m.configs {
		out = append(out, *c)
	}
	return out, int64(len(out)), nil
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

func (m *mockConfigRepo) Set(ctx context.Context, key, value string, isPublic bool, version, render, renderOptions string, groupID uint) error {
	if m.configs == nil {
		m.configs = map[string]*domain.SystemConfig{}
	}
	m.configs[key] = &domain.SystemConfig{Key: key, Value: value, IsPublic: isPublic, Version: version, Render: render, RenderOptions: renderOptions, GroupID: groupID}
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
	cfgs, total, err := svc.List(context.Background(), &query.Query{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal("list:", err)
	}
	if len(cfgs) != 2 || total != 2 {
		t.Errorf("want 2 configs, got %d total=%d", len(cfgs), total)
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
	if err := svc.Set(context.Background(), "key1", "val1", true, "1.1.0", "select", `[{"label":"开启","value":"1"}]`, 2); err != nil {
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
	if cfg.GroupID != 2 {
		t.Errorf("want group_id=2, got %d", cfg.GroupID)
	}
}

func TestConfigSetUpdate(t *testing.T) {
	repo := &mockConfigRepo{
		configs: map[string]*domain.SystemConfig{
			"site_name": {Key: "site_name", Value: "old"},
		},
	}
	svc := NewConfigService(repo, nil)
	if err := svc.Set(context.Background(), "site_name", "new", true, "1.2.0", "text", "", 1); err != nil {
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
	if cfg.GroupID != 1 {
		t.Errorf("want group_id=1, got %d", cfg.GroupID)
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

// ---- 配置分组 ----

type mockGroupRepo struct {
	groups []domain.ConfigGroup
}

func (m *mockGroupRepo) List(ctx context.Context) ([]domain.ConfigGroup, error) {
	return m.groups, nil
}

func (m *mockGroupRepo) Create(ctx context.Context, name string, sort int) (*domain.ConfigGroup, error) {
	g := domain.ConfigGroup{ID: uint(len(m.groups) + 1), Name: name, Sort: sort}
	m.groups = append(m.groups, g)
	return &g, nil
}

func (m *mockGroupRepo) Update(ctx context.Context, id uint, name string, sort int) error {
	for i := range m.groups {
		if m.groups[i].ID == id {
			m.groups[i].Name = name
			m.groups[i].Sort = sort
			return nil
		}
	}
	return fmt.Errorf("group %d not found", id)
}

func (m *mockGroupRepo) Delete(ctx context.Context, id uint) error {
	for i := range m.groups {
		if m.groups[i].ID == id {
			m.groups = append(m.groups[:i], m.groups[i+1:]...)
			return nil
		}
	}
	return nil
}

func TestConfigGroupCRUD(t *testing.T) {
	repo := &mockGroupRepo{}
	svc := NewConfigGroupService(repo)

	// Create
	g, err := svc.Create(context.Background(), "站点", 1)
	if err != nil {
		t.Fatal("create group:", err)
	}
	if g.Name != "站点" || g.Sort != 1 {
		t.Errorf("unexpected group: %+v", g)
	}
	if _, err := svc.Create(context.Background(), "安全", 2); err != nil {
		t.Fatal("create group:", err)
	}

	// List
	groups, err := svc.List(context.Background())
	if err != nil {
		t.Fatal("list groups:", err)
	}
	if len(groups) != 2 {
		t.Fatalf("want 2 groups, got %d", len(groups))
	}

	// Update
	if err := svc.Update(context.Background(), g.ID, "站点设置", 3); err != nil {
		t.Fatal("update group:", err)
	}
	groups, _ = svc.List(context.Background())
	found := false
	for _, gg := range groups {
		if gg.ID == g.ID && gg.Name == "站点设置" && gg.Sort == 3 {
			found = true
		}
	}
	if !found {
		t.Error("group not updated correctly")
	}

	// Delete
	if err := svc.Delete(context.Background(), g.ID); err != nil {
		t.Fatal("delete group:", err)
	}
	groups, _ = svc.List(context.Background())
	if len(groups) != 1 {
		t.Errorf("want 1 group after delete, got %d", len(groups))
	}
}
