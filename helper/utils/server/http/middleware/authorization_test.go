package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pulse/helper/utils/identity"

	"github.com/stretchr/testify/assert"
)

// stubClaims is a non-MapClaims identity.Claims to trigger the !ok branch in Guard().
type stubClaims struct{}

func (s *stubClaims) GetJti() string                  { return "" }
func (s *stubClaims) GetId() string                   { return "1" }
func (s *stubClaims) GetName() string                 { return "test" }
func (s *stubClaims) GetKind() string                 { return "user" }
func (s *stubClaims) GetSource() string               { return "" }
func (s *stubClaims) GetRoles() []string              { return nil }
func (s *stubClaims) GetPermissions() []string        { return nil }
func (s *stubClaims) GetExpiry() int64                { return 0 }
func (s *stubClaims) GetIssuer() string               { return "" }
func (s *stubClaims) GetAudience() []string           { return nil }
func (s *stubClaims) GetAttributes() map[string]string { return nil }
func (s *stubClaims) IsSingleSession() bool           { return false }

func TestGuard_NonMapClaimsDoesNotPanic(t *testing.T) {
	handler := Guard(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, "some:permission")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := identity.WithContext(req.Context(), &stubClaims{})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler(rec, req)
	})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGuard_AllowsMatchingPermission(t *testing.T) {
	handler := Guard(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, "logs:view")

	claims := &identity.MapClaims{}
	claims.SetPermissions([]string{"logs:view", "logs:ingest"})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := identity.WithContext(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGuard_DeniesNoMatchingPermission(t *testing.T) {
	handler := Guard(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, "tenant:manage")

	claims := &identity.MapClaims{}
	claims.SetPermissions([]string{"logs:view"})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := identity.WithContext(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
