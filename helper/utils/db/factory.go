package db

import (
	"context"
	"fmt"

	"pulse/helper/utils/cache"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	gormtracing "gorm.io/plugin/opentelemetry/tracing"
)

const (
	MysqlDBDriver    string = "mysql"
	PostgresDBDriver string = "postgres"
	SqliteDBDriver   string = "sqlite"
	MongoDBDriver    string = "mongodb"
	RedisDriver      string = "redis"
)

type (
	Option struct {
		GormMigrator     func(db *gorm.DB) error
		Cache            cache.Cache
		Resolver         *dbresolver.DBResolver
		TraceProvider    trace.TracerProvider
		TraceGormOptions []gormtracing.Option
		PrepareStmt      bool
	}
	Opt func(s *Option)
)

type factory struct {
	adapter     Adapter
	dbClient    *gorm.DB
	cacheClient cache.Cache
}

func Must(c BaseConfig, opts ...Opt) Database {
	conn, err := NewFactory(c, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}

func NewFactory(c BaseConfig, opts ...Opt) (Database, error) {
	adapterFactories := map[string]func(BaseConfig, ...Opt) (Adapter, error){
		PostgresDBDriver: MapGormAdapter,
		MysqlDBDriver:    MapGormAdapter,
		SqliteDBDriver:   MapGormAdapter,
		RedisDriver:      MapRedisAdapter,
	}
	factoryFunc, exists := adapterFactories[c.GetDriver()]
	if !exists {
		return nil, fmt.Errorf("unsupported database driver: %s", c.GetDriver())
	}

	adapter, err := factoryFunc(c, opts...)
	if err != nil {
		return nil, err
	}
	if err = adapter.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect adapter for %s: %w", c.GetDriver(), err)
	}
	return &factory{
		adapter: adapter,
	}, nil
}

func (f *factory) GetCache() cache.Cache {
	return f.adapter.GetCache()
}

func (f *factory) GetDB() *gorm.DB {
	return f.adapter.GetDB()
}
