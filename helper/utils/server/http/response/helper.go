package response

import (
	"bytes"
	"context"
	"pulse/helper/utils/server/encoding/binding"
	"pulse/helper/utils/toolkit/reflectx"
	"fmt"
	"github.com/zeromicro/go-zero/rest/httpx"
	"mime"
	"net/http"
	"path/filepath"
	"reflect"
)

func factory(ctx context.Context, data any, err error) Response {
	return New(ctx, data, err)
}

func SetResponseHandler(handler func(context.Context, Handler) any) {
	responseLock.Lock()
	defer responseLock.Unlock()
	responseHandler = handler
}

func SetIncidentHandler(handler func(ctx context.Context, err error)) {
	responseLock.Lock()
	defer responseLock.Unlock()
	incidentHandler = handler
}

func OkJson(ctx context.Context, w http.ResponseWriter, data any) {
	httpx.OkJsonCtx(ctx, w, factory(ctx, data, nil).GetBody())
	return
}

func OkJsonWithPagination(ctx context.Context, w http.ResponseWriter, items any, paging any) {
	if isSlice := reflect.ValueOf(items).Kind() == reflect.Slice; isSlice {
		if reflectx.IsZeroOfUnderlyingType(items) {
			items = []emptyStruct{}
		}
	}
	httpx.OkJsonCtx(ctx, w, factory(ctx, data{Records: items, Pagination: paging}, nil).GetBody())
	return
}

func OkMsg(ctx context.Context, w http.ResponseWriter) {
	httpx.OkJsonCtx(ctx, w, factory(ctx, nil, nil).GetBody())
	return
}

func Error(ctx context.Context, w http.ResponseWriter, err error) {
	resp := factory(ctx, nil, err)
	httpx.WriteJsonCtx(ctx, w, resp.GetStatus(), resp.GetBody())
	return
}

func ErrorSlice(ctx context.Context, w http.ResponseWriter, err error) {
	resp := factory(ctx, []emptyStruct{}, err)
	httpx.WriteJsonCtx(ctx, w, resp.GetStatus(), resp.GetBody())
	return
}

func FileByBuffer(ctx context.Context, w http.ResponseWriter, filePath string, buf *bytes.Buffer) {
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if len(contentType) == 0 {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
	return
}

func OkJsonLegacy(ctx context.Context, w http.ResponseWriter, items any, paging any) {
	var r = data{Pagination: paging}
	if isSlice := reflect.ValueOf(items).Kind() == reflect.Slice; isSlice {
		r.Records = items
		if reflectx.IsZeroOfUnderlyingType(items) {
			r.Records = []emptyStruct{}
		}
	} else {
		r.Record = items
	}
	OkJson(ctx, w, r)
	return
}

func Write(ctx context.Context, w http.ResponseWriter, r *http.Request, v any) error {
	codec, _ := binding.CodecForRequest(r, "Accept")
	if codec.Name() == "json" {
		OkJson(ctx, w, v)
		return nil
	}

	m, err := codec.Marshal(v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", binding.ContentType(codec.Name()))
	_, err = w.Write(m)
	if err != nil {
		return err
	}
	return nil
}
