package request

import (
	"pulse/helper/utils/server/encoding/binding"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/protobuf/proto"
	"net/http"
)

type RequestBinder interface {
	Bind(*http.Request) error
}

func Parse(r *http.Request, v any) error {
	switch m := v.(type) {
	case proto.Message:
		return binding.Parse(r, m)
	default:
		return httpx.Parse(r, v)
	}
}
