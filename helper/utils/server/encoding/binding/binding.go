package binding

import (
	"bytes"
	"pulse/helper/utils/errors"
	"pulse/helper/utils/server/encoding"
	_ "pulse/helper/utils/server/encoding/excel"
	"pulse/helper/utils/server/encoding/form"
	_ "pulse/helper/utils/server/encoding/json"
	_ "pulse/helper/utils/server/encoding/proto"
	"fmt"
	"github.com/zeromicro/go-zero/core/mapping"
	"github.com/zeromicro/go-zero/core/validation"
	"github.com/zeromicro/go-zero/rest/httpx"
	"io"
	"net/http"
	"reflect"
	"strings"
)

const (
	baseContentType = "application"
)

func Parse(r *http.Request, v any) error {
	if err := ParseJsonBody(r, v); err != nil {
		return err
	}

	kind := mapping.Deref(reflect.TypeOf(v)).Kind()
	if kind != reflect.Array && kind != reflect.Slice {
		if err := ParseForm(r, v); err != nil {
			return err
		}

		if err := httpx.ParsePath(r, v); err != nil {
			return err
		}

		if err := httpx.ParseHeaders(r, v); err != nil {
			return err
		}
	}

	if valid, ok := v.(validation.Validator); ok {
		return valid.Validate()
	}
	return nil
}

func ParseForm(req *http.Request, target any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	if err := encoding.GetCodec(form.Name).Unmarshal([]byte(req.Form.Encode()), target); err != nil {
		return errors.BadRequest(err)
	}
	return nil
}

func ParseJsonBody(r *http.Request, v any) error {
	codec, ok := CodecForRequest(r, "Content-Type")
	if !ok {
		return nil
	}
	data, err := io.ReadAll(r.Body)

	r.Body = io.NopCloser(bytes.NewBuffer(data))

	if err != nil {
		return errors.BadRequest(err)
	}
	if len(data) == 0 {
		return nil
	}
	if err = codec.Unmarshal(data, v); err != nil {
		return errors.NewBadRequest("INVALID_BODY", fmt.Sprintf("Failed to unmarshal request body: %s", err.Error()))
	}
	return nil
}

func CodecForRequest(r *http.Request, name string) (encoding.Codec, bool) {
	for _, accept := range r.Header[name] {
		codec := encoding.GetCodec(ContentSubtype(accept))
		if codec != nil {
			return codec, true
		}
	}
	return encoding.GetCodec("json"), false
}

func ContentSubtype(contentType string) string {
	left := strings.Index(contentType, "/")
	if left == -1 {
		return ""
	}
	right := strings.Index(contentType, ";")
	if right == -1 {
		right = len(contentType)
	}
	if right < left {
		return ""
	}
	return contentType[left+1 : right]
}

func ContentType(subtype string) string {
	return baseContentType + "/" + subtype
}
