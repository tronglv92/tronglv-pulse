package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"
)

// HashPath creates a deterministic hash from HTTP method and path.
// It normalizes the input by converting to uppercase method and lowercase path,
// removing trailing slashes, then generates a SHA-256 hash truncated to 21 characters.
//
// Example:
//
//	HashPath("GET", "/api/users") -> "f8e7d6c5b4a39210a8c7e" (21 chars)
//	HashPath("get", "/API/Users/") -> "f8e7d6c5b4a39210a8c7e" (same hash due to normalization)
//
// The hash is truncated to 21 hex characters for shorter storage
// while maintaining excellent collision resistance for permission matching.
func HashPath(method, path string) string {
	// Normalize method: uppercase
	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))

	// Normalize path: lowercase, remove trailing slash
	normalizedPath := strings.ToLower(strings.TrimSpace(path))
	normalizedPath = strings.TrimSuffix(normalizedPath, "/")

	// Handle empty path (root)
	if normalizedPath == "" {
		normalizedPath = "/"
	}

	// Create hash input: METHOD:PATH
	input := fmt.Sprintf("%s:%s", normalizedMethod, normalizedPath)

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(input))

	// Return first 21 hex characters
	return hex.EncodeToString(hash[:])[:21]
}

// PathToServiceName derives a service name from a REST path.
func PathToServiceName(path string) string {
	parts := strings.Split(path, "/")

	var resource string
	for _, p := range parts {
		if p == "" ||
			p == "api" ||
			isVersion(p) ||
			strings.HasPrefix(p, ":") {
			continue
		}

		// skip service prefix like ticket-svc
		if strings.HasSuffix(p, "-svc") {
			continue
		}

		resource = p
		break
	}

	if resource == "" {
		return "DefaultService"
	}

	return toPascalCase(resource) + "Service"
}

func isVersion(s string) bool {
	return len(s) > 1 && s[0] == 'v' && isDigits(s[1:])
}

func isDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_'
	})

	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		if len(p) > 1 {
			b.WriteString(strings.ToLower(p[1:]))
		}
	}
	return b.String()
}
