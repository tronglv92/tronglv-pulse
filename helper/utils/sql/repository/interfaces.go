package repository

import (
	"context"
	"pulse/helper/utils/model"
	"gorm.io/gorm"
	"time"
)

type Repository[T any] interface {
	CacheAccessor

	GetTableName() string
	GetDB(ctx context.Context, opts ...QueryOpt) *gorm.DB

	Transaction(fc func(tx *gorm.DB) error) error
	BeginTx(ctx context.Context) (*gorm.DB, context.Context)
	Paginate(value interface{}, paginator model.Paginator, db *gorm.DB) func(db *gorm.DB) *gorm.DB

	First(ctx context.Context, opts ...QueryOpt) (*T, error)
	Find(ctx context.Context, opts ...QueryOpt) ([]*T, error)
	FindById(ctx context.Context, id int32, preloads ...string) (*T, error)
	FindByUid(ctx context.Context, uid string, preloads ...string) (*T, error)

	QueryWithCache(ctx context.Context, key string, expiration time.Duration, refresh bool, opts ...QueryOpt) ([]*T, error)

	QueryWithPagination(ctx context.Context, limit int, page int, opts ...QueryOpt) ([]*T, model.Pagination, error)
	QueryWithCursor(ctx context.Context, limit int, prev, next, sortBy, sortOrder string, opts ...QueryOpt) ([]*T, *model.Cursor, error)

	Count(ctx context.Context, opts ...QueryOpt) (int64, error)
	CountWithCache(ctx context.Context, key string, expiration time.Duration, refresh bool, opts ...QueryOpt) (int64, error)

	Create(ctx context.Context, entity *T) error
	BulkCreate(ctx context.Context, entities []*T) error
	CreateWithReturn(ctx context.Context, entity *T) (*T, error)

	Update(ctx context.Context, params any, opts ...QueryOpt) error
	UpdateById(ctx context.Context, params any, id int32) error
	UpdateWithReturn(ctx context.Context, params any, opts ...QueryOpt) (*T, error)

	Delete(ctx context.Context, opts ...QueryOpt) error
	DeleteById(ctx context.Context, id int32) error
	HardDelete(ctx context.Context, opts ...QueryOpt) error
	HardDeleteById(ctx context.Context, id int32) error

	FirstOrCreate(ctx context.Context, entity *T, opts ...QueryOpt) (*T, error)
	UpdateOrCreate(ctx context.Context, create *T, update any, opts ...QueryOpt) error
	Upsert(ctx context.Context, entity *T, opts ...QueryOpt) error

	WithPreload(relation string, opts ...PreloadOpt) QueryOpt
	WithPreloads(relations ...string) QueryOpt
	WithOrder(sortBy string, sortOrder string, fields ...string) QueryOpt
	WithSelect(fields ...string) QueryOpt
	WithDebug() QueryOpt

	SortAble() map[string]string
}
