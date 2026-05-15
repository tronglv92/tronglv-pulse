package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"

	"pulse/helper/utils/errors"
	"pulse/helper/utils/server/http/response"
)

// RecoveryMiddleware returns a rest.Middleware that catches panics and returns a
// structured 500 JSON error via response.Error. In non-production environments
// the stack trace is included in the response metadata; in production the stack
// is logged but not exposed.
func RecoveryMiddleware(env string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if result := recover(); result != nil {
					ctx := r.Context()
					stack := string(debug.Stack())

					logx.WithContext(ctx).Errorw("panic recovered",
						logx.Field("panic", result),
						logx.Field("stack", stack),
					)

					var metadata map[string]string
					if env != service.ProMode {
						metadata = map[string]string{"stack": stack}
					}

					response.Error(ctx, w, errors.NewInternalServer(
						"PANIC_RECOVERED",
						"An unexpected server error occurred.",
					).WithMetadata(metadata))
				}
			}()
			next(w, r)
		}
	}
}
