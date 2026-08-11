package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"kingfisher/core/errcode"
	"kingfisher/core/query"
	"kingfisher/extends/doc/domain"
	"kingfisher/extends/doc/port"
)

// mockCache 内存缓存实现
type mockCache struct {
	store map[string]string
	del   map[string]bool
}

func newMockCache() *mockCache                                       { return &mockCache{store: map[string]string{}, del: map[string]bool{}} }
func (m *mockCache) Get(_ context.Context, k string) (string, error) { return m.store[k], nil }
func (m *mockCache) Set(_ context.Context, k string, v any, _ time.Duration) error {
	m.store[k] = v.(string)
	return nil
}
func (m *mockCache) Delete(_ context.Context, keys ...string) error {
	for _, k := range keys {
		delete(m.store, k)
		m.del[k] = true
	}
	return nil
}
func (m *mockCache) DeleteByPattern(_ context.Context, _ string) error         { return nil }
func (m *mockCache) Exists(_ context.Context, _ string) (bool, error)          { return false, nil }
func (m *mockCache) Incr(_ context.Context, _ string) (int64, error)           { return 0, nil }
func (m *mockCache) Expire(_ context.Context, _ string, _ time.Duration) error { return nil }

// stubRepo 函数字段式桩仓储：测试只需覆盖用到的行为
type stubRepo struct {
	findAllDirs   func(ctx context.Context) ([]domain.DocDirectory, error)
	findDirByID   func(ctx context.Context, id uint) (*domain.DocDirectory, error)
	getDocByID    func(ctx context.Context, id uint, userID uint, roleIDs []uint, isAdmin bool) (*domain.Document, error)
	createWithVer func(ctx context.Context, doc *domain.Document, ver *domain.DocVersion) (*domain.Document, error)
	updateWithVer func(ctx context.Context, id uint, title, content string, ownerID uint, note string) error
	setDirRoles   func(ctx context.Context, dirID uint, roleIDs []uint) error
}

func (s *stubRepo) FindAllDirs(ctx context.Context) ([]domain.DocDirectory, error) {
	if s.findAllDirs != nil {
		return s.findAllDirs(ctx)
	}
	return nil, nil
}
func (s *stubRepo) FindDirByID(ctx context.Context, id uint) (*domain.DocDirectory, error) {
	if s.findDirByID != nil {
		return s.findDirByID(ctx, id)
	}
	return nil, errors.New("not found")
}
func (s *stubRepo) CreateDir(ctx context.Context, d *domain.DocDirectory) error          { return nil }
func (s *stubRepo) UpdateDir(ctx context.Context, id uint, updates map[string]any) error { return nil }
func (s *stubRepo) DeleteDir(ctx context.Context, id uint) error                         { return nil }
func (s *stubRepo) HasDirChildren(ctx context.Context, parentID uint) (bool, error) {
	return false, nil
}
func (s *stubRepo) HasDirDocuments(ctx context.Context, dirID uint) (bool, error) { return false, nil }
func (s *stubRepo) GetDirRoleIDs(ctx context.Context, dirID uint) ([]uint, error) { return nil, nil }
func (s *stubRepo) SetDirRoles(ctx context.Context, dirID uint, roleIDs []uint) error {
	if s.setDirRoles != nil {
		return s.setDirRoles(ctx, dirID, roleIDs)
	}
	return nil
}
func (s *stubRepo) ListDocs(ctx context.Context, dirID uint, q *query.Query, userID uint, roleIDs []uint, isAdmin bool) ([]domain.Document, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) GetDocByID(ctx context.Context, id uint, userID uint, roleIDs []uint, isAdmin bool) (*domain.Document, error) {
	if s.getDocByID != nil {
		return s.getDocByID(ctx, id, userID, roleIDs, isAdmin)
	}
	return nil, nil
}
func (s *stubRepo) CreateWithVersion(ctx context.Context, doc *domain.Document, ver *domain.DocVersion) (*domain.Document, error) {
	if s.createWithVer != nil {
		return s.createWithVer(ctx, doc, ver)
	}
	return doc, nil
}
func (s *stubRepo) UpdateWithVersion(ctx context.Context, id uint, title, content string, ownerID uint, note string) error {
	if s.updateWithVer != nil {
		return s.updateWithVer(ctx, id, title, content, ownerID, note)
	}
	return nil
}
func (s *stubRepo) RestoreToVersion(ctx context.Context, docID uint, fromVersionNo int, ownerID uint, note string) error {
	return nil
}
func (s *stubRepo) SetDocStatus(ctx context.Context, id uint, status string, publishedAt *time.Time) error {
	return nil
}
func (s *stubRepo) DeleteDoc(ctx context.Context, id uint) error { return nil }
func (s *stubRepo) ListVersions(ctx context.Context, docID uint) ([]domain.DocVersion, error) {
	return nil, nil
}
func (s *stubRepo) GetVersion(ctx context.Context, docID uint, versionNo int) (*domain.DocVersion, error) {
	return nil, nil
}

func TestGetTreeFiltersByRole(t *testing.T) {
	repo := &stubRepo{
		findAllDirs: func(_ context.Context) ([]domain.DocDirectory, error) {
			return []domain.DocDirectory{
				{ID: 1, ParentID: 0, Name: "公开", GrantedRoles: []uint{1, 3, 4}},
				{ID: 2, ParentID: 0, Name: "仅admin", GrantedRoles: []uint{1}},
				{ID: 3, ParentID: 1, Name: "子目录", GrantedRoles: []uint{1, 3}},
				{ID: 4, ParentID: 0, Name: "未授权", GrantedRoles: nil}, // 默认拒绝
			}, nil
		},
	}
	svc := NewDocService(repo, nil)

	// 非 admin viewer(role 4)：只见 dir1，且 dir1 下的 dir3(无 role4) 被裁剪
	tree, err := svc.GetTree(context.Background(), []uint{4}, false)
	if err != nil {
		t.Fatalf("gettree: %v", err)
	}
	if len(tree) != 1 || tree[0].Name != "公开" {
		t.Fatalf("viewer 应只见 公开，got %+v", tree)
	}
	if len(tree[0].Children) != 0 {
		t.Fatalf("viewer 不应见 子目录，got %+v", tree[0].Children)
	}

	// admin：全量（含未授权目录）
	tree2, _ := svc.GetTree(context.Background(), nil, true)
	if len(tree2) != 3 || len(tree2[0].Children) != 1 {
		t.Fatalf("admin 应见全部 3 个顶级+1 子目录，got %+v", tree2)
	}

	// 无授权目录（GrantedRoles 空）= 默认拒绝：即使角色匹配也看不到未授权目录
	tree3, _ := svc.GetTree(context.Background(), []uint{4}, false)
	for _, n := range tree3 {
		if n.Name == "未授权" {
			t.Fatalf("未授权目录对非 admin 默认拒绝，got %+v", tree3)
		}
	}
}

func TestCreateDocSanitizesXSS(t *testing.T) {
	svc := NewDocService(&stubRepo{}, nil)
	svc.repo = &stubRepo{
		findDirByID: func(_ context.Context, _ uint) (*domain.DocDirectory, error) {
			return &domain.DocDirectory{ID: 1, GrantedRoles: []uint{1, 3, 4}}, nil
		},
		createWithVer: func(_ context.Context, doc *domain.Document, _ *domain.DocVersion) (*domain.Document, error) {
			return doc, nil
		},
	}
	doc, err := svc.CreateDoc(context.Background(), 1, "标题",
		"<p>ok</p><script>alert(1)</script><img src=x onerror=alert(2)>", 1,
		domain.VisibilityShared, "", []uint{4}, false)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if doc.Content != "<p>ok</p><img src=\"x\">" {
		t.Fatalf("XSS 未被清洗干净，got: %q", doc.Content)
	}
}

func TestCreateDocDirNotVisible(t *testing.T) {
	svc := NewDocService(&stubRepo{
		findDirByID: func(_ context.Context, _ uint) (*domain.DocDirectory, error) {
			return &domain.DocDirectory{ID: 1, GrantedRoles: []uint{1}}, nil // 仅 admin 可见
		},
	}, nil)
	_, err := svc.CreateDoc(context.Background(), 1, "标题", "x", 5, domain.VisibilityShared, "", []uint{4}, false)
	if err == nil {
		t.Fatal("无权限创建应报错")
	}
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrDocDirNotVisible {
		t.Fatalf("应返回 ErrDocDirNotVisible，got %v", err)
	}
}

func TestUpdateDocMapsVersionConflict(t *testing.T) {
	repo := &stubRepo{
		getDocByID: func(_ context.Context, _ uint, _ uint, _ []uint, _ bool) (*domain.Document, error) {
			return &domain.Document{ID: 1, OwnerID: 1}, nil
		},
		updateWithVer: func(_ context.Context, _ uint, _ string, _ string, _ uint, _ string) error {
			return port.ErrVersionConflict
		},
	}
	svc := NewDocService(repo, nil)
	_, err := svc.UpdateDoc(context.Background(), 1, "t", "c", "", 1, false)
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrDocVersionConflict {
		t.Fatalf("版本冲突应映射为 ErrDocVersionConflict，got %v", err)
	}
}

func TestSetDirRolesInvalidatesCache(t *testing.T) {
	cache := newMockCache()
	repo := &stubRepo{}
	svc := NewDocService(repo, cache)
	_, _ = svc.getFullTree(context.Background()) // 触发写缓存
	if _, ok := cache.store[treeCacheKey]; !ok {
		t.Fatal("目录树应写入缓存")
	}
	if err := svc.SetDirRoles(context.Background(), 1, []uint{1}); err != nil {
		t.Fatalf("setroles: %v", err)
	}
	if !cache.del[treeCacheKey] {
		t.Fatal("SetDirRoles 后应失效目录树缓存")
	}
}
