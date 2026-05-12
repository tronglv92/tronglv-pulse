package db

import (
	"context"
	"time"

	"pulse/helper/utils/cache"
	"gorm.io/gorm"
)

type Adapter interface {
	Connect(ctx context.Context) error
	Close() error
	GetDB() *gorm.DB
	GetCache() cache.Cache
}

type Database interface {
	GetCache() cache.Cache
	GetDB() *gorm.DB
}

type BaseConfig interface {
	GetDriver() string
}

type RDBMSConfig interface {
	BaseConfig
	GetHost() string
	GetPort() int
	GetDBName() string
	GetUsername() string
	GetPassword() string
	GetSchemaName() string
	GetTimeZone() string
	GetMaxIdleConnections() int
	GetMaxOpenConnections() int
	GetConnectTimeout() time.Duration
	GetConnMaxLifetime() time.Duration
	GetConnMaxIdleTime() time.Duration
	GetReplicas() []string
	GetSSLMode() string // Added for security
	GetLogLevel() string
	GetLogSlowThreshold() int
	GetLogIgnoreNotFound() bool
}

type RedisConfig interface {
	BaseConfig
	GetAddr() string // e.g., "localhost:6379"
	GetPassword() string
	GetDB() int
	GetPoolSize() int
	GetPoolTimeout() time.Duration
	GetTLS() bool // Added for security
}
