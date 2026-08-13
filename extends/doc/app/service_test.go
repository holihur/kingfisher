package app

import (
	"context"
	"errors"
	"strings"
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
	updateWithVer func(ctx context.Context, id uint, title, content, visibility string, ownerID uint, note string) error
	listAllPublic func(ctx context.Context) ([]domain.Document, error)
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
func (s *stubRepo) ListDocs(ctx context.Context, dirID uint, q *query.Query, userID uint, roleIDs []uint, isAdmin bool) ([]domain.Document, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) ListAllVisibleDocs(ctx context.Context, userID uint, roleIDs []uint, isAdmin bool) ([]domain.Document, error) {
	return nil, nil
}
func (s *stubRepo) ListAllPublicDocs(ctx context.Context) ([]domain.Document, error) {
	if s.listAllPublic != nil {
		return s.listAllPublic(ctx)
	}
	return nil, nil
}
func (s *stubRepo) GetDocByID(ctx context.Context, id uint, userID uint, roleIDs []uint, isAdmin bool) (*domain.Document, error) {
	if s.getDocByID != nil {
		return s.getDocByID(ctx, id, userID, roleIDs, isAdmin)
	}
	return nil, nil
}
func (s *stubRepo) GetPublicDoc(ctx context.Context, id uint) (*domain.Document, error) {
	return nil, errors.New("not found")
}
func (s *stubRepo) CreateWithVersion(ctx context.Context, doc *domain.Document, ver *domain.DocVersion) (*domain.Document, error) {
	if s.createWithVer != nil {
		return s.createWithVer(ctx, doc, ver)
	}
	return doc, nil
}
func (s *stubRepo) UpdateWithVersion(ctx context.Context, id uint, title, content, visibility string, ownerID uint, note string) error {
	if s.updateWithVer != nil {
		return s.updateWithVer(ctx, id, title, content, visibility, ownerID, note)
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

func TestGetTreeFiltersByVisibility(t *testing.T) {
	repo := &stubRepo{
		findAllDirs: func(_ context.Context) ([]domain.DocDirectory, error) {
			return []domain.DocDirectory{
				{ID: 1, ParentID: 0, Name: "公开", Visibility: domain.VisibilityShared},
				{ID: 2, ParentID: 0, Name: "仅admin", Visibility: domain.VisibilityPrivate},
				{ID: 3, ParentID: 1, Name: "子目录", Visibility: domain.VisibilityShared},
				{ID: 4, ParentID: 0, Name: "私有默认", Visibility: ""}, // 空视为非 shared → 非 admin 隐藏
			}, nil
		},
	}
	svc := NewDocService(repo, nil)

	// 非 admin：只见 shared 目录（private 及其下子树整体裁剪）
	tree, err := svc.GetTree(context.Background(), 99, []uint{4}, false)
	if err != nil {
		t.Fatalf("gettree: %v", err)
	}
	if len(tree) != 1 || tree[0].Name != "公开" {
		t.Fatalf("非 admin 应只见 公开，got %+v", tree)
	}
	if len(tree[0].Children) != 1 || tree[0].Children[0].Name != "子目录" {
		t.Fatalf("shared 子目录应可见，got %+v", tree[0].Children)
	}

	// admin：全量（含 private 目录）
	tree2, _ := svc.GetTree(context.Background(), 1, nil, true)
	if len(tree2) != 3 || len(tree2[0].Children) != 1 {
		t.Fatalf("admin 应见全部 3 个顶级+1 子目录，got %+v", tree2)
	}

	// 私有默认（visibility 空）对非 admin 隐藏
	for _, n := range tree {
		if n.Name == "私有默认" {
			t.Fatalf("私有目录对非 admin 应隐藏，got %+v", tree)
		}
	}
}

func TestGetPublicTree(t *testing.T) {
	repo := &stubRepo{
		findAllDirs: func(_ context.Context) ([]domain.DocDirectory, error) {
			return []domain.DocDirectory{
				{ID: 1, ParentID: 0, Name: "公开目录", Visibility: domain.VisibilityShared},
				{ID: 2, ParentID: 0, Name: "私有目录", Visibility: domain.VisibilityPrivate},
				{ID: 3, ParentID: 1, Name: "公开子目录", Visibility: domain.VisibilityShared},
			}, nil
		},
		listAllPublic: func(_ context.Context) ([]domain.Document, error) {
			return []domain.Document{
				{ID: 10, DirID: 1, Title: "文档A", Status: domain.DocStatusPublished, Visibility: domain.VisibilityShared},
				{ID: 11, DirID: 3, Title: "文档B", Status: domain.DocStatusPublished, Visibility: domain.VisibilityShared},
			}, nil
		},
	}
	svc := NewDocService(repo, nil)
	tree, err := svc.GetPublicTree(context.Background())
	if err != nil {
		t.Fatalf("GetPublicTree: %v", err)
	}
	// 私有目录被裁剪；公开目录挂上公开文档
	if len(tree) != 1 || tree[0].Name != "公开目录" {
		t.Fatalf("公开树应只剩 公开目录，got %+v", tree)
	}
	if len(tree[0].Docs) != 1 || tree[0].Docs[0].Title != "文档A" {
		t.Fatalf("公开目录应挂 文档A，got %+v", tree[0].Docs)
	}
	if len(tree[0].Children) != 1 || len(tree[0].Children[0].Docs) != 1 {
		t.Fatalf("公开子目录应挂 文档B，got %+v", tree[0].Children)
	}
}

// 一段合法的 Lexical editorState JSON（含标题/待办/代码块/提示框/文本格式）
const validDocStateJSON = `{"root":{"type":"root","direction":null,"format":"","indent":0,"version":1,"children":[` +
	`{"type":"heading","tag":"h1","version":1,"children":[{"type":"text","text":"标题","format":1,"style":"","mode":"normal","detail":0,"version":1}]},` +
	`{"type":"list","listType":"check","start":1,"tag":"ul","version":1,"children":[{"type":"listitem","checked":true,"value":1,"version":1,"children":[{"type":"text","text":"待办","format":0,"style":"color: rgb(225, 29, 72);","mode":"normal","detail":0,"version":1}]}]},` +
	`{"type":"code","language":"go","version":1,"children":[{"type":"code-highlight","text":"fmt.Println(1)","highlightType":"","format":0,"style":"","mode":"normal","detail":0,"version":1}]},` +
	`{"type":"callout","icon":"💡","version":1,"children":[{"type":"paragraph","version":1,"children":[]}]}` +
	`]}}`

func TestValidateDocContent(t *testing.T) {
	cases := []struct {
		name    string
		content string
		wantErr bool
	}{
		{name: "合法 editorState JSON", content: validDocStateJSON},
		{name: "空文档允许", content: ""},
		{name: "纯空格允许", content: "  \n "},
		{name: "旧 HTML 数据拒绝", content: "<p>ok</p><script>alert(1)</script>", wantErr: true},
		{name: "非 JSON 字符串拒绝", content: "hello", wantErr: true},
		{name: "缺 root 拒绝", content: `{"foo":"bar"}`, wantErr: true},
		{name: "未知节点类型拒绝", content: `{"root":{"children":[{"type":"evil","children":[]}]}}`, wantErr: true},
		{name: "超深嵌套拒绝", content: deepJSON(21), wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateDocContent(c.content)
			if c.wantErr != (err != nil) {
				t.Fatalf("validateDocContent(%q) err=%v, wantErr=%v", c.content, err, c.wantErr)
			}
			if err != nil {
				var appErr *Error
				if !errors.As(err, &appErr) || appErr.Code != errcode.ErrDocContentInvalid {
					t.Fatalf("应返回 ErrDocContentInvalid，got %v", err)
				}
			}
		})
	}
}

// deepJSON 构造深度为 depth 的嵌套段落 JSON
func deepJSON(depth int) string {
	var sb strings.Builder
	sb.WriteString(`{"root":{"children":[`)
	for range depth {
		sb.WriteString(`{"type":"paragraph","children":[`)
	}
	sb.WriteString(`{"type":"text","text":"x"}`)
	for range depth {
		sb.WriteString(`]}`)
	}
	sb.WriteString(`]}}`)
	return sb.String()
}

func TestCreateDocRejectsInvalidContent(t *testing.T) {
	svc := NewDocService(&stubRepo{}, nil)
	svc.repo = &stubRepo{
		findDirByID: func(_ context.Context, _ uint) (*domain.DocDirectory, error) {
			return &domain.DocDirectory{ID: 1, Visibility: domain.VisibilityShared}, nil
		},
		createWithVer: func(_ context.Context, doc *domain.Document, _ *domain.DocVersion) (*domain.Document, error) {
			return doc, nil
		},
	}
	// 旧 HTML 内容应被拒绝（不再是清洗后放行）
	_, err := svc.CreateDoc(context.Background(), 1, "标题",
		"<p>ok</p><script>alert(1)</script>", 1,
		domain.VisibilityShared, "", []uint{4}, false)
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrDocContentInvalid {
		t.Fatalf("旧 HTML 内容应返回 ErrDocContentInvalid，got %v", err)
	}
	// 合法 JSON 内容原样存储（不再清洗改动）
	doc, err := svc.CreateDoc(context.Background(), 1, "标题", validDocStateJSON, 1,
		domain.VisibilityShared, "", []uint{4}, false)
	if err != nil {
		t.Fatalf("合法 JSON 创建失败: %v", err)
	}
	if doc.Content != validDocStateJSON {
		t.Fatalf("内容应原样存储\n got: %s\nwant: %s", doc.Content, validDocStateJSON)
	}
}

func TestCreateDocDirNotVisible(t *testing.T) {
	svc := NewDocService(&stubRepo{
		findDirByID: func(_ context.Context, _ uint) (*domain.DocDirectory, error) {
			return &domain.DocDirectory{ID: 1, Visibility: domain.VisibilityPrivate}, nil // private 仅 admin 可见
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
		updateWithVer: func(_ context.Context, _ uint, _ string, _ string, _ string, _ uint, _ string) error {
			return port.ErrVersionConflict
		},
	}
	svc := NewDocService(repo, nil)
	_, err := svc.UpdateDoc(context.Background(), 1, "t", validDocStateJSON, "", "", 1, false)
	var appErr *Error
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrDocVersionConflict {
		t.Fatalf("版本冲突应映射为 ErrDocVersionConflict，got %v", err)
	}
}

func TestUpdateDirInvalidatesCache(t *testing.T) {
	cache := newMockCache()
	repo := &stubRepo{}
	svc := NewDocService(repo, cache)
	_, _ = svc.getFullTree(context.Background()) // 触发写缓存
	if _, ok := cache.store[treeCacheKey]; !ok {
		t.Fatal("目录树应写入缓存")
	}
	if err := svc.UpdateDir(context.Background(), 1, map[string]any{"visibility": domain.VisibilityPrivate}); err != nil {
		t.Fatalf("updatedir: %v", err)
	}
	if !cache.del[treeCacheKey] {
		t.Fatal("UpdateDir 后应失效目录树缓存")
	}
}
