package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterSwaggerHandler(svr *rest.Server) {
	h := NewSwaggerHandler()

	svr.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/swagger",
				Handler: h.SwaggerIndex(),
			},
			{
				Method:  http.MethodGet,
				Path:    "/swagger/docs/:name",
				Handler: h.SwaggerDocs(),
			},
		},
	)
}
