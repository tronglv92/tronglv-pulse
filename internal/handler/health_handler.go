package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

// NewHealthServer returns a go-zero REST server that serves only GET /health.
// No auth, no DB — safe to call before any dependency is wired.
func NewHealthServer(name, host string, port int) *rest.Server {
	conf := rest.RestConf{
		ServiceConf: service.ServiceConf{Name: name},
		Host:        host,
		Port:        port,
	}
	srv := rest.MustNewServer(conf)
	srv.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/health",
			Handler: healthHandler(),
		},
	})
	return srv
}

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
