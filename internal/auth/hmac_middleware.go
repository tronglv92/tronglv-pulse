package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"pulse/helper/utils/errors"
	"pulse/helper/utils/server/http/response"

	"github.com/zeromicro/go-zero/rest"
)

// HMACMiddleware returns a rest.Middleware that verifies the X-Signature header
// against the request body using HMAC-SHA256. The expected format is:
//
//	X-Signature: sha256=<hex-encoded-hmac>
//
// The request body is restored for downstream handlers after verification.
func HMACMiddleware(hmacSecret string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			sigHeader := r.Header.Get("X-Signature")
			if sigHeader == "" {
				response.Error(r.Context(), w, errors.NewUnauthorized("MISSING_SIGNATURE", "X-Signature header is required."))
				return
			}

			parts := strings.SplitN(sigHeader, "=", 2)
			if len(parts) != 2 || parts[0] != "sha256" {
				response.Error(r.Context(), w, errors.NewUnauthorized("INVALID_SIGNATURE_FORMAT", "X-Signature must be: sha256=<hex>."))
				return
			}

			expectedMAC, err := hex.DecodeString(parts[1])
			if err != nil {
				response.Error(r.Context(), w, errors.NewUnauthorized("INVALID_SIGNATURE_HEX", "X-Signature contains invalid hex."))
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				response.Error(r.Context(), w, errors.NewBadRequest("BODY_READ_FAILED", "Failed to read request body."))
				return
			}
			// Restore body for downstream handlers.
			r.Body = io.NopCloser(bytes.NewReader(body))

			mac := hmac.New(sha256.New, []byte(hmacSecret))
			mac.Write(body)
			actualMAC := mac.Sum(nil)

			if !hmac.Equal(expectedMAC, actualMAC) {
				response.Error(r.Context(), w, errors.NewUnauthorized("INVALID_SIGNATURE", "HMAC signature verification failed."))
				return
			}

			next(w, r)
		}
	}
}
