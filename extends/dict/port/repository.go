package port

import (
	"context"

	"kingfisher/core/query"
	"kingfisher/extends/dict/domain"
)

// DictTypeRepository 字典类型仓库接口
type DictTypeRepository interface {
	List(ctx context.Context, q *query.Query) ([]domain.DictType, int64, error)
	GetByID(ctx context.Context, id uint) (*domain.DictType, error)
	GetByCode(ctx context.Context, code string) (*domain.DictType, error)
	Create(ctx context.Context, code, name string, isPublic bool, status int, remark string) (*domain.DictType, error)
	Update(ctx context.Context, id uint, code, name string, isPublic bool, status int, remark string) error
	Delete(ctx context.Context, id uint) error
}

// DictEntryRepository 字典条目仓库接口
type DictEntryRepository interface {
	ListByTypeID(ctx context.Context, typeID uint, q *query.Query) ([]domain.DictEntry, int64, error)
	ListByTypeCode(ctx context.Context, code string) ([]domain.DictEntry, error)
	GetByID(ctx context.Context, id uint) (*domain.DictEntry, error)
	Create(ctx context.Context, typeID uint, label, value string, sort, status int, remark string) (*domain.DictEntry, error)
	Update(ctx context.Context, id uint, typeID uint, label, value string, sort, status int, remark string) error
	Delete(ctx context.Context, id uint) error
	DeleteByTypeID(ctx context.Context, typeID uint) error
}
