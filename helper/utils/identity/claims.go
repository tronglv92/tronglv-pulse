package identity

import (
	"pulse/helper/utils/toolkit/slicesx"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

const (
	authClaimsContext = "auth-claims-key"
	UserKind          = "user"
	ClientKind        = "client"
)

type Claims interface {
	GetJti() string           // Unique token ID
	GetId() string            // subject ID
	GetName() string          // subject name
	GetKind() string          // subject kind (e.g., "user", "client")
	GetSource() string        // identity provider or login source
	GetRoles() []string       // roles assigned to the subject
	GetPermissions() []string // permission codes assigned to the subject
	GetExpiry() int64         // token expiry (unix timestamp)
	GetIssuer() string        // JWT issuer
	GetAudience() []string    // JWT audience
	GetAttributes() map[string]string
	IsSingleSession() bool // indicate if this token is single-session mode
}

type MapClaims struct {
	jwt.RegisteredClaims
	Subject     string
	Name        string
	Kind        string
	Source      string
	Roles       []string
	Permissions []string
	Attributes  map[string]string
	SingleMode  bool
}

func (c *MapClaims) GetJti() string  { return c.ID }
func (c *MapClaims) GetId() string   { return c.Subject }
func (c *MapClaims) GetName() string { return c.Name }
func (c *MapClaims) GetKind() string { return c.Kind }
func (c *MapClaims) GetSource() string {
	if c.Source == "" {
		return "service"
	}
	return c.Source
}
func (c *MapClaims) GetRoles() []string {
	if c.Roles == nil {
		return []string{}
	}
	return c.Roles
}
func (c *MapClaims) GetPermissions() []string {
	if c.Permissions == nil {
		return []string{}
	}
	return c.Permissions
}
func (c *MapClaims) GetExpiry() int64 {
	if c.ExpiresAt != nil {
		return c.ExpiresAt.Unix()
	}
	return 0
}
func (c *MapClaims) GetIssuer() string { return c.Issuer }
func (c *MapClaims) GetAudience() []string {
	if c.Audience == nil {
		return []string{}
	}
	return c.Audience
}
func (c *MapClaims) GetAttributes() map[string]string { return c.Attributes }
func (c *MapClaims) SetPermissions(perms []string)    { c.Permissions = perms }
func (c *MapClaims) SetRoles(roles []string)          { c.Roles = roles }
func (c *MapClaims) HasRole(codes ...string) bool {
	return slicesx.HasAny(c.GetRoles(), codes...)
}
func (c *MapClaims) HasPermission(codes ...string) bool {
	return slicesx.HasAny(c.GetPermissions(), codes...)
}
func (c *MapClaims) IsSingleSession() bool { return c.SingleMode }

func MapClaimsFrom(c Claims) *MapClaims {
	return &MapClaims{
		Subject:     c.GetId(),
		Name:        c.GetName(),
		Kind:        c.GetKind(),
		Source:      c.GetSource(),
		Roles:       c.GetRoles(),
		Permissions: c.GetPermissions(),
		Attributes:  c.GetAttributes(),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        c.GetJti(),
			Issuer:    c.GetIssuer(),
			Audience:  c.GetAudience(),
			ExpiresAt: jwt.NewNumericDate(time.Unix(c.GetExpiry(), 0)),
		},
	}
}
