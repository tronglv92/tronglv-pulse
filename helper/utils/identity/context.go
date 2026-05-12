package identity

import (
	"context"

	"pulse/helper/utils/errors"
)

func WithContext(parent context.Context, claims Claims) context.Context {
	return context.WithValue(parent, authClaimsContext, claims)
}

func FromContext(ctx context.Context) Claims {
	claims, err := MustFromContext(ctx)
	if err != nil {
		return nil
	}
	return claims
}

func MustFromContext(ctx context.Context) (Claims, error) {
	claims, ok := ctx.Value(authClaimsContext).(Claims)
	if !ok {
		return nil, errors.NewUnauthorized("INVALID_CLAIMS", "Authentication failed due to invalid token claims.")
	}
	return claims, nil
}

func MapClaimsFromContext(ctx context.Context) (*MapClaims, error) {
	data, err := MustFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return MapClaimsFrom(data), nil
}

func UserClaimsFromContext(ctx context.Context) (*UserClaims, error) {
	claims, err := MapClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return UserClaimsFrom(claims), nil
}
