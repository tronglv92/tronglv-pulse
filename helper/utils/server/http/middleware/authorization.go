package middleware

import (
	"pulse/helper/utils/authorizer"
	"pulse/helper/utils/errors"
	"pulse/helper/utils/identity"
	"pulse/helper/utils/localize"
	"pulse/helper/utils/security"
	"pulse/helper/utils/server/http/response"
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

func Authorization(authorize authorizer.Authorizer) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			data, err := identity.MustFromContext(r.Context())
			if err != nil {
				response.Error(r.Context(), w, errors.Unauthorized(fmt.Errorf("missing or invalid claims in context: %w", err)))
				return
			}

			if err = authorize.ValidateClaims(data); err != nil {
				response.Error(r.Context(), w, errors.Unauthorized(err))
				return
			}

			data, err = authorize.GetPermissions(r.Context(), data)
			if err != nil {
				response.Error(r.Context(), w, errors.Unauthorized(err))
				return
			}

			ctx := identity.WithContext(r.Context(), data)
			next(w, r.WithContext(ctx))
		}
	}
}

// RequirePathPermission creates an authorization middleware with automatic
// hash-based permission checking.
//
// This middleware:
//  1. Validates JWT claims (expiry, issuer, audience)
//  2. Fetches user permissions based on roles (permissions are hashed method:path)
//  3. Automatically hashes the current request (method + path)
//  4. Checks if the request hash exists in the user's permissions
//
// Permission codes in the permission service should be hashes of "METHOD:/path".
// For example: sha256("GET:/api/users")
//
// Example usage:
//
//	srv.Use(middleware.RequirePathPermission(authorizerInstance))
func RequirePathPermission(authorize authorizer.Authorizer) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			data, err := identity.MustFromContext(r.Context())
			if err != nil {
				response.Error(r.Context(), w, errors.Unauthorized(fmt.Errorf("missing or invalid claims in context: %w", err)))
				return
			}

			if err = authorize.ValidateClaims(data); err != nil {
				response.Error(r.Context(), w, errors.Unauthorized(err))
				return
			}

			// Fetch user permissions based on roles (permissions are hashes)
			data, err = authorize.GetPermissions(r.Context(), data)
			if err != nil {
				response.Error(r.Context(), w, errors.Unauthorized(err))
				return
			}

			// Automatic hash-based path authorization
			// Hash the current request (method + path)
			requestHash := security.HashPath(r.Method, r.URL.Path)

			// Check if user has this permission (hash)
			jwtMap, ok := data.(*identity.MapClaims)
			if !ok {
				response.Error(r.Context(), w, errors.Unauthorized(fmt.Errorf("invalid claims type")))
				return
			}

			if !jwtMap.HasPermission(requestHash) {
				logx.WithContext(r.Context()).Infof("Permission denied for %s %s (hash: %s)", r.Method, r.URL.Path, requestHash)

				response.Error(r.Context(), w, errors.NewUnauthorized("PERMISSION_MISSING", localize.GetString(r.Context(), "auth_forbidden")))
				return
			}
			
			next(w, r.WithContext(identity.WithContext(r.Context(), data)))
		}
	}
}

// Guard checks if the user has any of the specified permission codes.
//
// DEPRECATED: This function is deprecated in favor of automatic hash-based
// authorization using RequirePathPermission middleware.
// This function is kept for backward compatibility only.
//
// Migration guide:
//  1. Update permission service: Change permission codes from readable strings
//     (e.g., "user.read") to hashes (e.g., sha256("GET:/api/users"))
//  2. Update role->permission mappings in permission service to use hashes
//  3. Replace Authorization() middleware with RequirePathPermission()
//  4. Remove Guard() wrapper calls from route definitions
//  5. Validate that automatic path checking works as expected
//
// Example migration:
//
//	Before (permission service):
//	  {"admin": ["user.read", "user.create", "user.delete"]}
//
//	After (permission service):
//	  {"admin": ["sha256(GET:/api/users)", "sha256(POST:/api/users)", "sha256(DELETE:/api/users/:id)"]}
//
//	Before (route definition):
//	  {Method: "GET", Path: "/api/users", Handler: middleware.Guard(handler, "user.read")}
//
//	After (route definition):
//	  {Method: "GET", Path: "/api/users", Handler: handler}
//	  // Permission automatically checked via hash("GET:/api/users")
func Guard(next http.HandlerFunc, permissionCodes ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := identity.MustFromContext(r.Context())
		if err != nil {
			response.Error(r.Context(), w, errors.NewUnauthorized(err.Error(), localize.GetString(r.Context(), "auth_claims_missing")))
			return
		}

		jwtMap, ok := data.(*identity.MapClaims)
		if !ok {
			response.Error(r.Context(), w, errors.NewUnauthorized(err.Error(), localize.GetString(r.Context(), "auth_claims_invalid")))
			return
		}

		for _, p := range permissionCodes {
			if jwtMap.HasPermission(p) {
				next(w, r)
				return
			}
		}

		response.Error(r.Context(), w, errors.NewUnauthorized(
			"PERMISSION_MISSING",
			localize.GetString(r.Context(), "auth_forbidden")),
		)
		return
	}
}
