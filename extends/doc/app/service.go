package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	coreCache "kingfisher/core/cache"
	"kingfisher/core/errcode"
	"kingfisher/core/query"
	"kingfisher/extends/doc/domain"
	"kingfisher/extends/doc/port"
)

// Error 携带 errcode 的错误类型，handler 层据此映射到 HTTP 错误码
type Error struct{ Code int }

func (e *Error) Error() string { return errcode.Msg(e.Code) }

// treeCacheKey 目录树缓存键（含每个目录的授权角色，读时按用户角色过滤）
const treeCacheKey = "doc:tree"

// DocService 文档服务
type DocService struct {
	repo  port.DocRepository
	cache coreCache.Cache
}

func NewDocService(repo port.DocRepository, cache coreCache.Cache) *DocService {
	return &DocService{repo: repo, cache: cache}
}

// ———— 目录 ————

// GetTree 返回当前用户可见的目录树（含每目录下可见的文档叶子）。
// 目录树本体缓存全量，读时按角色过滤；文档叶子不缓存，实时按可见性查询。
func (s *DocService) GetTree(ctx context.Context, userID uint, roleIDs []uint, isAdmin bool) ([]domain.DocDirectory, error) {
	tree, err := s.getFullTree(ctx)
	if err != nil {
		return nil, err
	}
	tree = filterTree(tree, roleIDs, isAdmin)
	if len(tree) == 0 {
		return tree, nil
	}
	docs, err := s.repo.ListAllVisibleDocs(ctx, userID, roleIDs, isAdmin)
	if err != nil {
		return nil, err
	}
	byDir := make(map[uint][]domain.DocTreeItem, len(docs))
	for _, d := range docs {
		byDir[d.DirID] = append(byDir[d.DirID], domain.DocTreeItem{
			ID: d.ID, Title: d.Title, Status: d.Status, Visibility: d.Visibility,
		})
	}
	attachDocs(tree, byDir)
	return tree, nil
}

// attachDocs 递归把可见文档挂到对应目录节点下。
func attachDocs(dirs []domain.DocDirectory, byDir map[uint][]domain.DocTreeItem) {
	for i := range dirs {
		dirs[i].Docs = byDir[dirs[i].ID]
		attachDocs(dirs[i].Children, byDir)
	}
}

func (s *DocService) getFullTree(ctx context.Context) ([]domain.DocDirectory, error) {
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, treeCacheKey); err == nil && cached != "" {
			var tree []domain.DocDirectory
			if json.Unmarshal([]byte(cached), &tree) == nil {
				return tree, nil
			}
		}
	}
	dirs, err := s.repo.FindAllDirs(ctx)
	if err != nil {
		return nil, err
	}
	tree := buildTree(dirs, 0)
	if s.cache != nil {
		if data, err := json.Marshal(tree); err == nil {
			_ = s.cache.Set(ctx, treeCacheKey, string(data), 10*time.Minute)
		}
	}
	return tree, nil
}

func (s *DocService) CreateDir(ctx context.Context, d *domain.DocDirectory) error {
	if err := s.repo.CreateDir(ctx, d); err != nil {
		return err
	}
	_ = s.invalidateTreeCache(ctx)
	return nil
}

func (s *DocService) UpdateDir(ctx context.Context, id uint, updates map[string]any) error {
	if err := s.repo.UpdateDir(ctx, id, updates); err != nil {
		return err
	}
	_ = s.invalidateTreeCache(ctx)
	return nil
}

func (s *DocService) DeleteDir(ctx context.Context, id uint) error {
	hasChildren, _ := s.repo.HasDirChildren(ctx, id)
	if hasChildren {
		return &Error{Code: errcode.ErrDocDirHasChildren}
	}
	hasDocs, _ := s.repo.HasDirDocuments(ctx, id)
	if hasDocs {
		return &Error{Code: errcode.ErrDocDirHasDocuments}
	}
	if err := s.repo.DeleteDir(ctx, id); err != nil {
		return err
	}
	_ = s.invalidateTreeCache(ctx)
	return nil
}

func (s *DocService) invalidateTreeCache(ctx context.Context) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.Delete(ctx, treeCacheKey)
}

// ———— 文档 ————

func (s *DocService) ListDocs(ctx context.Context, dirID uint, q *query.Query, userID uint, roleIDs []uint, isAdmin bool) ([]domain.Document, int64, error) {
	return s.repo.ListDocs(ctx, dirID, q, userID, roleIDs, isAdmin)
}

func (s *DocService) GetDoc(ctx context.Context, id uint, userID uint, roleIDs []uint, isAdmin bool) (*domain.Document, error) {
	doc, err := s.repo.GetDocByID(ctx, id, userID, roleIDs, isAdmin)
	if err != nil {
		return nil, &Error{Code: errcode.ErrDocNotFound}
	}
	return doc, nil
}

// GetPublicTree 公开目录树（匿名）：只含 shared 目录及其下 published+shared 文档叶子，
// 供公开文档页左侧导航。private 目录整体隐藏。
func (s *DocService) GetPublicTree(ctx context.Context) ([]domain.DocDirectory, error) {
	tree, err := s.getFullTree(ctx)
	if err != nil {
		return nil, err
	}
	publicDocs, err := s.repo.ListAllPublicDocs(ctx)
	if err != nil {
		return nil, err
	}
	return filterPublicTree(tree, publicDocs), nil
}

// GetPublicDoc 公开文档（已发布+共享），匿名可读；不满足公开条件 → 404 隐藏。
func (s *DocService) GetPublicDoc(ctx context.Context, id uint) (*domain.Document, error) {
	doc, err := s.repo.GetPublicDoc(ctx, id)
	if err != nil {
		return nil, &Error{Code: errcode.ErrDocNotFound}
	}
	return doc, nil
}

// CreateDoc 创建文档（初始 draft + 版本 1）。dirID 需对用户可见。
func (s *DocService) CreateDoc(ctx context.Context, dirID uint, title, content string, ownerID uint, visibility, note string, roleIDs []uint, isAdmin bool) (*domain.Document, error) {
	dir, err := s.repo.FindDirByID(ctx, dirID)
	if err != nil {
		return nil, &Error{Code: errcode.ErrDocDirNotFound}
	}
	if !canSeeDir(dir, roleIDs, isAdmin) {
		return nil, &Error{Code: errcode.ErrDocDirNotVisible}
	}
	if title == "" {
		return nil, &Error{Code: errcode.ErrInvalidParam}
	}
	if err := validateDocContent(content); err != nil {
		return nil, err
	}
	doc := &domain.Document{
		DirID: dirID, Title: title, Content: content, OwnerID: ownerID,
		Visibility: visibility, Status: domain.DocStatusDraft, CurrentVersion: 1,
	}
	ver := &domain.DocVersion{OwnerID: ownerID, Note: note}
	return s.repo.CreateWithVersion(ctx, doc, ver)
}

// UpdateDoc 更新文档（追加新版本）。仅作者或 admin 可写（含 private/shared 的写权限）。
func (s *DocService) UpdateDoc(ctx context.Context, id uint, title, content, visibility, note string, userID uint, isAdmin bool) (*domain.Document, error) {
	doc, err := s.repo.GetDocByID(ctx, id, userID, []uint{}, true) // 用 admin 视角取到行，再做写权限判定
	if err != nil {
		return nil, &Error{Code: errcode.ErrDocNotFound}
	}
	if !isAdmin && doc.OwnerID != userID {
		return nil, &Error{Code: errcode.ErrDocForbidden}
	}
	// 仅作者/admin 可改可见性（shared/private 是敏感属性）
	if visibility != "" && !isAdmin && doc.OwnerID != userID {
		return nil, &Error{Code: errcode.ErrDocForbidden}
	}
	if err := validateDocContent(content); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateWithVersion(ctx, id, title, content, visibility, userID, note); err != nil {
		if errors.Is(err, port.ErrVersionConflict) {
			return nil, &Error{Code: errcode.ErrDocVersionConflict}
		}
		return nil, err
	}
	doc.Title = title
	doc.Content = content
	doc.CurrentVersion++
	if visibility != "" {
		doc.Visibility = visibility
	}
	return doc, nil
}

// Publish 发布 draft → published（作者或 admin）。
func (s *DocService) Publish(ctx context.Context, id uint, userID uint, isAdmin bool) error {
	return s.setStatus(ctx, id, domain.DocStatusPublished, userID, isAdmin)
}

// Unpublish 撤稿 published → draft（作者或 admin）。
func (s *DocService) Unpublish(ctx context.Context, id uint, userID uint, isAdmin bool) error {
	return s.setStatus(ctx, id, domain.DocStatusDraft, userID, isAdmin)
}

func (s *DocService) setStatus(ctx context.Context, id uint, status string, userID uint, isAdmin bool) error {
	doc, err := s.repo.GetDocByID(ctx, id, userID, []uint{}, true)
	if err != nil {
		return &Error{Code: errcode.ErrDocNotFound}
	}
	if !isAdmin && doc.OwnerID != userID {
		return &Error{Code: errcode.ErrDocForbidden}
	}
	var publishedAt *time.Time
	if status == domain.DocStatusPublished {
		now := time.Now()
		publishedAt = &now
	}
	return s.repo.SetDocStatus(ctx, id, status, publishedAt)
}

// ListVersions 版本历史列表（作者或 admin）。
func (s *DocService) ListVersions(ctx context.Context, docID uint, userID uint, isAdmin bool) ([]domain.DocVersion, error) {
	if _, err := s.ensureOwnerOrAdmin(ctx, docID, userID, isAdmin); err != nil {
		return nil, err
	}
	return s.repo.ListVersions(ctx, docID)
}

// GetVersion 查看指定版本内容（作者或 admin）。
func (s *DocService) GetVersion(ctx context.Context, docID uint, versionNo int, userID uint, isAdmin bool) (*domain.DocVersion, error) {
	if _, err := s.ensureOwnerOrAdmin(ctx, docID, userID, isAdmin); err != nil {
		return nil, err
	}
	ver, err := s.repo.GetVersion(ctx, docID, versionNo)
	if err != nil {
		return nil, &Error{Code: errcode.ErrDocVersionNotFound}
	}
	return ver, nil
}

// Restore 还原到指定版本：追加新版本（内容=指定版本）+ 覆盖当前内容。
func (s *DocService) Restore(ctx context.Context, docID uint, versionNo int, userID uint, isAdmin bool) error {
	if _, err := s.ensureOwnerOrAdmin(ctx, docID, userID, isAdmin); err != nil {
		return err
	}
	if _, err := s.repo.GetVersion(ctx, docID, versionNo); err != nil {
		return &Error{Code: errcode.ErrDocVersionNotFound}
	}
	note := fmt.Sprintf("还原到版本 %d", versionNo)
	if err := s.repo.RestoreToVersion(ctx, docID, versionNo, userID, note); err != nil {
		if errors.Is(err, port.ErrVersionConflict) {
			return &Error{Code: errcode.ErrDocVersionConflict}
		}
		return err
	}
	return nil
}

func (s *DocService) DeleteDoc(ctx context.Context, id uint, userID uint, isAdmin bool) error {
	doc, err := s.repo.GetDocByID(ctx, id, userID, []uint{}, true)
	if err != nil {
		return &Error{Code: errcode.ErrDocNotFound}
	}
	if !isAdmin && doc.OwnerID != userID {
		return &Error{Code: errcode.ErrDocForbidden}
	}
	return s.repo.DeleteDoc(ctx, id)
}

func (s *DocService) ensureOwnerOrAdmin(ctx context.Context, docID uint, userID uint, isAdmin bool) (*domain.Document, error) {
	doc, err := s.repo.GetDocByID(ctx, docID, userID, []uint{}, true)
	if err != nil {
		return nil, &Error{Code: errcode.ErrDocNotFound}
	}
	if !isAdmin && doc.OwnerID != userID {
		return nil, &Error{Code: errcode.ErrDocForbidden}
	}
	return doc, nil
}

// ———— 可见性与树工具 ————

// canSeeDir 目录可见：admin 直通；shared 目录所有登录用户可见；private 仅 admin。
func canSeeDir(dir *domain.DocDirectory, _ []uint, isAdmin bool) bool {
	if isAdmin {
		return true
	}
	return dir != nil && dir.Visibility == domain.VisibilityShared
}

// filterTree 按可见性过滤树：不可见的节点连同其子树整体裁剪。
func filterTree(tree []domain.DocDirectory, roleIDs []uint, isAdmin bool) []domain.DocDirectory {
	var out []domain.DocDirectory
	for _, n := range tree {
		if !canSeeDir(&n, roleIDs, isAdmin) {
			continue
		}
		n.Children = filterTree(n.Children, roleIDs, isAdmin)
		out = append(out, n)
	}
	return out
}

// filterPublicTree 保留 shared 目录（private 目录连同子树裁掉），并把公开文档挂到对应目录。
func filterPublicTree(tree []domain.DocDirectory, publicDocs []domain.Document) []domain.DocDirectory {
	byDir := make(map[uint][]domain.DocTreeItem, len(publicDocs))
	for _, d := range publicDocs {
		byDir[d.DirID] = append(byDir[d.DirID], domain.DocTreeItem{
			ID: d.ID, Title: d.Title, Status: d.Status, Visibility: d.Visibility,
		})
	}
	var walk func([]domain.DocDirectory) []domain.DocDirectory
	walk = func(nodes []domain.DocDirectory) []domain.DocDirectory {
		var out []domain.DocDirectory
		for _, n := range nodes {
			if n.Visibility != domain.VisibilityShared {
				continue
			}
			n.Docs = byDir[n.ID]
			n.Children = walk(n.Children)
			out = append(out, n)
		}
		return out
	}
	return walk(tree)
}

func buildTree(dirs []domain.DocDirectory, parentID uint) []domain.DocDirectory {
	var result []domain.DocDirectory
	for _, d := range dirs {
		if d.ParentID == parentID {
			d.Children = buildTree(dirs, d.ID)
			result = append(result, d)
		}
	}
	return result
}

// 文档内容的资源上限：防止单文档过大 / 恶意深嵌套拖垮前端渲染与预览。
const (
	docContentMaxBytes   = 2 << 20 // 2MB
	docContentMaxNodes   = 20000
	docContentMaxDepth   = 20
	docContentMaxTextLen = 1 << 20 // 单文本节点 1MB
)

// 允许的 Lexical 节点类型白名单（前端编辑器可产出的全部节点）。
// 渲染侧只认这些类型，其余一律按纯文本/丢弃降级，杜绝未知节点夹带危险内容。
var docAllowedNodeTypes = map[string]bool{
	"root": true, "paragraph": true, "text": true, "heading": true, "quote": true,
	"code": true, "code-highlight": true, "list": true, "listitem": true, "link": true,
	"table": true, "tablerow": true, "tablecell": true, "horizontalrule": true,
	"linebreak": true, "callout": true, "toggle": true, "image": true,
}

// validateDocContent 校验文档内容为合法的 Lexical editorState JSON。
// content 是 JSON 数据结构（非 HTML），渲染走 React 转义 + 预览序列化器白名单，
// 因此这里不做 HTML 清洗，只做结构/资源校验：
// 长度上限、合法 JSON、node type 白名单、节点数/深度/单文本长度上限。
func validateDocContent(content string) *Error {
	if len(content) > docContentMaxBytes {
		return &Error{Code: errcode.ErrDocContentInvalid}
	}
	if strings.TrimSpace(content) == "" {
		return nil // 允许空文档
	}
	var state struct {
		Root *struct {
			Children []json.RawMessage `json:"children"`
		} `json:"root"`
	}
	if err := json.Unmarshal([]byte(content), &state); err != nil {
		return &Error{Code: errcode.ErrDocContentInvalid}
	}
	if state.Root == nil {
		return &Error{Code: errcode.ErrDocContentInvalid}
	}
	count := 0
	var walk func(raw json.RawMessage, depth int) bool
	walk = func(raw json.RawMessage, depth int) bool {
		if depth > docContentMaxDepth {
			return false
		}
		var node struct {
			Type     string            `json:"type"`
			Children []json.RawMessage `json:"children"`
			Text     string            `json:"text"`
		}
		if err := json.Unmarshal(raw, &node); err != nil {
			return false
		}
		count++
		if count > docContentMaxNodes {
			return false
		}
		if !docAllowedNodeTypes[node.Type] {
			return false
		}
		if len(node.Text) > docContentMaxTextLen {
			return false
		}
		for _, c := range node.Children {
			if !walk(c, depth+1) {
				return false
			}
		}
		return true
	}
	for _, c := range state.Root.Children {
		if !walk(c, 1) {
			return &Error{Code: errcode.ErrDocContentInvalid}
		}
	}
	return nil
}
