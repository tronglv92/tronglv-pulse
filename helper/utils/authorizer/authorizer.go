package authorizer

import (
	"context"

	"pulse/helper/utils/identity"
)

// Authorizer validates JWT claims and enriches them with role-based permissions.
type Authorizer interface {
	ValidateClaims(claims identity.Claims) error
	GetPermissions(ctx context.Context, claims identity.Claims) (identity.Claims, error)
}
