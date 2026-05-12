package repository

import (
	"context"
	"pulse/helper/utils/toolkit/stringx"
	"fmt"
	"strings"
	"time"

	"pulse/helper/utils/cache"
	"pulse/helper/utils/db"
	gormcst "pulse/helper/utils/db/gorm"
	"pulse/helper/utils/errors"
	"pulse/helper/utils/model"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	DBContextKey = "DB"
)

type (
	QueryOpt                 func(*gorm.DB) *gorm.DB
	PreloadOpt               any
	defaultRepository[T any] struct {
		db    *gorm.DB
		cache cache.Cache
		*RepoOption
	}
	RepoOption struct {
		tableName string
		sortBy    string
		sortOrder string
		batchSize int
	}
	Option func(s *RepoOption)

	CacheAccessor interface {
		GetCache() cache.Cache
	}
)

func WithTableName(name string) Option {
	return func(m *RepoOption) {
		m.tableName = name
	}
}

func WithBatchSize(size int) Option {
	return func(cfg *RepoOption) {
		cfg.batchSize = size
	}
}

func NewRepository[T any](db db.Database, opts ...Option) Repository[T] {
	repoOpt := &RepoOption{
		batchSize: 100,
	}
	for _, opt := range opts {
		opt(repoOpt)
	}
	return &defaultRepository[T]{
		db:         db.GetDB(),
		cache:      db.GetCache(),
		RepoOption: repoOpt,
	}
}

func (r *defaultRepository[T]) GetCache() cache.Cache {
	return r.cache
}

func (r *defaultRepository[T]) GetTableName() string {
	return r.tableName
}

func (r *defaultRepository[T]) GetDB(ctx context.Context, opts ...QueryOpt) *gorm.DB {
	l := r.db.WithContext(ctx)
	if tx, ok := ctx.Value(DBContextKey).(*gorm.DB); ok {
		l = tx.WithContext(ctx)
	}
	for _, opt := range opts {
		l = opt(l)
	}
	return l
}

func (r *defaultRepository[T]) queryMapping(q string) string {
	if len(r.GetTableName()) > 0 {
		return fmt.Sprintf("%s.%s", r.GetTableName(), q)
	}
	return q
}

func (r *defaultRepository[T]) WithContext(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, DBContextKey, db)
}

func (r *defaultRepository[T]) First(ctx context.Context, opts ...QueryOpt) (*T, error) {
	var result T
	if err := r.GetDB(ctx, opts...).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *defaultRepository[T]) Find(ctx context.Context, opts ...QueryOpt) ([]*T, error) {
	var result []*T
	if err := r.GetDB(ctx, opts...).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r *defaultRepository[T]) FindById(ctx context.Context, id int32, preloads ...string) (*T, error) {
	var result T
	var opts []QueryOpt
	if len(preloads) > 0 {
		opts = append(opts, r.WithPreloads(preloads...))
	}
	if err := r.GetDB(ctx, opts...).Where("id = ?", id).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *defaultRepository[T]) FindByUid(ctx context.Context, uid string, preloads ...string) (*T, error) {
	var result T
	var opts []QueryOpt
	if len(preloads) > 0 {
		opts = append(opts, r.WithPreloads(preloads...))
	}
	if err := r.GetDB(ctx, opts...).Where("uid = ?", uid).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *defaultRepository[T]) QueryWithCache(ctx context.Context, key string, expiration time.Duration, refresh bool, opts ...QueryOpt) ([]*T, error) {
	if r.cache == nil {
		return nil, errors.NewInternalServer("MISSING_CACHE", "Cache instance is not initialized")
	}
	var results []*T
	if !refresh {
		if err := r.cache.Get(key, &results); err == nil {
			return results, nil
		}
	}

	records, err := r.Find(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return records, nil
	}
	if err = r.cache.SetWithExpire(key, records, expiration); err != nil {
		logx.WithContext(ctx).Error(err)
	}
	return records, nil
}

func (r *defaultRepository[T]) QueryWithPagination(ctx context.Context, limit int, page int, opts ...QueryOpt) ([]*T, model.Pagination, error) {
	var results []*T
	paginator := model.NewPaginator(page, limit)
	query := r.GetDB(ctx, opts...)
	err := query.Scopes(r.Paginate(results, paginator, query)).
		Find(&results).
		Error
	if err != nil {
		return nil, nil, err
	}

	return results, paginator.ToPagination(), nil
}

func (r *defaultRepository[T]) QueryWithCursor(ctx context.Context, limit int, prev, next, sortBy, sortOrder string, opts ...QueryOpt) ([]*T, *model.Cursor, error) {
	var results []*T
	var keys = []string{"Id"}
	if len(sortBy) > 0 {
		keys = strings.Split(sortBy, ",")
	}

	p := gormcst.NewCursorPaginator(limit, prev, next, sortOrder, keys)
	_, c, err := p.Paginate(r.GetDB(ctx, opts...), &results)
	if err != nil {
		return nil, nil, err
	}

	respCursor := model.Cursor{Limit: limit}
	if c.After != nil {
		respCursor.Next = *c.After
	}
	if c.Before != nil {
		respCursor.Prev = *c.Before
	}
	return results, &respCursor, err
}

func (r *defaultRepository[T]) Count(ctx context.Context, opts ...QueryOpt) (int64, error) {
	var m T
	var total int64
	if err := r.GetDB(ctx, opts...).Model(&m).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *defaultRepository[T]) CountWithCache(ctx context.Context, key string, expiration time.Duration, refresh bool, opts ...QueryOpt) (int64, error) {
	if r.cache == nil {
		return 0, errors.NewInternalServer("MISSING_CACHE", "Cache instance is not initialized")
	}
	var total int64
	if !refresh {
		if err := r.cache.Get(key, &total); err == nil {
			return total, nil
		}
	}

	count, err := r.Count(ctx, opts...)
	if err != nil {
		return 0, err
	}
	if err = r.cache.SetWithExpire(key, count, expiration); err != nil {
		logx.WithContext(ctx).Error(err)
	}
	return count, nil
}

func (r *defaultRepository[T]) Create(ctx context.Context, model *T) error {
	if err := r.GetDB(ctx).Create(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *defaultRepository[T]) BulkCreate(ctx context.Context, models []*T) error {
	if err := r.GetDB(ctx).CreateInBatches(models, r.batchSize).Error; err != nil {
		return err
	}
	return nil
}

func (r *defaultRepository[T]) CreateWithReturn(ctx context.Context, model *T) (*T, error) {
	if err := r.GetDB(ctx).Create(model).Error; err != nil {
		return nil, err
	}
	return model, nil
}

func (r *defaultRepository[T]) Update(ctx context.Context, params any, opts ...QueryOpt) error {
	var result T
	if err := r.GetDB(ctx, opts...).Model(&result).Updates(params).Error; err != nil {
		return err
	}
	return nil
}

func (r *defaultRepository[T]) UpdateById(ctx context.Context, params any, id int32) error {
	return r.Update(ctx, params, func(g *gorm.DB) *gorm.DB {
		return g.Where("id=?", id)
	})
}

func (r *defaultRepository[T]) UpdateWithReturn(ctx context.Context, params any, opts ...QueryOpt) (*T, error) {
	var result T
	if err := r.GetDB(ctx, opts...).Clauses(clause.Returning{}).Model(&result).Updates(params).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *defaultRepository[T]) Delete(ctx context.Context, opts ...QueryOpt) error {
	var result T
	return r.GetDB(ctx, opts...).Delete(&result).Error
}

func (r *defaultRepository[T]) DeleteById(ctx context.Context, id int32) error {
	return r.Delete(ctx, func(g *gorm.DB) *gorm.DB {
		return g.Where("id=?", id)
	})
}

func (r *defaultRepository[T]) HardDelete(ctx context.Context, opts ...QueryOpt) error {
	var result T
	return r.GetDB(ctx, opts...).Unscoped().Delete(&result).Error
}

func (r *defaultRepository[T]) HardDeleteById(ctx context.Context, id int32) error {
	return r.HardDelete(ctx, func(g *gorm.DB) *gorm.DB {
		return g.Where("id=?", id)
	})
}

func (r *defaultRepository[T]) Transaction(fc func(tx *gorm.DB) error) error {
	return r.db.Transaction(fc)
}

func (r *defaultRepository[T]) BeginTx(ctx context.Context) (*gorm.DB, context.Context) {
	tx := r.db.Begin()
	return tx, r.WithContext(ctx, tx)
}

func (r *defaultRepository[T]) Paginate(value interface{}, paginator model.Paginator, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRecords int64
	db.Model(value).Count(&totalRecords)

	paginator.WithTotalCount(totalRecords)
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(paginator.Offset()).Limit(paginator.PageSize())
	}
}

func (r *defaultRepository[T]) FirstOrCreate(ctx context.Context, entity *T, opts ...QueryOpt) (*T, error) {
	result, err := r.First(ctx, opts...)
	if err != nil {
		if !errors.IsRecordNotFound(err) {
			return nil, err
		}
		return r.CreateWithReturn(ctx, entity)
	}
	return result, nil
}

func (r *defaultRepository[T]) UpdateOrCreate(ctx context.Context, create *T, update any, opts ...QueryOpt) error {
	_, err := r.First(ctx, opts...)
	if err != nil {
		if errors.IsRecordNotFound(err) {
			return r.Create(ctx, create)
		}
		return err
	}
	return r.Update(ctx, update, opts...)
}

func (r *defaultRepository[T]) Upsert(ctx context.Context, entity *T, opts ...QueryOpt) error {
	_, err := r.First(ctx, opts...)
	if err != nil {
		if errors.IsRecordNotFound(err) {
			return r.Create(ctx, entity)
		}
		return err
	}
	return r.Update(ctx, entity, opts...)
}

func (r *defaultRepository[T]) WithPreload(relation string, opts ...PreloadOpt) QueryOpt {
	return func(g *gorm.DB) *gorm.DB {
		var preloadOpts []interface{}
		for _, opt := range opts {
			preloadOpts = append(preloadOpts, opt)
		}
		return g.Preload(relation, preloadOpts...)
	}
}

func (r *defaultRepository[T]) WithPreloads(relations ...string) QueryOpt {
	return func(g *gorm.DB) *gorm.DB {
		if len(relations) == 0 {
			return g
		}
		for _, relation := range relations {
			g = g.Preload(relation)
		}
		return g
	}
}

func (r *defaultRepository[T]) WithOrder(sortBy string, sortOrder string, fields ...string) QueryOpt {
	return func(db *gorm.DB) *gorm.DB {
		if len(sortBy) == 0 {
			sortBy = r.sortBy
		}
		if len(sortOrder) == 0 {
			sortOrder = r.sortOrder
		}
		m := r.SortAble()
		if len(fields) > 0 {
			for _, v := range fields {
				m[v] = v
			}
		}
		if val, ok := m[sortBy]; ok {
			return db.Order(fmt.Sprintf("%s %s", val, stringx.SortOrder(sortOrder)))
		}
		return db
	}
}

func (r *defaultRepository[T]) SortAble() map[string]string {
	return map[string]string{
		"id":         r.queryMapping("id"),
		"created_at": r.queryMapping("created_at"),
	}
}

func (r *defaultRepository[T]) WithSelect(fields ...string) QueryOpt {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select(fields)
	}
}

func (r *defaultRepository[T]) WithDebug() QueryOpt {
	return func(db *gorm.DB) *gorm.DB {
		return db.Debug()
	}
}
