package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/microcosm-cc/bluemonday"

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
	repo     port.DocRepository
	cache    coreCache.Cache
	sanitize *bluemonday.Policy
}

func NewDocService(repo port.DocRepository, cache coreCache.Cache) *DocService {
	return &DocService{repo: repo, cache: cache, sanitize: newSanitizer()}
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

func (s *DocService) GetDirRoleIDs(ctx context.Context, dirID uint) ([]uint, error) {
	return s.repo.GetDirRoleIDs(ctx, dirID)
}

// SetDirRoles 全量替换目录可见角色；设置后目录仅对这些角色可见（默认拒绝）。
func (s *DocService) SetDirRoles(ctx context.Context, dirID uint, roleIDs []uint) error {
	if err := s.repo.SetDirRoles(ctx, dirID, roleIDs); err != nil {
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
	clean := s.sanitize.Sanitize(content)
	doc := &domain.Document{
		DirID: dirID, Title: title, Content: clean, OwnerID: ownerID,
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
	clean := s.sanitize.Sanitize(content)
	if err := s.repo.UpdateWithVersion(ctx, id, title, clean, visibility, userID, note); err != nil {
		if errors.Is(err, port.ErrVersionConflict) {
			return nil, &Error{Code: errcode.ErrDocVersionConflict}
		}
		return nil, err
	}
	doc.Title = title
	doc.Content = clean
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

// canSeeDir 目录可见（默认拒绝）：admin 直通；有授权记录且角色交集非空才可见。
func canSeeDir(dir *domain.DocDirectory, roleIDs []uint, isAdmin bool) bool {
	if isAdmin {
		return true
	}
	if dir == nil || len(dir.GrantedRoles) == 0 {
		return false
	}
	for _, gid := range dir.GrantedRoles {
		for _, rid := range roleIDs {
			if gid == rid {
				return true
			}
		}
	}
	return false
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

// newSanitizer 构造富文本白名单清洗策略（XSS 防护：删除 script/事件属性/javascript: 协议）。
func newSanitizer() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	// Quill 常用元素（UGCPolicy 已含 p/ul/ol/li/a/img 等，补充标题/代码块/对齐）
	p.AllowElements("h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "pre", "code", "span", "br", "strike", "sub", "sup")
	p.AllowAttrs("align", "class", "style").OnElements("p", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "span", "div", "pre", "code")
	p.AllowAttrs("href", "target", "rel").OnElements("a")
	p.AllowAttrs("src", "alt", "title", "width", "height").OnElements("img")
	return p
}
