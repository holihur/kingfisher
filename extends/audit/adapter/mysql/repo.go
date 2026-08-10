package adapter

import (
	"context"

	"gorm.io/gorm"

	"kingfisher/core/query"
	"kingfisher/extends/audit/domain"
)

type AuditRepo struct{ db *gorm.DB }

func NewAuditRepo(db *gorm.DB) *AuditRepo { return &AuditRepo{db: db} }
func (r *AuditRepo) InsertBatch(ctx context.Context, logs []domain.AuditLog) error {
	pos := make([]auditPO, len(logs))
	for i, l := range logs {
		pos[i] = auditPO{UserID: l.UserID, Username: l.Username, Action: l.Action, Resource: l.Resource, ResourceID: l.ResourceID, Detail: l.Detail, Result: l.Result, Latency: l.Latency, Message: l.Message, IP: l.IP, UserAgent: l.UserAgent}
	}
	return r.db.WithContext(ctx).Create(&pos).Error
}
func (r *AuditRepo) FindAll(ctx context.Context, q *query.Query) ([]domain.AuditLog, int64, error) {
	var pos []auditPO
	total, err := q.Find(r.db.WithContext(ctx).Model(&auditPO{}), &pos)
	if err != nil {
		return nil, 0, err
	}
	logs := make([]domain.AuditLog, len(pos))
	for i, p := range pos {
		logs[i] = domain.AuditLog{ID: p.ID, UserID: p.UserID, Username: p.Username, Action: p.Action, Resource: p.Resource, ResourceID: p.ResourceID, Detail: p.Detail, Result: p.Result, Latency: p.Latency, Message: p.Message, IP: p.IP, UserAgent: p.UserAgent, CreatedAt: p.CreatedAt}
	}
	return logs, total, nil
}
