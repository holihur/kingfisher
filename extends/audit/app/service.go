package app
import ("context"; "time"; "kingfisher/core/logger"; "kingfisher/extends/audit/domain"; "kingfisher/extends/audit/adapter/mysql")
type AuditService struct {
    repo *adapter.AuditRepo; buffer chan *domain.AuditLog
}
func NewAuditService(repo *adapter.AuditRepo) *AuditService {
    s := &AuditService{repo: repo, buffer: make(chan *domain.AuditLog, 1000)}
    go s.worker()
    return s
}
func (s *AuditService) Log(ctx context.Context, l *domain.AuditLog) {
    select { case s.buffer <- l: default: logger.Get().Warn("audit buffer full, dropping log") }
}
func (s *AuditService) worker() {
    batch := make([]domain.AuditLog, 0, 50); ticker := time.NewTicker(2 * time.Second)
    for {
        select {
        case l := <-s.buffer: batch = append(batch, *l); if len(batch) >= 50 { s.repo.InsertBatch(context.Background(), batch); batch = batch[:0] }
        case <-ticker.C: if len(batch) > 0 { s.repo.InsertBatch(context.Background(), batch); batch = batch[:0] }
        }
    }
}
func (s *AuditService) Flush() {
    // flush is called from Shutdown — not implemented for simplicity
}
func (s *AuditService) FindAll(ctx context.Context, page, pageSize int, filters map[string]any) ([]domain.AuditLog, int64, error) {
    return s.repo.FindAll(ctx, page, pageSize, filters)
}
