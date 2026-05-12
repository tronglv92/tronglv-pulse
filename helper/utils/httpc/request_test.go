package httpc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWithBaseURL(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.URL.Path))
	}))
	defer server.Close()

	tests := []struct {
		name        string
		baseURL     string
		requestPath string
		expectedURL string
	}{
		{
			name:        "relative path with base URL",
			baseURL:     server.URL,
			requestPath: "/api/users",
			expectedURL: "/api/users",
		},
		{
			name:        "relative path without leading slash",
			baseURL:     server.URL,
			requestPath: "api/users",
			expectedURL: "/api/users",
		},
		{
			name:        "absolute URL overrides base URL",
			baseURL:     server.URL,
			requestPath: server.URL + "/other/path",
			expectedURL: "/other/path",
		},
		{
			name:        "base URL with trailing slash",
			baseURL:     server.URL + "/",
			requestPath: "/api/users",
			expectedURL: "/api/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewWithBaseURL("test-client", tt.baseURL)

			resp, err := client.Get(context.Background(), tt.requestPath)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		requestURL  string
		expectedURL string
	}{
		{
			name:        "no base URL",
			baseURL:     "",
			requestURL:  "http://example.com/api/users",
			expectedURL: "http://example.com/api/users",
		},
		{
			name:        "base URL with relative path",
			baseURL:     "http://api.example.com",
			requestURL:  "/users",
			expectedURL: "http://api.example.com/users",
		},
		{
			name:        "base URL with relative path without slash",
			baseURL:     "http://api.example.com",
			requestURL:  "users",
			expectedURL: "http://api.example.com/users",
		},
		{
			name:        "base URL with trailing slash",
			baseURL:     "http://api.example.com/",
			requestURL:  "/users",
			expectedURL: "http://api.example.com/users",
		},
		{
			name:        "absolute URL ignores base URL",
			baseURL:     "http://api.example.com",
			requestURL:  "http://other.com/users",
			expectedURL: "http://other.com/users",
		},
		{
			name:        "https absolute URL ignores base URL",
			baseURL:     "http://api.example.com",
			requestURL:  "https://secure.com/api",
			expectedURL: "https://secure.com/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := clientService{
				name:    "test",
				baseURL: tt.baseURL,
			}

			result := client.buildURL(tt.requestURL)
			assert.Equal(t, tt.expectedURL, result)
		})
	}
}

func TestClientWithBaseURL_AllMethods(t *testing.T) {
	requestedPaths := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPaths[r.Method] = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewWithBaseURL("test-client", server.URL)
	ctx := context.Background()

	// Test GET
	_, err := client.Get(ctx, "/api/get")
	assert.NoError(t, err)
	assert.Equal(t, "/api/get", requestedPaths[http.MethodGet])

	// Test POST
	_, err = client.Post(ctx, "/api/post", map[string]string{"key": "value"})
	assert.NoError(t, err)
	assert.Equal(t, "/api/post", requestedPaths[http.MethodPost])

	// Test PUT
	_, err = client.Put(ctx, "/api/put", map[string]string{"key": "value"})
	assert.NoError(t, err)
	assert.Equal(t, "/api/put", requestedPaths[http.MethodPut])

	// Test PATCH
	_, err = client.Patch(ctx, "/api/patch", map[string]string{"key": "value"})
	assert.NoError(t, err)
	assert.Equal(t, "/api/patch", requestedPaths[http.MethodPatch])

	// Test DELETE
	_, err = client.Delete(ctx, "/api/delete", nil)
	assert.NoError(t, err)
	assert.Equal(t, "/api/delete", requestedPaths[http.MethodDelete])

	// Test HEAD
	_, err = client.Head(ctx, "/api/head")
	assert.NoError(t, err)
	assert.Equal(t, "/api/head", requestedPaths[http.MethodHead])
}
