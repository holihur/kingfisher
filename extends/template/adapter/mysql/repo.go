package adapter

import (
	"context"

	"gorm.io/gorm"

	"kingfisher/core/query"
	"kingfisher/extends/template/domain"
)

type TemplateRepo struct{ db *gorm.DB }

func NewTemplateRepo(db *gorm.DB) *TemplateRepo { return &TemplateRepo{db: db} }

func (r *TemplateRepo) List(ctx context.Context, q *query.Query) ([]domain.Template, int64, error) {
	var pos []templatePO
	total, err := q.Find(r.db.WithContext(ctx).Model(&templatePO{}), &pos)
	if err != nil {
		return nil, 0, err
	}
	return toTemplateList(pos), total, nil
}

func (r *TemplateRepo) GetByID(ctx context.Context, id uint) (*domain.Template, error) {
	var po templatePO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
	if err != nil {
		return nil, err
	}
	return toTemplate(&po), nil
}

func (r *TemplateRepo) GetByCode(ctx context.Context, code string) (*domain.Template, error) {
	var po templatePO
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error
	if err != nil {
		return nil, err
	}
	return toTemplate(&po), nil
}

func (r *TemplateRepo) Create(ctx context.Context, name, code, templateType, title, content string, status int, remark, version string) (*domain.Template, error) {
	po := templatePO{
		Name: name, Code: code, TemplateType: templateType,
		Title: title, Content: content, Status: status, Remark: remark, Version: version,
	}
	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return nil, err
	}
	return toTemplate(&po), nil
}

func (r *TemplateRepo) Update(ctx context.Context, id uint, name, code, templateType, title, content string, status int, remark, version string) error {
	return r.db.WithContext(ctx).Model(&templatePO{}).Where("id = ?", id).Updates(map[string]any{
		"name":          name,
		"code":          code,
		"template_type": templateType,
		"title":         title,
		"content":       content,
		"status":        status,
		"remark":        remark,
		"version":       version,
	}).Error
}

func (r *TemplateRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&templatePO{}).Error
}

func (r *TemplateRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&templatePO{}).Error
}

func (r *TemplateRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	return r.db.WithContext(ctx).Model(&templatePO{}).Where("id IN ?", ids).Update("status", status).Error
}

func toTemplate(p *templatePO) *domain.Template {
	return &domain.Template{
		ID:           p.ID,
		Name:         p.Name,
		Code:         p.Code,
		TemplateType: p.TemplateType,
		Title:        p.Title,
		Content:      p.Content,
		Status:       p.Status,
		Remark:       p.Remark,
		Version:      p.Version,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func toTemplateList(pos []templatePO) []domain.Template {
	out := make([]domain.Template, len(pos))
	for i, p := range pos {
		out[i] = *toTemplate(&p)
	}
	return out
}
