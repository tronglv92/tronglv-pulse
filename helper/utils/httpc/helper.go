package httpc

import (
	"net/http"
	"strings"
)

// StripBearerPrefix removes the "Bearer " prefix (case-insensitive) from a token string.
func StripBearerPrefix(token string) string {
	if strings.HasPrefix(strings.ToUpper(token), "BEARER ") {
		return token[7:]
	}
	return token
}

// ExtractBearerToken extracts the token from the Authorization header, removing the Bearer prefix.
func ExtractBearerToken(r *http.Request) string {
	return StripBearerPrefix(GetAuthorizationHeader(r))
}

// GetAuthorizationHeader returns the raw Authorization header from the request.
func GetAuthorizationHeader(r *http.Request) string {
	return r.Header.Get("Authorization")
}

// IsJSONContentType returns true if the content type is exactly "application/json".
func IsJSONContentType(contentType string) bool {
	return contentType == "application/json"
}

// IsMultipartFormData returns true if the content type is "multipart/form-data" (including boundaries).
func IsMultipartFormData(contentType string) bool {
	return strings.HasPrefix(contentType, "multipart/form-data")
}
