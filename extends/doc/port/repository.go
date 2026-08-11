package port

import (
	"context"
	"errors"
	"time"

	"kingfisher/core/query"
	"kingfisher/extends/doc/domain"
)

// ErrVersionConflict 并发保存同一文档导致版本号冲突（唯一索引 (doc_id, version_no) 兜底），
// service 层应返回重试提示。
var ErrVersionConflict = errors.New("doc version conflict")

// DocRepository 文档模块仓储接口。
// 文档的读写方法均接收当前用户可见性上下文（userID/roleIDs/isAdmin），
// 在 SQL 层完成可见性过滤（目录默认拒绝 + draft 仅作者 + private 仅作者）。
type DocRepository interface {
	// —— 目录 ——
	FindAllDirs(ctx context.Context) ([]domain.DocDirectory, error)
	FindDirByID(ctx context.Context, id uint) (*domain.DocDirectory, error)
	CreateDir(ctx context.Context, d *domain.DocDirectory) error
	UpdateDir(ctx context.Context, id uint, updates map[string]any) error
	DeleteDir(ctx context.Context, id uint) error // 顺带删除 doc_dir_roles 授权行
	HasDirChildren(ctx context.Context, parentID uint) (bool, error)
	HasDirDocuments(ctx context.Context, dirID uint) (bool, error)
	GetDirRoleIDs(ctx context.Context, dirID uint) ([]uint, error)
	SetDirRoles(ctx context.Context, dirID uint, roleIDs []uint) error // 全量替换

	// —— 文档 ——
	ListDocs(ctx context.Context, dirID uint, q *query.Query, userID uint, roleIDs []uint, isAdmin bool) ([]domain.Document, int64, error)
	// ListAllVisibleDocs 返回当前用户可见的全部文档（不分页，用于目录树叶子节点；可见性同 ListDocs）
	ListAllVisibleDocs(ctx context.Context, userID uint, roleIDs []uint, isAdmin bool) ([]domain.Document, error)
	GetDocByID(ctx context.Context, id uint, userID uint, roleIDs []uint, isAdmin bool) (*domain.Document, error)
	CreateWithVersion(ctx context.Context, doc *domain.Document, ver *domain.DocVersion) (*domain.Document, error)
	UpdateWithVersion(ctx context.Context, id uint, title, content string, ownerID uint, note string) error
	RestoreToVersion(ctx context.Context, docID uint, fromVersionNo int, ownerID uint, note string) error
	SetDocStatus(ctx context.Context, id uint, status string, publishedAt *time.Time) error
	DeleteDoc(ctx context.Context, id uint) error // 级联删除 doc_versions

	// —— 版本 ——
	ListVersions(ctx context.Context, docID uint) ([]domain.DocVersion, error)
	GetVersion(ctx context.Context, docID uint, versionNo int) (*domain.DocVersion, error)
}
