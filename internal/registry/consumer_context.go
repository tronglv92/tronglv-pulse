package registry

import (
	"pulse/internal/config"
	"pulse/internal/contract"
)

// ConsumerContext is the DI context for cmd/worker (Kafka consumers + outbox drainer).
type ConsumerContext interface {
	BaseContext
	GetConfig() config.WorkerConfig
	RepositoryContext
	GetCache() contract.Cache
}

type consumerContext struct {
	*baseContext
	RepositoryContext
	config config.WorkerConfig
	cache  contract.Cache
}

// NewConsumerContext opens the DB so consumers can read/write repos and the outbox.
func NewConsumerContext(c config.WorkerConfig) ConsumerContext {
	return &consumerContext{
		baseContext:       newBaseContext(),
		config:            c,
		RepositoryContext: NewRepositoryContext(mustOpenDB(c.DB.DataSource)),
		cache:             newRedisCache(c.Cache),
	}
}

func (c *consumerContext) GetConfig() config.WorkerConfig { return c.config }
func (c *consumerContext) GetCache() contract.Cache       { return c.cache }
