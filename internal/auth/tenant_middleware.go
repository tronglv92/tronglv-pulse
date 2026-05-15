package auth

import (
	"context"
	"net/http"
	"strconv"

	"pulse/helper/utils/errors"
	"pulse/helper/utils/identity"
	"pulse/helper/utils/server/http/response"

	"github.com/zeromicro/go-zero/rest"
)

type tenantCtxKey struct{}

// TenantMiddleware returns a rest.Middleware that extracts tenant_id from the
// JWT claims attributes and injects it into the request context.
// Must run after AuthMiddleware.
func TenantMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims := identity.FromContext(r.Context())
			if claims == nil {
				response.Error(r.Context(), w, errors.NewUnauthorized("MISSING_CLAIMS", "Authentication required."))
				return
			}

			attrs := claims.GetAttributes()
			tidStr, ok := attrs["tenant_id"]
			if !ok || tidStr == "" {
				response.Error(r.Context(), w, errors.NewUnauthorized("MISSING_TENANT", "Token missing tenant_id attribute."))
				return
			}

			tenantID, err := strconv.ParseInt(tidStr, 10, 64)
			if err != nil {
				response.Error(r.Context(), w, errors.NewUnauthorized("INVALID_TENANT", "Invalid tenant_id in token."))
				return
			}

			ctx := context.WithValue(r.Context(), tenantCtxKey{}, tenantID)
			next(w, r.WithContext(ctx))
		}
	}
}

// TenantIDFromContext extracts the tenant ID injected by TenantMiddleware.
func TenantIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(tenantCtxKey{}).(int64)
	return id, ok
}
