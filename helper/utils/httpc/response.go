package httpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"net/http"
	"strings"

	"pulse/helper/utils/errors"
	"github.com/zeromicro/go-zero/core/mapping"
	"github.com/zeromicro/go-zero/rest/httpc"
)

const (
	ContentType              = "Content-Type"
	ApplicationJson          = "application/json"
	JsonContentType          = "application/json; charset=utf-8"
	FormUrlEncodeContentType = "application/x-www-form-urlencoded"
	RequestFailed            = "API request was not successful"
)

type HttpError struct {
	StatusCode int
	Url        string
	Body       string
}

func (r HttpError) Error() string {
	return fmt.Sprintf("restp response - %d - %s - %s", r.StatusCode, r.Url, r.Body)
}

func Parse(resp *http.Response, val any) error {
	return httpc.Parse(resp, val)
}

func ParseHeaders(resp *http.Response, val any) error {
	return httpc.ParseHeaders(resp, val)
}

func ParseJsonBody(resp *http.Response, val any) error {
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode >= http.StatusBadRequest {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, resp.Body); err != nil {
			return err
		}

		logx.Errorf("REQUEST_FAILED: %s, %s", resp.Request.URL.Path, string(buf.Bytes()))
		return errors.Newf(
			resp.StatusCode,
			"REQUEST_FAILED",
			"An error occurred while processing your request.",
		)
	}
	if isContentTypeJson(resp) {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, resp.Body); err != nil {
			return err
		}
		if err := json.Unmarshal(buf.Bytes(), val); err != nil {
			return err
		}
		return nil
	}
	return mapping.UnmarshalJsonMap(nil, val)
}

func isContentTypeJson(r *http.Response) bool {
	return strings.Contains(r.Header.Get(ContentType), ApplicationJson)
}
