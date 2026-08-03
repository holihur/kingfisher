package port

import (
	"context"

	"kingfisher/extends/audit/domain"
)

type AuditRepository interface {
	InsertBatch(ctx context.Context, logs []domain.AuditLog) error
	FindAll(ctx context.Context, page, pageSize int, filters map[string]any) ([]domain.AuditLog, int64, error)
}
