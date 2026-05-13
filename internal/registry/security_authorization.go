package registry

// AuthorizationContext is a placeholder — Pulse has no external permission service.
// Extend this interface when RBAC is introduced.
type AuthorizationContext interface{}

type authorizationContext struct{}

func NewAuthorizationContext() AuthorizationContext {
	return &authorizationContext{}
}
