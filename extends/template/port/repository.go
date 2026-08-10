package port

import (
	"context"

	"kingfisher/core/query"
	"kingfisher/extends/template/domain"
)

// TemplateRepository 模版仓库接口
type TemplateRepository interface {
	List(ctx context.Context, q *query.Query) ([]domain.Template, int64, error)
	GetByID(ctx context.Context, id uint) (*domain.Template, error)
	GetByCode(ctx context.Context, code string) (*domain.Template, error)
	Create(ctx context.Context, name, code, templateType, title, content string, status int, remark, version string) (*domain.Template, error)
	Update(ctx context.Context, id uint, name, code, templateType, title, content string, status int, remark, version string) error
	Delete(ctx context.Context, id uint) error
	DeleteBatch(ctx context.Context, ids []uint) error
	UpdateStatusBatch(ctx context.Context, ids []uint, status int) error
}
