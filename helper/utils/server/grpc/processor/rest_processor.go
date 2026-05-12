package processor

import (
	"pulse/helper/utils/server"
	"github.com/zeromicro/go-zero/gateway"
)

func WithRestProcessor(h server.RestHandler) func(*gateway.Server) {
	return func(s *gateway.Server) {
		h.Register(s.Server)
	}
}
