package registry

import "pulse/internal/config"

// SecurityContext composes authentication, authorization, and HTTP middleware.
type SecurityContext interface {
	AuthenticationContext
	AuthorizationContext
	HttpSecurityContext
}

type securityContext struct {
	AuthenticationContext
	AuthorizationContext
	HttpSecurityContext
}

// NewSecurityContext wires all three security sub-contexts for the API server.
func NewSecurityContext(c config.APIConfig) SecurityContext {
	return &securityContext{
		AuthenticationContext: NewAuthenticationContext(c),
		AuthorizationContext:  NewAuthorizationContext(),
		HttpSecurityContext:   NewHttpSecurityContext(),
	}
}
