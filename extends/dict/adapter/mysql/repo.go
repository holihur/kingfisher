package adapter

import (
	"context"

	"gorm.io/gorm"

	"kingfisher/core/query"
	"kingfisher/extends/dict/domain"
)

// ---- DictType repo ----

type DictTypeRepo struct{ db *gorm.DB }

func NewDictTypeRepo(db *gorm.DB) *DictTypeRepo { return &DictTypeRepo{db: db} }

func (r *DictTypeRepo) List(ctx context.Context, q *query.Query) ([]domain.DictType, int64, error) {
	var pos []dictTypePO
	total, err := q.Find(r.db.WithContext(ctx).Model(&dictTypePO{}), &pos)
	if err != nil {
		return nil, 0, err
	}
	return toDictTypeList(pos), total, nil
}

func (r *DictTypeRepo) GetByID(ctx context.Context, id uint) (*domain.DictType, error) {
	var po dictTypePO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
	if err != nil {
		return nil, err
	}
	return toDictType(&po), nil
}

func (r *DictTypeRepo) GetByCode(ctx context.Context, code string) (*domain.DictType, error) {
	var po dictTypePO
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error
	if err != nil {
		return nil, err
	}
	return toDictType(&po), nil
}

func (r *DictTypeRepo) Create(ctx context.Context, code, name string, isPublic bool, status int, remark, version string) (*domain.DictType, error) {
	po := dictTypePO{Code: code, Name: name, IsPublic: isPublic, Status: status, Remark: remark, Version: version}
	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return nil, err
	}
	return toDictType(&po), nil
}

func (r *DictTypeRepo) Update(ctx context.Context, id uint, code, name string, isPublic bool, status int, remark, version string) error {
	return r.db.WithContext(ctx).Model(&dictTypePO{}).Where("id = ?", id).Updates(map[string]any{
		"code":      code,
		"name":      name,
		"is_public": isPublic,
		"status":    status,
		"remark":    remark,
		"version":   version,
	}).Error
}

func (r *DictTypeRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&dictTypePO{}).Error
}

func (r *DictTypeRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&dictTypePO{}).Error
}

func (r *DictTypeRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	return r.db.WithContext(ctx).Model(&dictTypePO{}).Where("id IN ?", ids).Update("status", status).Error
}

// ---- DictEntry repo ----

type DictEntryRepo struct{ db *gorm.DB }

func NewDictEntryRepo(db *gorm.DB) *DictEntryRepo { return &DictEntryRepo{db: db} }

func (r *DictEntryRepo) ListByTypeID(ctx context.Context, typeID uint, q *query.Query) ([]domain.DictEntry, int64, error) {
	var pos []dictEntryPO
	total, err := q.Find(r.db.WithContext(ctx).Model(&dictEntryPO{}).Where("type_id = ?", typeID), &pos)
	if err != nil {
		return nil, 0, err
	}
	return toDictEntryList(pos), total, nil
}

func (r *DictEntryRepo) ListByTypeCode(ctx context.Context, code string) ([]domain.DictEntry, error) {
	var pos []dictEntryPO
	// JOIN dict_types，仅查询公开+启用的类型下的启用条目
	err := r.db.WithContext(ctx).
		Table("dict_entries").
		Select("dict_entries.*").
		Joins("JOIN dict_types ON dict_types.id = dict_entries.type_id").
		Where("dict_types.code = ? AND dict_types.is_public = ? AND dict_types.status = ? AND dict_entries.status = ?",
			code, true, 1, 1).
		Order("dict_entries.sort ASC, dict_entries.id ASC").
		Find(&pos).Error
	if err != nil {
		return nil, err
	}
	return toDictEntryList(pos), nil
}

func (r *DictEntryRepo) GetByID(ctx context.Context, id uint) (*domain.DictEntry, error) {
	var po dictEntryPO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
	if err != nil {
		return nil, err
	}
	return toDictEntry(&po), nil
}

func (r *DictEntryRepo) Create(ctx context.Context, typeID uint, label, value string, sort, status int, remark, version string) (*domain.DictEntry, error) {
	po := dictEntryPO{TypeID: typeID, Label: label, Value: value, Sort: sort, Status: status, Remark: remark, Version: version}
	if err := r.db.WithContext(ctx).Create(&po).Error; err != nil {
		return nil, err
	}
	return toDictEntry(&po), nil
}

func (r *DictEntryRepo) Update(ctx context.Context, id uint, typeID uint, label, value string, sort, status int, remark, version string) error {
	return r.db.WithContext(ctx).Model(&dictEntryPO{}).Where("id = ?", id).Updates(map[string]any{
		"type_id": typeID,
		"label":   label,
		"value":   value,
		"sort":    sort,
		"status":  status,
		"remark":  remark,
		"version": version,
	}).Error
}

func (r *DictEntryRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&dictEntryPO{}).Error
}

func (r *DictEntryRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&dictEntryPO{}).Error
}

func (r *DictEntryRepo) UpdateStatusBatch(ctx context.Context, ids []uint, status int) error {
	return r.db.WithContext(ctx).Model(&dictEntryPO{}).Where("id IN ?", ids).Update("status", status).Error
}

func (r *DictEntryRepo) DeleteByTypeID(ctx context.Context, typeID uint) error {
	return r.db.WithContext(ctx).Where("type_id = ?", typeID).Delete(&dictEntryPO{}).Error
}

// ---- helpers ----

func toDictType(p *dictTypePO) *domain.DictType {
	return &domain.DictType{
		ID:        p.ID,
		Code:      p.Code,
		Name:      p.Name,
		IsPublic:  p.IsPublic,
		Status:    p.Status,
		Remark:    p.Remark,
		Version:   p.Version,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toDictTypeList(pos []dictTypePO) []domain.DictType {
	out := make([]domain.DictType, len(pos))
	for i, p := range pos {
		out[i] = *toDictType(&p)
	}
	return out
}

func toDictEntry(p *dictEntryPO) *domain.DictEntry {
	return &domain.DictEntry{
		ID:        p.ID,
		TypeID:    p.TypeID,
		Label:     p.Label,
		Value:     p.Value,
		Sort:      p.Sort,
		Status:    p.Status,
		Remark:    p.Remark,
		Version:   p.Version,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toDictEntryList(pos []dictEntryPO) []domain.DictEntry {
	out := make([]domain.DictEntry, len(pos))
	for i, p := range pos {
		out[i] = *toDictEntry(&p)
	}
	return out
}
