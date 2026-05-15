package auth

import (
	"net/http"
	"strings"

	"pulse/helper/utils/errors"
	"pulse/helper/utils/identity"
	"pulse/helper/utils/server/http/response"

	"github.com/zeromicro/go-zero/rest"
)

// AuthMiddleware returns a rest.Middleware that validates JWT Bearer tokens.
// Refresh tokens are rejected — only access tokens (kind="user") are accepted.
func AuthMiddleware(jwtSecret string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(r.Context(), w, errors.NewUnauthorized("MISSING_TOKEN", "Authorization header is required."))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Error(r.Context(), w, errors.NewUnauthorized("INVALID_AUTH_HEADER", "Authorization header must be: Bearer <token>."))
				return
			}

			token := parts[1]
			claims, err := identity.FromToken(token, nil, []byte(jwtSecret))
			if err != nil {
				response.Error(r.Context(), w, errors.NewUnauthorized("INVALID_TOKEN", "Invalid or expired token."))
				return
			}

			// Reject refresh tokens used as access tokens.
			if claims.GetKind() == "refresh" {
				response.Error(r.Context(), w, errors.NewUnauthorized("INVALID_TOKEN_KIND", "Refresh tokens cannot be used for API access."))
				return
			}

			ctx := identity.WithContext(r.Context(), claims)
			next(w, r.WithContext(ctx))
		}
	}
}
