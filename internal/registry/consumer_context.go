package registry

import "pulse/internal/config"

// ConsumerContext is the DI context for cmd/worker (Kafka consumers + outbox drainer).
type ConsumerContext interface {
	BaseContext
	GetConfig() config.WorkerConfig
	RepositoryContext
}

type consumerContext struct {
	*baseContext
	RepositoryContext
	config config.WorkerConfig
}

// NewConsumerContext opens the DB so consumers can read/write repos and the outbox.
func NewConsumerContext(c config.WorkerConfig) ConsumerContext {
	return &consumerContext{
		baseContext:       newBaseContext(),
		config:            c,
		RepositoryContext: NewRepositoryContext(mustOpenDB(c.DB.DataSource)),
	}
}

func (c *consumerContext) GetConfig() config.WorkerConfig { return c.config }
