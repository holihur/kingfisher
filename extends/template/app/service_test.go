package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"kingfisher/core/errcode"
	"kingfisher/core/query"
	"kingfisher/extends/template/domain"
)

// mockTemplateRepo 内存实现，测试 TemplateService 用
type mockTemplateRepo struct {
	templates map[uint]*domain.Template
	seq       uint
}

func (m *mockTemplateRepo) List(ctx context.Context, q *query.Query) ([]domain.Template, int64, error) {
	var out []domain.Template
	for _, t := range m.templates {
		out = append(out, *t)
	}
	return out, int64(len(out)), nil
}

func (m *mockTemplateRepo) GetByID(ctx context.Context, id uint) (*domain.Template, error) {
	t, ok := m.templates[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return t, nil
}

func (m *mockTemplateRepo) GetByCode(ctx context.Context, code string) (*domain.Template, error) {
	for _, t := range m.templates {
		if t.Code == code {
			return t, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockTemplateRepo) Create(ctx context.Context, name, code, templateType, title, content string, status int, remark, version string) (*domain.Template, error) {
	m.seq++
	t := &domain.Template{ID: m.seq, Name: name, Code: code, TemplateType: templateType, Title: title, Content: content, Status: status, Remark: remark, Version: version}
	m.templates[t.ID] = t
	return t, nil
}

func (m *mockTemplateRepo) Update(ctx context.Context, id uint, name, code, templateType, title, content string, status int, remark, version string) error {
	t, ok := m.templates[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	t.Name = name
	t.Code = code
	t.TemplateType = templateType
	t.Title = title
	t.Content = content
	t.Status = status
	t.Remark = remark
	t.Version = version
	return nil
}

func (m *mockTemplateRepo) Delete(ctx context.Context, id uint) error {
	delete(m.templates, id)
	return nil
}

func (m *mockTemplateRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		delete(m.templates, id)
	}
	return nil
}

func (m *mockTemplateRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	for _, id := range ids {
		if t, ok := m.templates[id]; ok {
			t.Status = status
		}
	}
	return nil
}

func TestTemplateCRUD(t *testing.T) {
	svc := NewTemplateService(&mockTemplateRepo{templates: map[uint]*domain.Template{}})
	ctx := context.Background()

	// Create
	tpl, err := svc.Create(ctx, "欢迎消息", "welcome", "message", "欢迎 {{nickname}}", "你好 {{nickname}}", 1, "", "1.0.0")
	if err != nil {
		t.Fatal("create:", err)
	}
	if tpl.Code != "welcome" || tpl.TemplateType != "message" || tpl.Status != 1 {
		t.Errorf("unexpected template: %+v", tpl)
	}

	// Duplicate code rejected
	_, err = svc.Create(ctx, "其他", "welcome", "text", "", "", 1, "", "1.0.0")
	if err == nil {
		t.Error("should fail on duplicate code")
	}
	var e *Error
	if !errors.As(err, &e) || e.Code != errcode.ErrTemplateCodeExists {
		t.Errorf("want ErrTemplateCodeExists, got %v", err)
	}

	// List
	items, total, err := svc.List(ctx, &query.Query{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal("list:", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("want 1 template, got total=%d n=%d", total, len(items))
	}

	// GetByID
	got, err := svc.GetByID(ctx, tpl.ID)
	if err != nil {
		t.Fatal("get by id:", err)
	}
	if got.Name != "欢迎消息" {
		t.Errorf("name mismatch: %+v", got)
	}

	// Update
	if err := svc.Update(ctx, tpl.ID, "欢迎消息V2", "welcome_v2", "message", "新标题", "新内容...", 0, "备注", "1.1.0"); err != nil {
		t.Fatal("update:", err)
	}
	got, _ = svc.GetByID(ctx, tpl.ID)
	if got.Code != "welcome_v2" || got.Name != "欢迎消息V2" || got.Status != 0 {
		t.Errorf("update not applied: %+v", got)
	}

	// Update to an existing code should fail
	_, _ = svc.Create(ctx, "另一个", "welcome", "message", "", "", 1, "", "1.0.0")
	if err := svc.Update(ctx, tpl.ID, "x", "welcome", "message", "", "", 1, "", "1.1.0"); err == nil {
		t.Error("should fail when updating to duplicate code")
	}

	// BatchUpdateStatus
	if err := svc.BatchUpdateStatus(ctx, []uint{tpl.ID}, 1); err != nil {
		t.Fatal("batch status:", err)
	}
	got, _ = svc.GetByID(ctx, tpl.ID)
	if got.Status != 1 {
		t.Error("batch status not applied")
	}

	// Delete
	if err := svc.BatchDelete(ctx, []uint{tpl.ID}); err != nil {
		t.Fatal("batch delete:", err)
	}
	items, total, _ = svc.List(ctx, &query.Query{Page: 1, PageSize: 20})
	if total != 1 || len(items) != 1 {
		t.Error("want 1 remaining template after batch delete")
	}
	// 剩余的应是第二次创建的模版（code=welcome），tpl 已被删除
	if items[0].Code != "welcome" || items[0].Name != "另一个" {
		t.Errorf("remaining template mismatch: %+v", items[0])
	}
}
