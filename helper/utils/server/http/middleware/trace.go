package middleware

import (
	"pulse/helper/utils/toolkit/contextx"
	"github.com/zeromicro/go-zero/core/utils"
	"github.com/zeromicro/go-zero/rest"
	"net/http"
)

func TraceMiddleware(env, name string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = contextx.WithValue(ctx, contextx.KeyElapsedTimer, utils.NewElapsedTimer())
			ctx = contextx.WithValue(ctx, contextx.KeyServiceEnv, env)
			ctx = contextx.WithValue(ctx, contextx.KeyServiceName, name)

			next(w, r.WithContext(ctx))
		}
	}
}
