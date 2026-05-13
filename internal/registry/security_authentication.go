package registry

import "pulse/internal/config"

// AuthenticationContext exposes credentials needed by JWT middleware.
type AuthenticationContext interface {
	GetJWTSecret() string
}

type authenticationContext struct {
	jwtSecret string
}

func NewAuthenticationContext(c config.APIConfig) AuthenticationContext {
	return &authenticationContext{jwtSecret: c.Auth.JwtSecret}
}

func (a *authenticationContext) GetJWTSecret() string { return a.jwtSecret }
