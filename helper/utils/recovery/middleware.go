package recovery

import (
	"pulse/helper/utils/errors"
	"pulse/helper/utils/server/http/response"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"net/http"
)

func recovery(w http.ResponseWriter, r *http.Request, env string) {
	if result := recover(); result != nil {
		var metadata = make(map[string]string)
		if env == service.ProMode {
			logx.ErrorStack(result)
		} else {
			PrintStack()
			metadata["stack"] = SprintStack()
		}
		response.Error(r.Context(), w, errors.NewInternalServer("PANIC_RECOVERED", "An unexpected server error occurred.").WithMetadata(metadata))
		return
	}
}

func RestMiddleware(env string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer recovery(w, r, env)
			next(w, r)
		}
	}
}

func MuxMiddleware(env string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer recovery(w, r, env)
			next.ServeHTTP(w, r)
		})
	}
}
