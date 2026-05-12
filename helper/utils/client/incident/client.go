package incident

import (
	"context"
	"pulse/helper/utils/errors"
	"pulse/helper/utils/httpc"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"time"
)

// Client defines an interface for reporting or notifying system incidents and errors.
type Client interface {
	// NotifyError sends a notification or logs the provided error.
	// Intended for general error reporting, without requiring structured incident data.
	NotifyError(ctx context.Context, err error) error

	// NotifyIfCritical checks the incident's level and type,
	// and sends a notification only if it's classified as critical.
	NotifyIfCritical(ctx context.Context, req IncidentDescriptor) error

	// NotifyAlways sends the incident notification unconditionally,
	// regardless of its severity level or type.
	NotifyAlways(ctx context.Context, req IncidentDescriptor) error
}

type (
	Config interface {
		GetName() string
		GetUrl() string
		GetClientId() string
		GetClientSecret() string
		GetHttpClient() httpc.Service
	}

	clientSvc struct {
		httpClient   httpc.Service
		serviceName  string
		gatewayUrl   string
		clientId     string
		clientSecret string
	}

	IncidentDescriptor interface {
		GetService() string
		GetLevel() string
		GetMessage() string
		GetType() string
		GetTimestamp() string
		GetTraceId() string
		GetStackTrace() string
		GetPath() string
		GetMethod() string
		GetStatusCode() int
	}
)

func New(c Config) Client {
	return &clientSvc{
		serviceName:  c.GetName(),
		gatewayUrl:   c.GetUrl(),
		clientId:     c.GetClientId(),
		clientSecret: c.GetClientSecret(),
		httpClient:   c.GetHttpClient(),
	}
}

func (c *clientSvc) NotifyError(ctx context.Context, err error) error {
	e := errors.FromNotNil(err)
	if e == nil {
		return nil
	}
	return c.send(ctx, IncidentPayload{
		Service:    c.serviceName,
		Level:      "error",
		Message:    err.Error(),
		Type:       fmt.Sprintf("%T", err),
		StatusCode: int(e.GetCode()),
		Timestamp:  time.Now().Format(time.RFC3339),
		TraceId:    trace.SpanContextFromContext(ctx).TraceID().String(),
		StackTrace: e.GetMeta("stack"),
	})
}

func (c *clientSvc) NotifyIfCritical(ctx context.Context, req IncidentDescriptor) error {
	return c.send(ctx, req)
}

func (c *clientSvc) NotifyAlways(ctx context.Context, req IncidentDescriptor) error {
	return c.send(ctx, req)
}

func (c *clientSvc) send(ctx context.Context, payload IncidentDescriptor) error {
	_, err := c.httpClient.Post(ctx,
		fmt.Sprintf("%s/central-log-svc/api/v1/collectors", c.gatewayUrl),
		IncidentPayload{
			Service:    payload.GetService(),
			Type:       payload.GetType(),
			Level:      payload.GetLevel(),
			Message:    payload.GetMessage(),
			TraceId:    payload.GetTraceId(),
			StackTrace: payload.GetStackTrace(),
			Path:       payload.GetPath(),
			Method:     payload.GetMethod(),
			StatusCode: payload.GetStatusCode(),
			Timestamp:  payload.GetTimestamp(),
		},
		httpc.WithBasicAuth(c.clientId, c.clientSecret),
	)
	return err
}
