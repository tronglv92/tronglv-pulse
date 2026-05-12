package security

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pulse/helper/utils/httpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/collection"
)

func TestNewPermissionProvider(t *testing.T) {
	cache, err := collection.NewCache(time.Minute)
	require.NoError(t, err)

	staleCache, err := collection.NewCache(time.Hour)
	require.NoError(t, err)

	httpClient := httpc.New("test-client")

	provider := NewPermissionProvider("test-service", "http://localhost", cache, staleCache, httpClient)

	assert.NotNil(t, provider)
	assert.IsType(t, &permissionProvider{}, provider)
}

func TestPermissionProvider_Fetch(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantRoles      map[string][]string
	}{
		{
			name: "successful response with multiple roles",
			responseBody: RolePermissionResponse{
				Data: struct {
					Roles map[string]RolePermissions `json:"roles"`
				}{
					Roles: map[string]RolePermissions{
						"0114": {
							Permissions: []string{"fe8e9fea14703f73e2ba1", "d89899234733d7a5f71b6"},
						},
						"0115": {
							Permissions: []string{"7e635ffa83c987592dc45", "f378faf9e5edb26557f24"},
						},
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantRoles: map[string][]string{
				"0114": {"fe8e9fea14703f73e2ba1", "d89899234733d7a5f71b6"},
				"0115": {"7e635ffa83c987592dc45", "f378faf9e5edb26557f24"},
			},
		},
		{
			name: "successful response with single role",
			responseBody: RolePermissionResponse{
				Data: struct {
					Roles map[string]RolePermissions `json:"roles"`
				}{
					Roles: map[string]RolePermissions{
						"admin": {
							Permissions: []string{"hash1", "hash2", "hash3"},
						},
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantRoles: map[string][]string{
				"admin": {"hash1", "hash2", "hash3"},
			},
		},
		{
			name: "empty roles response",
			responseBody: RolePermissionResponse{
				Data: struct {
					Roles map[string]RolePermissions `json:"roles"`
				}{
					Roles: map[string]RolePermissions{},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantRoles:      map[string][]string{},
		},
		{
			name:           "server error",
			responseBody:   nil,
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
			wantRoles:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.Query().Get("service_code"), "test-service")

				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			// Create provider
			cache, _ := collection.NewCache(time.Minute)
			staleCache, _ := collection.NewCache(time.Hour)
			httpClient := httpc.New("test-client")

			provider := &permissionProvider{
				code:        "test-service",
				baseUrl:     server.URL,
				httpClient:  httpClient,
				cacheClient: cache,
				staleCache:  NewStaleAwareCache(cache, staleCache),
			}

			// Test fetch
			roles, err := provider.fetch(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantRoles, roles)
		})
	}
}

func TestPermissionProvider_GetPermissions(t *testing.T) {
	tests := []struct {
		name            string
		mockRoles       map[string][]string
		userRoles       []string
		wantPermissions []string
	}{
		{
			name: "single role with permissions",
			mockRoles: map[string][]string{
				"admin": {"perm1", "perm2", "perm3"},
			},
			userRoles:       []string{"admin"},
			wantPermissions: []string{"perm1", "perm2", "perm3"},
		},
		{
			name: "multiple roles with overlapping permissions",
			mockRoles: map[string][]string{
				"admin": {"perm1", "perm2", "perm3"},
				"user":  {"perm2", "perm4"},
			},
			userRoles:       []string{"admin", "user"},
			wantPermissions: []string{"perm1", "perm2", "perm3", "perm4"},
		},
		{
			name: "role not found",
			mockRoles: map[string][]string{
				"admin": {"perm1", "perm2"},
			},
			userRoles:       []string{"guest"},
			wantPermissions: []string{},
		},
		{
			name: "empty user roles",
			mockRoles: map[string][]string{
				"admin": {"perm1", "perm2"},
			},
			userRoles:       []string{},
			wantPermissions: []string{},
		},
		{
			name: "multiple roles some not found",
			mockRoles: map[string][]string{
				"admin": {"perm1", "perm2"},
				"user":  {"perm3"},
			},
			userRoles:       []string{"admin", "guest", "user"},
			wantPermissions: []string{"perm1", "perm2", "perm3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server that returns mock roles
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := RolePermissionResponse{
					Data: struct {
						Roles map[string]RolePermissions `json:"roles"`
					}{
						Roles: make(map[string]RolePermissions),
					},
				}
				for role, perms := range tt.mockRoles {
					resp.Data.Roles[role] = RolePermissions{Permissions: perms}
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			// Create provider
			cache, _ := collection.NewCache(time.Minute)
			staleCache, _ := collection.NewCache(time.Hour)
			httpClient := httpc.New("test-client")

			provider := NewPermissionProvider("test-service", server.URL, cache, staleCache, httpClient)

			// Get permissions
			perms, err := provider.GetPermissions(context.Background(), tt.userRoles)
			require.NoError(t, err)

			// Check result (order may vary due to map iteration)
			assert.ElementsMatch(t, tt.wantPermissions, perms)
		})
	}
}

func TestPermissionProvider_GetPermissions_Caching(t *testing.T) {
	callCount := 0

	// Create test server that counts calls
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := RolePermissionResponse{
			Data: struct {
				Roles map[string]RolePermissions `json:"roles"`
			}{
				Roles: map[string]RolePermissions{
					"admin": {Permissions: []string{"perm1", "perm2"}},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create provider with cache
	cache, _ := collection.NewCache(time.Minute)
	staleCache, _ := collection.NewCache(time.Hour)
	httpClient := httpc.New("test-client")

	provider := NewPermissionProvider("test-service", server.URL, cache, staleCache, httpClient)

	// First call - should hit the server
	perms1, err := provider.GetPermissions(context.Background(), []string{"admin"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"perm1", "perm2"}, perms1)
	assert.Equal(t, 1, callCount, "First call should hit the server")

	// Second call - should use cache
	perms2, err := provider.GetPermissions(context.Background(), []string{"admin"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"perm1", "perm2"}, perms2)
	assert.Equal(t, 1, callCount, "Second call should use cache, not hit server")
}

func TestPermissionProvider_Refresh(t *testing.T) {
	callCount := 0

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := RolePermissionResponse{
			Data: struct {
				Roles map[string]RolePermissions `json:"roles"`
			}{
				Roles: map[string]RolePermissions{
					"admin": {Permissions: []string{"perm1"}},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create provider
	cache, _ := collection.NewCache(time.Minute)
	staleCache, _ := collection.NewCache(time.Hour)
	httpClient := httpc.New("test-client")

	provider := NewPermissionProvider("test-service", server.URL, cache, staleCache, httpClient)

	// Refresh without any prior cache - should fetch from server
	err := provider.Refresh(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, callCount, "Refresh should fetch when cache is empty")

	// Refresh again - should not fetch again (cache exists)
	// Note: Refresh() uses Take() which doesn't force refresh if cache exists
	err = provider.Refresh(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, callCount, "Refresh doesn't force refetch when cache exists")
}

func TestPermissionProvider_DeduplicatePermissions(t *testing.T) {
	// Create test server with duplicate permissions across roles
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := RolePermissionResponse{
			Data: struct {
				Roles map[string]RolePermissions `json:"roles"`
			}{
				Roles: map[string]RolePermissions{
					"admin": {Permissions: []string{"perm1", "perm2", "perm3"}},
					"user":  {Permissions: []string{"perm2", "perm3", "perm4"}},
					"guest": {Permissions: []string{"perm3", "perm5"}},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create provider
	cache, _ := collection.NewCache(time.Minute)
	staleCache, _ := collection.NewCache(time.Hour)
	httpClient := httpc.New("test-client")

	provider := NewPermissionProvider("test-service", server.URL, cache, staleCache, httpClient)

	// Get permissions for all three roles
	perms, err := provider.GetPermissions(context.Background(), []string{"admin", "user", "guest"})
	require.NoError(t, err)

	// Should have all unique permissions (no duplicates)
	assert.ElementsMatch(t, []string{"perm1", "perm2", "perm3", "perm4", "perm5"}, perms)
	assert.Len(t, perms, 5, "Should deduplicate permissions")
}
