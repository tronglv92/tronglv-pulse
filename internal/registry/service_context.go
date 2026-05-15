package registry

import (
	"pulse/helper/utils/toolkit/oncex"
	"pulse/internal/config"
	"pulse/internal/service"
	"pulse/internal/types/entity"

	migratelib "github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ServiceContext is the root DI context for cmd/api.
// Composes BaseContext, RepositoryContext, and SecurityContext; vends a lazy ServiceFactory.
type ServiceContext interface {
	BaseContext
	GetConfig() config.APIConfig
	RepositoryContext
	SecurityContext
	GetServiceFactory() *service.ServiceFactory
}

type serviceContext struct {
	*baseContext
	RepositoryContext
	SecurityContext
	config         config.APIConfig
	serviceFactory oncex.OnceValue[*service.ServiceFactory]
}

// NewServiceContext opens the DB, runs AutoMigrate, and wires all sub-contexts.
// Panics if the DB is unavailable — fail fast so misconfigured services don't start silently.
func NewServiceContext(c config.APIConfig) ServiceContext {
	return &serviceContext{
		baseContext:       newBaseContext(),
		config:            c,
		RepositoryContext: NewRepositoryContext(mustOpenDB(c.DB.DataSource)),
		SecurityContext:   NewSecurityContext(c),
	}
}

func (s *serviceContext) GetConfig() config.APIConfig { return s.config }

func (s *serviceContext) GetServiceFactory() *service.ServiceFactory {
	return s.serviceFactory.MustGet(func() *service.ServiceFactory {
		return service.NewServiceFactory(s)
	})
}

// IngestContext is the DI context for cmd/ingest.
// No DB — ingest publishes to Kafka only via the outbox pattern (TASK-032+).
type IngestContext interface {
	BaseContext
	GetConfig() config.IngestConfig
}

type ingestContext struct {
	*baseContext
	config config.IngestConfig
}

func NewIngestContext(c config.IngestConfig) IngestContext {
	return &ingestContext{
		baseContext: newBaseContext(),
		config:      c,
	}
}

func (i *ingestContext) GetConfig() config.IngestConfig { return i.config }

// mustOpenDB opens a GORM/Postgres connection, ensures Postgres extensions and
// custom types exist, then runs AutoMigrate (additive-only). Panics on error.
//
// Full schema setup (hypertables, indexes) requires `make migrate-up` once.
// AutoMigrate handles day-to-day column additions without re-running migrations.
func mustOpenDB(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	logx.Must(err)
	mustMigrate(db)
	// logx.Must(ensurePrerequisites(db))
	logx.Must(db.AutoMigrate(entity.All()...))
	return db
}

// mustMigrate runs all pending SQL migrations from db/migrations/ using golang-migrate.
// It reuses the existing GORM connection so no second DB connection is opened.
// ErrNoChange (already at latest) is treated as success.
func mustMigrate(db *gorm.DB) {
	sqlDB, err := db.DB()
	logx.Must(err)

	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	logx.Must(err)

	m, err := migratelib.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	logx.Must(err)

	if err = m.Up(); err != nil && err != migratelib.ErrNoChange {
		logx.Must(err)
	}
}

// ensurePrerequisites creates Postgres extensions and ENUM types that AutoMigrate
// requires but cannot create itself. All statements are idempotent.
func ensurePrerequisites(db *gorm.DB) error {
	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS vector`,
		`CREATE EXTENSION IF NOT EXISTS pg_trgm`,
		`CREATE EXTENSION IF NOT EXISTS citext`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'incident_status') THEN
				CREATE TYPE incident_status AS ENUM (
					'open', 'investigating', 'resolving', 'resolved', 'closed'
				);
			END IF;
		END $$`,
	}
	for _, sql := range statements {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}
