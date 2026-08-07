package app

import (
	"context"
	"fmt"
	"testing"

	"kingfisher/core/query"
	"kingfisher/extends/dict/domain"
)

// ---- mock repos ----

type mockDictTypeRepo struct {
	types map[uint]*domain.DictType
	seq   uint
}

func (m *mockDictTypeRepo) List(ctx context.Context, q *query.Query) ([]domain.DictType, int64, error) {
	var out []domain.DictType
	for _, t := range m.types {
		out = append(out, *t)
	}
	return out, int64(len(out)), nil
}

func (m *mockDictTypeRepo) GetByID(ctx context.Context, id uint) (*domain.DictType, error) {
	t, ok := m.types[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return t, nil
}

func (m *mockDictTypeRepo) GetByCode(ctx context.Context, code string) (*domain.DictType, error) {
	for _, t := range m.types {
		if t.Code == code {
			return t, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockDictTypeRepo) Create(ctx context.Context, code, name string, isPublic bool, status int, remark, version string) (*domain.DictType, error) {
	m.seq++
	t := &domain.DictType{ID: m.seq, Code: code, Name: name, IsPublic: isPublic, Status: status, Remark: remark, Version: version}
	m.types[t.ID] = t
	return t, nil
}

func (m *mockDictTypeRepo) Update(ctx context.Context, id uint, code, name string, isPublic bool, status int, remark, version string) error {
	t, ok := m.types[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	t.Code = code
	t.Name = name
	t.IsPublic = isPublic
	t.Status = status
	t.Remark = remark
	t.Version = version
	return nil
}

func (m *mockDictTypeRepo) Delete(ctx context.Context, id uint) error {
	delete(m.types, id)
	return nil
}

func (m *mockDictTypeRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		delete(m.types, id)
	}
	return nil
}

func (m *mockDictTypeRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	return nil
}

type mockDictEntryRepo struct {
	entries map[uint]*domain.DictEntry
	seq     uint
}

func (m *mockDictEntryRepo) ListByTypeID(ctx context.Context, typeID uint, q *query.Query) ([]domain.DictEntry, int64, error) {
	var out []domain.DictEntry
	for _, e := range m.entries {
		if e.TypeID == typeID {
			out = append(out, *e)
		}
	}
	return out, int64(len(out)), nil
}

func (m *mockDictEntryRepo) ListByTypeCode(ctx context.Context, code string) ([]domain.DictEntry, error) {
	// mock: code 匹配 type 需由外部手动构造匹配关系，这里简单返回所有 status=1 的条目
	var out []domain.DictEntry
	for _, e := range m.entries {
		if e.Status == 1 {
			out = append(out, *e)
		}
	}
	return out, nil
}

func (m *mockDictEntryRepo) GetByID(ctx context.Context, id uint) (*domain.DictEntry, error) {
	e, ok := m.entries[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return e, nil
}

func (m *mockDictEntryRepo) Create(ctx context.Context, typeID uint, label, value string, sort, status int, remark, version string) (*domain.DictEntry, error) {
	m.seq++
	e := &domain.DictEntry{ID: m.seq, TypeID: typeID, Label: label, Value: value, Sort: sort, Status: status, Remark: remark, Version: version}
	m.entries[e.ID] = e
	return e, nil
}

func (m *mockDictEntryRepo) Update(ctx context.Context, id uint, typeID uint, label, value string, sort, status int, remark, version string) error {
	e, ok := m.entries[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	e.TypeID = typeID
	e.Label = label
	e.Value = value
	e.Sort = sort
	e.Status = status
	e.Remark = remark
	e.Version = version
	return nil
}

func (m *mockDictEntryRepo) Delete(ctx context.Context, id uint) error {
	delete(m.entries, id)
	return nil
}

func (m *mockDictEntryRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		delete(m.entries, id)
	}
	return nil
}

func (m *mockDictEntryRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	return nil
}

func (m *mockDictEntryRepo) DeleteByTypeID(ctx context.Context, typeID uint) error {
	for id, e := range m.entries {
		if e.TypeID == typeID {
			delete(m.entries, id)
		}
	}
	return nil
}

// ---- tests ----

func TestDictTypeCRUD(t *testing.T) {
	typeRepo := &mockDictTypeRepo{types: map[uint]*domain.DictType{}}
	entryRepo := &mockDictEntryRepo{entries: map[uint]*domain.DictEntry{}}
	svc := NewDictTypeService(typeRepo, entryRepo)
	ctx := context.Background()

	// Create
	dt, err := svc.Create(ctx, "gender", "性别", true, 1, "", "1.0.0")
	if err != nil {
		t.Fatal("create type:", err)
	}
	if dt.Code != "gender" || dt.Name != "性别" || !dt.IsPublic {
		t.Errorf("unexpected type: %+v", dt)
	}

	// Create duplicate code
	_, err = svc.Create(ctx, "gender", "性别2", false, 1, "", "1.0.0")
	if err == nil {
		t.Error("should fail on duplicate code")
	}

	// List
	types, _, err := svc.List(ctx, &query.Query{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal("list types:", err)
	}
	if len(types) != 1 {
		t.Errorf("want 1 type, got %d", len(types))
	}

	// GetByID
	got, err := svc.GetByID(ctx, dt.ID)
	if err != nil {
		t.Fatal("get by id:", err)
	}
	if got.Code != "gender" {
		t.Error("code mismatch")
	}

	// Update
	if err := svc.Update(ctx, dt.ID, "gender_v2", "性别V2", false, 0, "备注", "1.1.0"); err != nil {
		t.Fatal("update type:", err)
	}
	got, _ = svc.GetByID(ctx, dt.ID)
	if got.Code != "gender_v2" || got.Name != "性别V2" || got.IsPublic || got.Status != 0 {
		t.Errorf("update not applied: %+v", got)
	}

	// Delete with entries should fail
	_, _ = entryRepo.Create(ctx, dt.ID, "男", "male", 1, 1, "", "1.0.0")
	if err := svc.Delete(ctx, dt.ID); err == nil {
		t.Error("should fail when entries exist")
	}

	// Delete entries then delete type
	_ = entryRepo.DeleteByTypeID(ctx, dt.ID)
	if err := svc.Delete(ctx, dt.ID); err != nil {
		t.Fatal("delete type:", err)
	}
	types, _, _ = svc.List(ctx, &query.Query{Page: 1, PageSize: 20})
	if len(types) != 0 {
		t.Error("should be empty after delete")
	}
}

func TestDictEntryCRUD(t *testing.T) {
	typeRepo := &mockDictTypeRepo{types: map[uint]*domain.DictType{
		1: {ID: 1, Code: "gender", Name: "性别", IsPublic: true, Status: 1},
	}}
	entryRepo := &mockDictEntryRepo{entries: map[uint]*domain.DictEntry{}}
	svc := NewDictEntryService(entryRepo, typeRepo)
	ctx := context.Background()

	// Create
	e, err := svc.Create(ctx, 1, "男", "male", 1, 1, "", "1.0.0")
	if err != nil {
		t.Fatal("create entry:", err)
	}
	if e.Label != "男" || e.Value != "male" {
		t.Errorf("unexpected entry: %+v", e)
	}

	_, _ = svc.Create(ctx, 1, "女", "female", 2, 1, "", "1.0.0")

	// ListByTypeID
	entries, _, err := svc.ListByTypeID(ctx, 1, &query.Query{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal("list entries:", err)
	}
	if len(entries) != 2 {
		t.Errorf("want 2 entries, got %d", len(entries))
	}

	// Update
	if err := svc.Update(ctx, e.ID, 1, "男性", "male", 1, 0, "", "1.1.0"); err != nil {
		t.Fatal("update entry:", err)
	}
	got, _ := svc.GetByID(ctx, e.ID)
	if got.Label != "男性" || got.Status != 0 {
		t.Errorf("update not applied: %+v", got)
	}

	// Delete
	if err := svc.Delete(ctx, e.ID); err != nil {
		t.Fatal("delete entry:", err)
	}
	entries, _, _ = svc.ListByTypeID(ctx, 1, &query.Query{Page: 1, PageSize: 20})
	if len(entries) != 1 {
		t.Errorf("want 1 entry after delete, got %d", len(entries))
	}
}

func TestDictPublicEntries(t *testing.T) {
	typeRepo := &mockDictTypeRepo{types: map[uint]*domain.DictType{
		1: {ID: 1, Code: "gender", Name: "性别", IsPublic: true, Status: 1},
		2: {ID: 2, Code: "secret", Name: "私密", IsPublic: false, Status: 1},
	}}
	entryRepo := &mockDictEntryRepo{entries: map[uint]*domain.DictEntry{
		1: {ID: 1, TypeID: 1, Label: "男", Value: "male", Status: 1},
		2: {ID: 2, TypeID: 1, Label: "女", Value: "female", Status: 1},
	}}
	svc := NewDictEntryService(entryRepo, typeRepo)
	ctx := context.Background()

	// Public type should return entries
	entries, err := svc.ListPublicByCode(ctx, "gender")
	if err != nil {
		t.Fatal("list public entries:", err)
	}
	if len(entries) != 2 {
		t.Errorf("want 2 public entries, got %d", len(entries))
	}

	// Private type should return error
	_, err = svc.ListPublicByCode(ctx, "secret")
	if err == nil {
		t.Error("should fail for non-public type")
	}

	// Non-existent type
	_, err = svc.ListPublicByCode(ctx, "notexist")
	if err == nil {
		t.Error("should fail for non-existent type")
	}
}
