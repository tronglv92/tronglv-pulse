package response

import (
	"context"
	"pulse/helper/utils/toolkit/reflectx"
	"encoding/json"
	"reflect"
)

type emptyStruct struct{}

type responseHttp struct {
	Meta metaResponse    `json:"meta"`
	Data json.RawMessage `json:"data"`
}

type metaResponse struct {
	TraceId  string            `json:"trace_id"`
	Code     string            `json:"code"`
	Message  string            `json:"message"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func MetaDataResponse(_ context.Context, handler Handler) any {
	return responseHttp{
		Meta: metaResponse{
			TraceId:  handler.GetTraceId(),
			Code:     handler.GetCode(),
			Message:  handler.GetMessage(),
			Metadata: handler.GetMetadata(),
		},
		Data: handler.GetResult(),
	}
}

func OldDataResponse(_ context.Context, handler Handler) any {
	return responseHttp{
		Meta: metaResponse{
			TraceId:  handler.GetTraceId(),
			Code:     handler.GetCode(),
			Message:  handler.GetMessage(),
			Metadata: handler.GetMetadata(),
		},
		Data: handler.GetResult(),
	}
}

type data struct {
	Pagination any `json:"pagination,omitempty"`
	Records    any `json:"records,omitempty"`
	Record     any `json:"record,omitempty"`
}

func (i *data) SetData(result interface{}) data {
	isNil := reflectx.IsZeroOfUnderlyingType(result)
	if isSlice := reflect.ValueOf(result).Kind() == reflect.Slice; isSlice {
		if isNil {
			result = []emptyStruct{}
		}
		i.Records = result
	} else {
		if isNil {
			result = emptyStruct{}
		}
		i.Record = result
	}
	return *i
}
