package app

import (
	"context"

	"kingfisher/core/errcode"
	"kingfisher/core/query"
	"kingfisher/extends/dict/domain"
	"kingfisher/extends/dict/port"
)

// Error 携带 errcode 的错误类型，handler 层据此映射到 HTTP 错误码
type Error struct{ Code int }

func (e *Error) Error() string { return errcode.Msg(e.Code) }

// ---- DictTypeService ----

type DictTypeService struct {
	repo      port.DictTypeRepository
	entryRepo port.DictEntryRepository
}

func NewDictTypeService(repo port.DictTypeRepository, entryRepo port.DictEntryRepository) *DictTypeService {
	return &DictTypeService{repo: repo, entryRepo: entryRepo}
}

func (s *DictTypeService) List(ctx context.Context, q *query.Query) ([]domain.DictType, int64, error) {
	return s.repo.List(ctx, q)
}

func (s *DictTypeService) GetByID(ctx context.Context, id uint) (*domain.DictType, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DictTypeService) GetByCode(ctx context.Context, code string) (*domain.DictType, error) {
	return s.repo.GetByCode(ctx, code)
}

func (s *DictTypeService) Create(ctx context.Context, code, name string, isPublic bool, status int, remark, version string) (*domain.DictType, error) {
	_, err := s.repo.GetByCode(ctx, code)
	if err == nil {
		return nil, &Error{Code: errcode.ErrDictTypeCodeExists}
	}
	return s.repo.Create(ctx, code, name, isPublic, status, remark, version)
}

func (s *DictTypeService) Update(ctx context.Context, id uint, code, name string, isPublic bool, status int, remark, version string) error {
	existing, err := s.repo.GetByCode(ctx, code)
	if err == nil && existing.ID != id {
		return &Error{Code: errcode.ErrDictTypeCodeExists}
	}
	return s.repo.Update(ctx, id, code, name, isPublic, status, remark, version)
}

func (s *DictTypeService) Delete(ctx context.Context, id uint) error {
	_, total, err := s.entryRepo.ListByTypeID(ctx, id, &query.Query{Page: 1, PageSize: 1})
	if err != nil {
		return err
	}
	if total > 0 {
		return &Error{Code: errcode.ErrDictTypeHasEntries}
	}
	return s.repo.Delete(ctx, id)
}

// BatchDelete 批量删除字典类型：任一类型下存在条目则整批拒绝
func (s *DictTypeService) BatchDelete(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		_, total, err := s.entryRepo.ListByTypeID(ctx, id, &query.Query{Page: 1, PageSize: 1})
		if err != nil {
			return err
		}
		if total > 0 {
			return &Error{Code: errcode.ErrDictTypeHasEntries}
		}
	}
	return s.repo.DeleteBatch(ctx, ids)
}

func (s *DictTypeService) BatchUpdateStatus(ctx context.Context, ids []uint, status int) error {
	return s.repo.UpdateStatusBatch(ctx, ids, status)
}

// ---- DictEntryService ----

type DictEntryService struct {
	repo     port.DictEntryRepository
	typeRepo port.DictTypeRepository
}

func NewDictEntryService(repo port.DictEntryRepository, typeRepo port.DictTypeRepository) *DictEntryService {
	return &DictEntryService{repo: repo, typeRepo: typeRepo}
}

func (s *DictEntryService) ListByTypeID(ctx context.Context, typeID uint, q *query.Query) ([]domain.DictEntry, int64, error) {
	return s.repo.ListByTypeID(ctx, typeID, q)
}

// ListPublicByCode 获取公开字典类型的条目（无需认证）
func (s *DictEntryService) ListPublicByCode(ctx context.Context, code string) ([]domain.DictEntry, error) {
	t, err := s.typeRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, &Error{Code: errcode.ErrDictTypeNotFound}
	}
	if !t.IsPublic {
		return nil, &Error{Code: errcode.ErrDictTypeNotPublic}
	}
	return s.repo.ListByTypeCode(ctx, code)
}

func (s *DictEntryService) GetByID(ctx context.Context, id uint) (*domain.DictEntry, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DictEntryService) Create(ctx context.Context, typeID uint, label, value string, sort, status int, remark, version string) (*domain.DictEntry, error) {
	return s.repo.Create(ctx, typeID, label, value, sort, status, remark, version)
}

func (s *DictEntryService) Update(ctx context.Context, id uint, typeID uint, label, value string, sort, status int, remark, version string) error {
	return s.repo.Update(ctx, id, typeID, label, value, sort, status, remark, version)
}

func (s *DictEntryService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *DictEntryService) BatchDelete(ctx context.Context, ids []uint) error {
	return s.repo.DeleteBatch(ctx, ids)
}

func (s *DictEntryService) BatchUpdateStatus(ctx context.Context, ids []uint, status int) error {
	return s.repo.UpdateStatusBatch(ctx, ids, status)
}
