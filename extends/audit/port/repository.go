package port

import (
	"context"

	"kingfisher/core/query"
	"kingfisher/extends/audit/domain"
)

type AuditRepository interface {
	InsertBatch(ctx context.Context, logs []domain.AuditLog) error
	FindAll(ctx context.Context, q *query.Query) ([]domain.AuditLog, int64, error)
}
