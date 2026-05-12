package security

import (
	"pulse/helper/utils/errors"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func RegisterSecurityHandler(svr *rest.Server, handler SecurityHandler, prefix string) {
	svr.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/security/internal/webhook",
				Handler: handler.Webhook(),
			},
		},
		rest.WithPrefix(prefix),
	)
}

type SecurityHandler interface {
	Webhook() http.HandlerFunc
}

type securityHandler struct {
	perms PermissionProvider
}

func NewSecurityHandler(p PermissionProvider) SecurityHandler {
	return &securityHandler{
		perms: p,
	}
}

func (s *securityHandler) Webhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.perms.Refresh(r.Context()); err != nil {
			httpx.ErrorCtx(r.Context(), w, errors.NewInternalServer("PERMISSION_REFRESH_FAILED", err.Error()))
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]any{
			"success": true,
		})
	}
}
