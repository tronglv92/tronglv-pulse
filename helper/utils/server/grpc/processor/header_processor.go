package processor

import (
	"fmt"
	"github.com/zeromicro/go-zero/gateway"
	"net/http"
)

func WithHeaderProcessor(keys ...string) func(*gateway.Server) {
	return gateway.WithHeaderProcessor(func(header http.Header) []string {
		keys = append(keys, "Authorization")
		keys = append(keys, "IAuthorization")
		var values []string
		for _, v := range keys {
			values = append(values, fmt.Sprintf("%s:%s", v, header.Get(v)))
		}
		return values
	})
}
