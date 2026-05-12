package localize

import (
	"context"
	"github.com/zeromicro/go-zero/rest"
	"net/http"
	"strings"
)

func middleware(r *http.Request, localizer Localizer) context.Context {
	lang := Fallback
	query := r.URL.Query()
	if val, ok := query["lang"]; ok {
		lang = strings.Join(val, "")
	}
	return WithLangContext(WithContext(r.Context(), localizer.SetLocale(lang)), lang)
}

func RestMiddleware(localizer Localizer) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r.WithContext(middleware(r, localizer)))
		}
	}
}

func MuxMiddleware(localizer Localizer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(middleware(r, localizer)))
		})
	}
}
