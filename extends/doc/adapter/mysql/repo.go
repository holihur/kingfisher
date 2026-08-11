package adapter

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"kingfisher/core/query"
	"kingfisher/extends/doc/domain"
	"kingfisher/extends/doc/port"
)

type DocRepo struct{ db *gorm.DB }

func NewDocRepo(db *gorm.DB) *DocRepo { return &DocRepo{db: db} }

// —— 目录 ——

func (r *DocRepo) FindAllDirs(ctx context.Context) ([]domain.DocDirectory, error) {
	var pos []docDirectoryPO
	if err := r.db.WithContext(ctx).Order("sort ASC").Find(&pos).Error; err != nil {
		return nil, err
	}
	// 批量加载目录授权，避免 N+1
	var grants []docDirRolePO
	if err := r.db.WithContext(ctx).Find(&grants).Error; err != nil {
		return nil, err
	}
	byDir := map[uint][]uint{}
	for _, g := range grants {
		byDir[g.DirID] = append(byDir[g.DirID], g.RoleID)
	}
	dirs := make([]domain.DocDirectory, len(pos))
	for i, p := range pos {
		dirs[i] = domain.DocDirectory{
			ID: p.ID, ParentID: p.ParentID, Name: p.Name, Sort: p.Sort,
			Status: p.Status, Version: p.Version, GrantedRoles: byDir[p.ID],
			CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
		}
	}
	return dirs, nil
}

func (r *DocRepo) FindDirByID(ctx context.Context, id uint) (*domain.DocDirectory, error) {
	var po docDirectoryPO
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error; err != nil {
		return nil, err
	}
	roleIDs, err := r.GetDirRoleIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.DocDirectory{
		ID: po.ID, ParentID: po.ParentID, Name: po.Name, Sort: po.Sort,
		Status: po.Status, Version: po.Version, GrantedRoles: roleIDs,
		CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt,
	}, nil
}

func (r *DocRepo) CreateDir(ctx context.Context, d *domain.DocDirectory) error {
	return r.db.WithContext(ctx).Create(&docDirectoryPO{
		ParentID: d.ParentID, Name: d.Name, Sort: d.Sort, Status: d.Status, Version: d.Version,
	}).Error
}

func (r *DocRepo) UpdateDir(ctx context.Context, id uint, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&docDirectoryPO{}).Where("id = ?", id).Updates(updates).Error
}

func (r *DocRepo) DeleteDir(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dir_id = ?", id).Delete(&docDirRolePO{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&docDirectoryPO{}).Error
	})
}

func (r *DocRepo) HasDirChildren(ctx context.Context, parentID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&docDirectoryPO{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count > 0, err
}

func (r *DocRepo) HasDirDocuments(ctx context.Context, dirID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&documentPO{}).Where("dir_id = ?", dirID).Count(&count).Error
	return count > 0, err
}

func (r *DocRepo) GetDirRoleIDs(ctx context.Context, dirID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&docDirRolePO{}).Where("dir_id = ?", dirID).Pluck("role_id", &ids).Error
	return ids, err
}

func (r *DocRepo) SetDirRoles(ctx context.Context, dirID uint, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dir_id = ?", dirID).Delete(&docDirRolePO{}).Error; err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		rows := make([]docDirRolePO, 0, len(roleIDs))
		for _, rid := range roleIDs {
			rows = append(rows, docDirRolePO{DirID: dirID, RoleID: rid})
		}
		return tx.Create(&rows).Error
	})
}

// —— 文档 ——

// visibleScope 复合可见性 WHERE（列表与单条读取共用，避免两处实现漂移）。
// 非管理员：作者直接可见；否则须 status=published AND visibility=shared 且目录可见（默认拒绝，EXISTS 授权）。
func (r *DocRepo) visibleScope(userID uint, roleIDs []uint, isAdmin bool) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isAdmin {
			return db
		}
		dirVisible := "EXISTS (SELECT 1 FROM doc_dir_roles dr WHERE dr.dir_id = documents.dir_id AND dr.role_id IN ?)"
		return db.Where(
			"(documents.owner_id = ? OR (documents.status = ? AND documents.visibility = ? AND "+dirVisible+"))",
			userID, domain.DocStatusPublished, domain.VisibilityShared, roleIDs,
		)
	}
}

func (r *DocRepo) ListDocs(ctx context.Context, dirID uint, q *query.Query, userID uint, roleIDs []uint, isAdmin bool) ([]domain.Document, int64, error) {
	var pos []documentPO
	base := r.db.WithContext(ctx).Model(&documentPO{}).Where("dir_id = ?", dirID).Scopes(r.visibleScope(userID, roleIDs, isAdmin))
	total, err := q.Find(base, &pos)
	if err != nil {
		return nil, 0, err
	}
	docs := toDocumentList(pos)
	if err := r.attachDocOwnerNames(ctx, docs); err != nil {
		return nil, 0, err
	}
	return docs, total, nil
}

// ListAllVisibleDocs 当前用户可见的全部文档（不分页，按 sort/id 排序；目录树叶子用）
func (r *DocRepo) ListAllVisibleDocs(ctx context.Context, userID uint, roleIDs []uint, isAdmin bool) ([]domain.Document, error) {
	var pos []documentPO
	if err := r.db.WithContext(ctx).Model(&documentPO{}).
		Scopes(r.visibleScope(userID, roleIDs, isAdmin)).
		Order("sort ASC").Order("id ASC").
		Find(&pos).Error; err != nil {
		return nil, err
	}
	docs := toDocumentList(pos)
	if err := r.attachDocOwnerNames(ctx, docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *DocRepo) GetDocByID(ctx context.Context, id uint, userID uint, roleIDs []uint, isAdmin bool) (*domain.Document, error) {
	var po documentPO
	if err := r.db.WithContext(ctx).Model(&documentPO{}).
		Where("documents.id = ?", id).
		Scopes(r.visibleScope(userID, roleIDs, isAdmin)).
		First(&po).Error; err != nil {
		return nil, err
	}
	doc := toDocument(&po)
	if err := r.attachDocOwnerNames(ctx, []domain.Document{*doc}); err != nil {
		return nil, err
	}
	return doc, nil
}

// GetPublicDoc 公开文档：已发布 + 共享，不受目录角色白名单限制（匿名可读）。
func (r *DocRepo) GetPublicDoc(ctx context.Context, id uint) (*domain.Document, error) {
	var po documentPO
	if err := r.db.WithContext(ctx).Model(&documentPO{}).
		Where("id = ? AND status = ? AND visibility = ?", id, domain.DocStatusPublished, domain.VisibilityShared).
		First(&po).Error; err != nil {
		return nil, err
	}
	doc := toDocument(&po)
	if err := r.attachDocOwnerNames(ctx, []domain.Document{*doc}); err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *DocRepo) CreateWithVersion(ctx context.Context, doc *domain.Document, ver *domain.DocVersion) (*domain.Document, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		po := &documentPO{
			DirID: doc.DirID, Title: doc.Title, Content: doc.Content,
			OwnerID: doc.OwnerID, Visibility: doc.Visibility, Status: doc.Status,
			CurrentVersion: 1, Sort: doc.Sort,
		}
		if err := tx.Create(po).Error; err != nil {
			return err
		}
		doc.ID = po.ID
		doc.CurrentVersion = 1
		vpo := &docVersionPO{
			DocID: po.ID, VersionNo: 1, Title: po.Title, Content: po.Content,
			OwnerID: po.OwnerID, Note: ver.Note,
		}
		if err := tx.Create(vpo).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *DocRepo) UpdateWithVersion(ctx context.Context, id uint, title, content string, ownerID uint, note string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cur documentPO
		if err := tx.Where("id = ?", id).First(&cur).Error; err != nil {
			return err
		}
		next := cur.CurrentVersion + 1
		vpo := &docVersionPO{DocID: id, VersionNo: next, Title: title, Content: content, OwnerID: ownerID, Note: note}
		if err := tx.Create(vpo).Error; err != nil {
			return r.versionErr(err)
		}
		if err := tx.Model(&documentPO{}).Where("id = ?", id).
			Updates(map[string]any{"title": title, "content": content, "current_version": next}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *DocRepo) RestoreToVersion(ctx context.Context, docID uint, fromVersionNo int, ownerID uint, note string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cur documentPO
		if err := tx.Where("id = ?", docID).First(&cur).Error; err != nil {
			return err
		}
		var from docVersionPO
		if err := tx.Where("doc_id = ? AND version_no = ?", docID, fromVersionNo).First(&from).Error; err != nil {
			return err
		}
		next := cur.CurrentVersion + 1
		vpo := &docVersionPO{DocID: docID, VersionNo: next, Title: from.Title, Content: from.Content, OwnerID: ownerID, Note: note}
		if err := tx.Create(vpo).Error; err != nil {
			return r.versionErr(err)
		}
		if err := tx.Model(&documentPO{}).Where("id = ?", docID).
			Updates(map[string]any{"title": from.Title, "content": from.Content, "current_version": next}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *DocRepo) SetDocStatus(ctx context.Context, id uint, status string, publishedAt *time.Time) error {
	return r.db.WithContext(ctx).Model(&documentPO{}).Where("id = ?", id).
		Updates(map[string]any{"status": status, "published_at": publishedAt}).Error
}

func (r *DocRepo) DeleteDoc(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("doc_id = ?", id).Delete(&docVersionPO{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&documentPO{}).Error
	})
}

// —— 版本 ——

func (r *DocRepo) ListVersions(ctx context.Context, docID uint) ([]domain.DocVersion, error) {
	var pos []docVersionPO
	if err := r.db.WithContext(ctx).Where("doc_id = ?", docID).Order("version_no DESC").Find(&pos).Error; err != nil {
		return nil, err
	}
	vers := toVersionList(pos)
	if err := r.attachVersionOwnerNames(ctx, vers); err != nil {
		return nil, err
	}
	return vers, nil
}

func (r *DocRepo) GetVersion(ctx context.Context, docID uint, versionNo int) (*domain.DocVersion, error) {
	var po docVersionPO
	if err := r.db.WithContext(ctx).Where("doc_id = ? AND version_no = ?", docID, versionNo).First(&po).Error; err != nil {
		return nil, err
	}
	return toVersion(&po), nil
}

// —— helpers ——

// versionErr 将唯一索引冲突（并发保存同号）转为语义错误，service 层映射为 ErrDocVersionConflict。
func (r *DocRepo) versionErr(err error) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return port.ErrVersionConflict
	}
	return err
}

func (r *DocRepo) attachDocOwnerNames(ctx context.Context, docs []domain.Document) error {
	if len(docs) == 0 {
		return nil
	}
	names, err := r.loadUsernames(ctx)
	if err != nil {
		return err
	}
	for i := range docs {
		docs[i].OwnerName = names[docs[i].OwnerID]
	}
	return nil
}

func (r *DocRepo) attachVersionOwnerNames(ctx context.Context, vers []domain.DocVersion) error {
	if len(vers) == 0 {
		return nil
	}
	names, err := r.loadUsernames(ctx)
	if err != nil {
		return err
	}
	for i := range vers {
		vers[i].OwnerName = names[vers[i].OwnerID]
	}
	return nil
}

// loadUsernames 一次性加载全部用户 id→username 映射（用户量小，可接受）。
func (r *DocRepo) loadUsernames(ctx context.Context) (map[uint]string, error) {
	type row struct {
		ID       uint
		Username string
	}
	var rows []row
	if err := r.db.WithContext(ctx).Table("users").Select("id, username").Scan(&rows).Error; err != nil {
		return nil, err
	}
	names := map[uint]string{}
	for _, rw := range rows {
		names[rw.ID] = rw.Username
	}
	return names, nil
}

func toDocument(p *documentPO) *domain.Document {
	return &domain.Document{
		ID: p.ID, DirID: p.DirID, Title: p.Title, Content: p.Content,
		OwnerID: p.OwnerID, Visibility: p.Visibility, Status: p.Status,
		CurrentVersion: p.CurrentVersion, Sort: p.Sort, PublishedAt: p.PublishedAt,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func toDocumentList(pos []documentPO) []domain.Document {
	out := make([]domain.Document, len(pos))
	for i, p := range pos {
		out[i] = *toDocument(&p)
	}
	return out
}

func toVersion(p *docVersionPO) *domain.DocVersion {
	return &domain.DocVersion{
		ID: p.ID, DocID: p.DocID, VersionNo: p.VersionNo, Title: p.Title,
		Content: p.Content, OwnerID: p.OwnerID, Note: p.Note, CreatedAt: p.CreatedAt,
	}
}

func toVersionList(pos []docVersionPO) []domain.DocVersion {
	out := make([]domain.DocVersion, len(pos))
	for i, p := range pos {
		out[i] = *toVersion(&p)
	}
	return out
}
