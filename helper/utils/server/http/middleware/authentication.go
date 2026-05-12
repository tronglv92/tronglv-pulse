package middleware

import (
	"pulse/helper/utils/authenticator"
	"pulse/helper/utils/errors"
	"pulse/helper/utils/identity"
	"pulse/helper/utils/localize"
	"pulse/helper/utils/server/http/response"
	"pulse/helper/utils/toolkit/contextx"
	"github.com/zeromicro/go-zero/rest"
	"net/http"
	"strings"
)

func AuthMiddleware(auth authenticator.Authenticator, enrich bool) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tokenType, tokenData, err := authenticator.ExtractAuthToken(r.Header.Get("Authorization"))
			if err != nil {
				response.Error(r.Context(), w, errors.NewUnauthorized(err.Error(), localize.GetString(r.Context(), "auth_token_missing")))
				return
			}

			var mapClaims identity.Claims
			switch strings.ToUpper(tokenType) {
			case "BASIC":
				mapClaims, err = auth.SignInWithBasic(r.Context(), tokenData)
			case "BEARER":
				if enrich {
					mapClaims, err = auth.VerifyTokenWithPermissions(r.Context(), tokenData)
				} else {
					mapClaims, err = auth.VerifyToken(r.Context(), tokenData)
				}
			default:
				http.Error(w, "unsupported token type", http.StatusUnauthorized)
				return
			}
			if err != nil {
				response.Error(r.Context(), w, errors.NewUnauthorized(err.Error(), localize.GetString(r.Context(), "auth_token_expired")))
				return
			}

			next(w, r.WithContext(identity.WithContext(
				contextx.WithAuthorizationToken(ctx, tokenData), mapClaims),
			))
		}
	}
}

func StaticTokenAuthMiddleware(headerKey, token string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			headerToken := r.Header.Get(headerKey)
			if len(headerToken) == 0 || len(token) == 0 {
				response.Error(r.Context(), w, errors.NewUnauthorized(
					"TOKEN_MISSING",
					localize.GetString(r.Context(), "auth_token_missing"),
				))
				return
			}

			if headerToken != token {
				response.Error(r.Context(), w, errors.NewUnauthorized(
					"TOKEN_INVALID",
					localize.GetString(r.Context(), "auth_token_invalid"),
				))
				return
			}
			next(w, r)
		}
	}
}
