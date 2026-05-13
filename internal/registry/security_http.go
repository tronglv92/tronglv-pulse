package registry

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

// HttpSecurityContext provides REST middleware for the API server.
type HttpSecurityContext interface {
	// GetAuthMiddleware returns JWT authentication middleware.
	// Stub: passthrough until the auth task wires real JWT verification.
	GetAuthMiddleware() rest.Middleware
	// GetTenantMiddleware injects the authenticated tenant into context.
	// Stub: passthrough until the auth task wires real tenant extraction.
	GetTenantMiddleware() rest.Middleware
}

type httpSecurityContext struct {
	authMiddleware   rest.Middleware
	tenantMiddleware rest.Middleware
}

func NewHttpSecurityContext() HttpSecurityContext {
	passthrough := func(next http.HandlerFunc) http.HandlerFunc { return next }
	return &httpSecurityContext{
		authMiddleware:   passthrough,
		tenantMiddleware: passthrough,
	}
}

func (h *httpSecurityContext) GetAuthMiddleware() rest.Middleware   { return h.authMiddleware }
func (h *httpSecurityContext) GetTenantMiddleware() rest.Middleware { return h.tenantMiddleware }
