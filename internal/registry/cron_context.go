package registry

import (
	"pulse/helper/utils/toolkit/oncex"
	"pulse/internal/config"
	"pulse/internal/service"
)

// CronContext is the DI context for cmd/cron (scheduled background jobs).
type CronContext interface {
	BaseContext
	GetConfig() config.CronConfig
	RepositoryContext
	GetServiceFactory() *service.ServiceFactory
}

type cronContext struct {
	*baseContext
	RepositoryContext
	config         config.CronConfig
	serviceFactory oncex.OnceValue[*service.ServiceFactory]
}

// NewCronContext opens the DB for cron job reads/writes (anomaly scan, retention, etc.).
func NewCronContext(c config.CronConfig) CronContext {
	return &cronContext{
		baseContext:       newBaseContext(),
		config:            c,
		RepositoryContext: NewRepositoryContext(mustOpenDB(c.DB.DataSource)),
	}
}

func (c *cronContext) GetConfig() config.CronConfig { return c.config }

func (c *cronContext) GetServiceFactory() *service.ServiceFactory {
	return c.serviceFactory.MustGet(func() *service.ServiceFactory {
		return service.NewServiceFactory(c)
	})
}
