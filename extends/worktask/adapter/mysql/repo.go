package adapter

import (
	"context"

	"gorm.io/gorm"

	"kingfisher/core/dataaccess"
	"kingfisher/core/query"
	"kingfisher/extends/worktask/domain"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) List(ctx context.Context, q *query.Query, scope dataaccess.Scope) ([]domain.Task, int64, error) {
	var rows []taskPO
	base := dataaccess.Apply(r.db.WithContext(ctx).Model(&taskPO{}), scope)
	total, err := q.Find(base, &rows)
	if err != nil {
		return nil, 0, err
	}
	return toDomainList(rows), total, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint, scope dataaccess.Scope) (*domain.Task, error) {
	var row taskPO
	query := dataaccess.Apply(r.db.WithContext(ctx).Where("id = ?", id), scope)
	if err := query.First(&row).Error; err != nil {
		return nil, err
	}
	return toDomain(&row), nil
}

func (r *Repository) Create(ctx context.Context, task *domain.Task) error {
	row := taskPO{Title: task.Title, Description: task.Description, OwnerID: task.OwnerID, DepartmentID: task.DepartmentID, Status: task.Status}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	task.ID = row.ID
	task.CreatedAt = row.CreatedAt
	task.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *Repository) Update(ctx context.Context, id uint, updates map[string]any, scope dataaccess.Scope) error {
	result := dataaccess.Apply(r.db.WithContext(ctx).Model(&taskPO{}).Where("id = ?", id), scope).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uint, scope dataaccess.Scope) error {
	result := dataaccess.Apply(r.db.WithContext(ctx).Where("id = ?", id), scope).Delete(&taskPO{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func toDomain(row *taskPO) *domain.Task {
	return &domain.Task{ID: row.ID, Title: row.Title, Description: row.Description, OwnerID: row.OwnerID, DepartmentID: row.DepartmentID, Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func toDomainList(rows []taskPO) []domain.Task {
	items := make([]domain.Task, len(rows))
	for i := range rows {
		items[i] = *toDomain(&rows[i])
	}
	return items
}
