package adapter
import ("context"; "gorm.io/gorm"; "kingfisher/extends/audit/domain")
type AuditRepo struct{ db *gorm.DB }
func NewAuditRepo(db *gorm.DB) *AuditRepo { return &AuditRepo{db: db} }
func (r *AuditRepo) InsertBatch(ctx context.Context, logs []domain.AuditLog) error {
    pos := make([]auditPO, len(logs))
    for i, l := range logs { pos[i] = auditPO{UserID: l.UserID, Username: l.Username, Action: l.Action, Resource: l.Resource, ResourceID: l.ResourceID, IP: l.IP, UserAgent: l.UserAgent} }
    return r.db.WithContext(ctx).Create(&pos).Error
}
func (r *AuditRepo) FindAll(ctx context.Context, page, pageSize int, filters map[string]any) ([]domain.AuditLog, int64, error) {
    var pos []auditPO; var total int64
    q := r.db.WithContext(ctx).Model(&auditPO{})
    if v, ok := filters["user_id"]; ok { q = q.Where("user_id = ?", v) }
    if v, ok := filters["resource"]; ok { q = q.Where("resource = ?", v) }
    if v, ok := filters["action"]; ok { q = q.Where("action = ?", v) }
    q.Count(&total)
    offset := (page-1)*pageSize; if offset<0 { offset=0 }
    err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&pos).Error; if err != nil { return nil, 0, err }
    logs := make([]domain.AuditLog, len(pos))
    for i, p := range pos { logs[i] = domain.AuditLog{ID: p.ID, UserID: p.UserID, Username: p.Username, Action: p.Action, Resource: p.Resource, ResourceID: p.ResourceID, IP: p.IP, UserAgent: p.UserAgent, CreatedAt: p.CreatedAt} }
    return logs, total, nil
}
