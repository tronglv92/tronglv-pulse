package recovery

import (
	"context"
	"pulse/helper/utils/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"runtime/debug"
)

// PrintStack prints the current stack trace to stderr.
// Typically used for debugging in development environments.
func PrintStack() {
	debug.PrintStack()
}

// SprintStack returns the current stack trace as a string.
// Useful for logging the stack without printing to stderr directly.
func SprintStack() string {
	return string(debug.Stack())
}

// PanicReporter defines the interface for handling panic recovery
// with external incident reporting and environment awareness.
type PanicReporter interface {
	GetServiceName() string                           // Returns the name of the service for incident reporting
	NotifyError(ctx context.Context, err error) error // Used to send incident reports
}

// Recover handles panic recovery by logging the panic, recording it in traces,
// and printing the stack trace if in a non-production environment.
// Use this in a `defer` statement: defer Recover(ctx)
func Recover(ctx context.Context) {
	handleRecover(ctx, nil)
}

// RecoverWithReporter is an extended version of Recovery that also reports
// the panic to an external incident system via the provided RecoveryStrategy.
// Use this in a `defer` statement: defer PanicReporter(ctx, reporter)
func RecoverWithReporter(ctx context.Context, reporter PanicReporter) {
	handleRecover(ctx, reporter)
}

// handleRecover is the internal shared implementation that recovers from panics,
// logs the error, records tracing spans, and optionally notifies an incident system.
// It should not be used directly; prefer Recovery or PanicReporter.
func handleRecover(ctx context.Context, reporter PanicReporter) {
	if r := recover(); r != nil {
		err := errors.ToError(r)

		span := trace.SpanFromContext(ctx)
		if span != nil {
			defer span.End()
			span.RecordError(err)
		}

		stackStr := SprintStack()
		logx.WithContext(ctx).
			WithFields(
				logx.Field("panic", true),
				logx.Field("stack", stackStr),
			).
			Error(err.Error())

		if reporter != nil {
			err = errors.NewInternalServer("PANIC_RECOVERED", "An unexpected server error occurred.").WithMetadata(map[string]string{
				"stack": stackStr,
			})
			if err = reporter.NotifyError(ctx, err); err != nil {
				logx.WithContext(ctx).Error(err)
			}
			return
		}
		PrintStack()
	}
}
