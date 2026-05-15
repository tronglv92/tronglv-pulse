package registry

import (
	"pulse/internal/auth"
	"pulse/internal/config"

	"github.com/zeromicro/go-zero/rest"
)

// HttpSecurityContext provides REST middleware for the API server.
type HttpSecurityContext interface {
	// GetAuthMiddleware returns JWT authentication middleware.
	GetAuthMiddleware() rest.Middleware
	// GetTenantMiddleware injects the authenticated tenant into context.
	GetTenantMiddleware() rest.Middleware
}

type httpSecurityContext struct {
	authMiddleware   rest.Middleware
	tenantMiddleware rest.Middleware
}

func NewHttpSecurityContext(c config.APIConfig) HttpSecurityContext {
	return &httpSecurityContext{
		authMiddleware:   auth.AuthMiddleware(c.Auth.JwtSecret),
		tenantMiddleware: auth.TenantMiddleware(),
	}
}

func (h *httpSecurityContext) GetAuthMiddleware() rest.Middleware   { return h.authMiddleware }
func (h *httpSecurityContext) GetTenantMiddleware() rest.Middleware { return h.tenantMiddleware }
