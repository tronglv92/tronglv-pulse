package auth

import "time"

// LoginRequest is the payload for POST /v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest is the payload for POST /v1/auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// CreateAPIKeyRequest is the payload for POST /v1/auth/keys.
type CreateAPIKeyRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// TokenResponse is returned by login and refresh endpoints.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// APIKeyResponse is returned when a new API key is created (raw key shown once).
type APIKeyResponse struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	RawKey    string     `json:"raw_key"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// APIKeyItem is an individual key in a list response (no raw key or hash exposed).
type APIKeyItem struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// APIKeyListResponse wraps a list of API key summaries.
type APIKeyListResponse struct {
	Keys []APIKeyItem `json:"keys"`
}
