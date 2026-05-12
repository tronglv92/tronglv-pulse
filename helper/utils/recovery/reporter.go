package recovery

import (
	"context"
	"pulse/helper/utils/client/incident"
)

type reporter struct {
	serviceName    string
	incidentClient incident.Client
}

func NewReporter(serviceName string, client incident.Client) PanicReporter {
	return &reporter{
		serviceName:    serviceName,
		incidentClient: client,
	}
}

func (r *reporter) GetServiceName() string {
	return r.serviceName
}

func (r *reporter) NotifyError(ctx context.Context, err error) error {
	return r.incidentClient.NotifyError(ctx, err)
}
