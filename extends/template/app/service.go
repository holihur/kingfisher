package app

import (
	"context"

	"kingfisher/core/errcode"
	"kingfisher/core/query"
	"kingfisher/extends/template/domain"
	"kingfisher/extends/template/port"
)

// Error 携带 errcode 的错误类型，handler 层据此映射到 HTTP 错误码
type Error struct{ Code int }

func (e *Error) Error() string { return errcode.Msg(e.Code) }

// TemplateService 模版服务
type TemplateService struct {
	repo port.TemplateRepository
}

func NewTemplateService(repo port.TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

func (s *TemplateService) List(ctx context.Context, q *query.Query) ([]domain.Template, int64, error) {
	return s.repo.List(ctx, q)
}

func (s *TemplateService) GetByID(ctx context.Context, id uint) (*domain.Template, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, &Error{Code: errcode.ErrTemplateNotFound}
	}
	return t, nil
}

func (s *TemplateService) Create(ctx context.Context, name, code, templateType, title, content string, status int, remark, version string) (*domain.Template, error) {
	if _, err := s.repo.GetByCode(ctx, code); err == nil {
		return nil, &Error{Code: errcode.ErrTemplateCodeExists}
	}
	return s.repo.Create(ctx, name, code, templateType, title, content, status, remark, version)
}

func (s *TemplateService) Update(ctx context.Context, id uint, name, code, templateType, title, content string, status int, remark, version string) error {
	existing, err := s.repo.GetByCode(ctx, code)
	if err == nil && existing.ID != id {
		return &Error{Code: errcode.ErrTemplateCodeExists}
	}
	return s.repo.Update(ctx, id, name, code, templateType, title, content, status, remark, version)
}

func (s *TemplateService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *TemplateService) BatchDelete(ctx context.Context, ids []uint) error {
	return s.repo.DeleteBatch(ctx, ids)
}

func (s *TemplateService) BatchUpdateStatus(ctx context.Context, ids []uint, status int) error {
	return s.repo.UpdateStatusBatch(ctx, ids, status)
}
