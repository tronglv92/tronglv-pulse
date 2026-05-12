package identity

import (
	"context"
	"slices"
)

func HasRole(c Claims, role string) bool {
	return slices.Contains(c.GetRoles(), role)
}

func HasRoleFromContext(ctx context.Context, role string) bool {
	claims, err := MapClaimsFromContext(ctx)
	if err != nil {
		return false
	}
	return HasRole(claims, role)
}

func HasPermission(c Claims, permission string) bool {
	return slices.Contains(c.GetPermissions(), permission)
}

func HasPermissionFromContext(ctx context.Context, permission string) bool {
	claims, err := MapClaimsFromContext(ctx)
	if err != nil {
		return false
	}
	return HasPermission(claims, permission)
}
