package response

import (
	"context"
	"pulse/helper/utils/errors"
	"pulse/helper/utils/server/encoding"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"strconv"
	"sync"
)

var (
	incidentHandler func(ctx context.Context, err error)
	responseHandler func(context.Context, Handler) any
	responseLock    sync.RWMutex
)

type (
	Response interface {
		GetBody() any
		GetStatus() int
	}

	Handler interface {
		GetTraceId() string
		GetMessage() string
		GetCode() string
		GetResult() []byte
		GetMetadata() map[string]string
	}
)

type responseSvc struct {
	ctx    context.Context
	err    *errors.Error
	codec  encoding.Codec
	result any
}

func New(ctx context.Context, result any, err error) Response {
	return &responseSvc{
		ctx:    ctx,
		result: result,
		codec:  encoding.GetCodec("json"),
		err:    errors.FromNotNil(err),
	}
}

func (r *responseSvc) GetBody() any {
	responseLock.RLock()
	handler := responseHandler
	responseLock.RUnlock()

	if handler == nil {
		handler = MetaDataResponse
	}
	return handler(r.ctx, r)
}

func (r *responseSvc) GetResult() []byte {
	if r.result == nil {
		return []byte("{}")
	}

	result, err := r.codec.Marshal(r.result)
	if err != nil {
		return []byte("{}")
	}
	return result
}

func (r *responseSvc) GetTraceId() string {
	spanCtx := trace.SpanFromContext(r.ctx).SpanContext()
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return "unknown-trace-id"
}

func (r *responseSvc) GetCode() string {
	if r.err != nil {
		return strconv.Itoa(int(r.err.GetCode()))
	}
	return fmt.Sprintf("%d", r.GetStatus())
}

func (r *responseSvc) GetMessage() string {
	if r.err != nil {
		return r.err.GetMessage()
	}
	return "Success"
}

func (r *responseSvc) GetStatus() int {
	if r.err == nil {
		return http.StatusOK
	}
	defer r.tracing()
	return int(r.err.GetCode())
}

func (r *responseSvc) GetMetadata() map[string]string {
	if r.err != nil {
		return r.err.GetMetadata()
	}
	return nil
}

func (r *responseSvc) tracing() {
	span := trace.SpanFromContext(r.ctx)
	defer span.End()
	if span.IsRecording() {
		span.RecordError(r.err)
		if len(r.err.GetMetadata()) > 0 {
			var items []attribute.KeyValue
			for k, v := range r.err.GetMetadata() {
				items = append(items, attribute.String(k, v))
			}
			span.SetAttributes(items...)
		}
	}
	if r.err != nil {
		logx.WithContext(r.ctx).Error(r.err)
	}
	if r.err.GetCode() == http.StatusInternalServerError && incidentHandler != nil {
		incidentHandler(r.ctx, r.err)
	}
}
