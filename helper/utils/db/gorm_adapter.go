package db

import (
	"context"
	"fmt"
	"time"

	"pulse/helper/utils/cache"
	gormcst "pulse/helper/utils/db/gorm"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	gormtracing "gorm.io/plugin/opentelemetry/tracing"
)

type (
	Dialect interface {
		ConnectString() gorm.Dialector
	}

	dialectSvc struct {
		config RDBMSConfig
	}

	gormAdapter struct {
		config RDBMSConfig
		conn   *gorm.DB
		*Option
	}
)

func WithGormMigrator(fnc func(db *gorm.DB) error) Opt {
	return func(m *Option) {
		m.GormMigrator = fnc
	}
}

func WithCache(client cache.Cache) Opt {
	return func(m *Option) {
		m.Cache = client
	}
}

func MapGormAdapter(c BaseConfig, opts ...Opt) (Adapter, error) {
	pc, ok := c.(RDBMSConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for gorm")
	}
	return NewGormAdapter(pc, opts...)
}

func NewGormAdapter(config RDBMSConfig, opts ...Opt) (Adapter, error) {
	s := &gormAdapter{
		config: config,
		Option: &Option{
			PrepareStmt:   false,
			TraceProvider: otel.GetTracerProvider(),
		},
	}
	for _, opt := range opts {
		opt(s.Option)
	}
	return s, nil
}

func (s *gormAdapter) Connect(_ context.Context) error {
	conn, err := gorm.Open(NewDialect(s.config).ConnectString(), s.getOptions())
	if err != nil {
		return err
	}

	if err = s.setConn(conn); err != nil {
		return err
	}

	if err = s.registerMigration(); err != nil {
		return err
	}

	if err = s.registerConnectionPool(); err != nil {
		logx.Error(err)
	}

	if err = s.registerTracing(); err != nil {
		logx.Error(err)
	}

	return nil
}

func (s *gormAdapter) Close() error {
	return nil
}

func (s *gormAdapter) GetDB() *gorm.DB {
	return s.conn
}

func (s *gormAdapter) GetCache() cache.Cache {
	return s.Cache
}

func (s *gormAdapter) setConn(conn *gorm.DB) error {
	s.conn = conn
	return nil
}

func (s *gormAdapter) getOptions() *gorm.Config {
	var schemaName string
	if len(s.config.GetSchemaName()) > 0 {
		schemaName = fmt.Sprintf("%s.", s.config.GetSchemaName())
	}
	return &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			TablePrefix:   schemaName,
		},
		PrepareStmt: s.PrepareStmt,
		Logger: gormcst.NewSqlLogger(
			gormcst.GetLogLevel(s.config.GetLogLevel()),
			time.Duration(s.config.GetLogSlowThreshold())*time.Millisecond,
			s.config.GetLogIgnoreNotFound(),
		),
	}
}

func (s *gormAdapter) registerConnectionPool() error {
	sqlDB, e := s.conn.DB()
	if e != nil {
		return e
	}
	sqlDB.SetMaxIdleConns(s.config.GetMaxIdleConnections())
	sqlDB.SetMaxOpenConns(s.config.GetMaxOpenConnections())
	sqlDB.SetConnMaxIdleTime(time.Second * s.config.GetConnMaxIdleTime())
	sqlDB.SetConnMaxLifetime(time.Second * s.config.GetConnMaxLifetime())
	return nil
}

func (s *gormAdapter) registerTracing() error {
	opts := append(
		[]gormtracing.Option{
			gormtracing.WithTracerProvider(s.TraceProvider),
			gormtracing.WithoutMetrics(),
			gormtracing.WithoutServerAddress(),
		},
		s.TraceGormOptions...,
	)
	return s.conn.Use(gormtracing.NewPlugin(opts...))
}

func (s *gormAdapter) registerMigration() error {
	if s.GormMigrator == nil {
		return nil
	}
	return s.GormMigrator(s.conn)
}

func NewDialect(config RDBMSConfig) Dialect {
	return &dialectSvc{
		config: config,
	}
}

func (s *dialectSvc) ConnectString() gorm.Dialector {
	var dialect gorm.Dialector
	switch s.config.GetDriver() {
	case PostgresDBDriver:
		dialect = s.postgresOpen()
	case MysqlDBDriver:
		dialect = s.mysqlOpen()
	case SqliteDBDriver:
		dialect = s.sqliteOpen()
	}
	return dialect
}

func (s *dialectSvc) postgresOpen() gorm.Dialector {
	return postgres.Open(fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=%s TimeZone=%s connect_timeout=%d",
		s.config.GetHost(),
		s.config.GetPort(),
		s.config.GetUsername(),
		s.config.GetPassword(),
		s.config.GetDBName(),
		"disable",
		s.config.GetSchemaName(),
		s.config.GetTimeZone(),
		s.config.GetConnectTimeout(),
	))
}

func (s *dialectSvc) mysqlOpen() gorm.Dialector {
	return mysql.Open(fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local&timeout=%ds",
		s.config.GetUsername(),
		s.config.GetPassword(),
		s.config.GetHost(),
		s.config.GetPort(),
		s.config.GetDBName(),
		s.config.GetConnectTimeout(),
	))
}

func (s *dialectSvc) sqliteOpen() gorm.Dialector {
	return sqlite.Open(s.config.GetDBName())
}
