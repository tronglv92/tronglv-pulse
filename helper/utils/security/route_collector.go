package security

import (
	"context"
	"pulse/helper/utils/authenticator"
	"pulse/helper/utils/httpc"
	"fmt"
	"path/filepath"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

const (
	RegistrationURL = "/permission-svc/api/v1/endpoints/register"
)

// RegisterOption is a function that configures route registration.
type RegisterOption func(*registerConfig)

// registerConfig holds configuration for route registration.
type registerConfig struct {
	baseUrl         string
	serviceCode     string
	registrationUrl string
	httpClient      httpc.Service
	nameFunc        func(method, path string) string
	ignorePaths     []string
}

// WithAuthToken configures the HTTP client with an authentication token.
func WithAuthToken(token string) RegisterOption {
	return func(r *registerConfig) {
		r.httpClient = httpc.New(r.serviceCode, httpc.WithAuthToken(token))
	}
}

// WithHTTPClient sets a custom HTTP client for route registration.
func WithHTTPClient(client httpc.Service) RegisterOption {
	return func(r *registerConfig) {
		r.httpClient = client
	}
}

// WithNameFunc sets a custom naming function for endpoints.
func WithNameFunc(nameFunc func(method, path string) string) RegisterOption {
	return func(r *registerConfig) {
		r.nameFunc = nameFunc
	}
}

// WithTokenManager is a RegisterOption that configures the HTTP client using a TokenManager.
// This ensures the client uses a valid access token managed by the TokenManager.
//
// Example:
//
//	security.RegisterRoutesAtStartup(
//	    restHandler,
//	    c.Security.ServiceCode,
//	    c.Security.BaseUrl,
//	    security.WithTokenManager(tokenMgr),
//	)
func WithTokenManager(tokenManager authenticator.TokenManager) RegisterOption {
	return func(r *registerConfig) {
		client, err := tokenManager.GetHTTPClient(context.Background(), r.serviceCode)
		if err != nil {
			logx.Errorf("Failed to create HTTP client with token manager: %v", err)
			return
		}
		r.httpClient = client
	}
}

// WithIgnorePaths sets paths to exclude from registration.
// Supports exact matches and wildcard patterns using * and ?.
//
// Examples:
//   - Exact match: "/health", "/api/internal/debug"
//   - Wildcard: "/api/internal/*", "/swagger/*", "*/health"
//   - Pattern: "/api/v?/users", "/health*"
func WithIgnorePaths(paths ...string) RegisterOption {
	return func(r *registerConfig) {
		r.ignorePaths = paths
	}
}

// RouteCollector collects routes from a REST server handler.
type RouteCollector interface {
	Routes() []rest.Route
}

// shouldIgnorePath checks if a path matches any of the ignore patterns.
// Supports wildcard patterns using filepath.Match (* and ?).
func shouldIgnorePath(path string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		if ok, _ := filepath.Match(pattern, path); ok {
			return true
		}
	}
	return false
}

// CollectEndpoints converts rest.Route slice to Endpoint slice for registration.
// It automatically generates hashes for each route.
//
// Parameters:
//   - routes: List of routes from go-zero rest server
//   - nameFunc: Optional function to generate human-readable names for endpoints.
//     If nil, uses default naming: "METHOD /path"
//
// Example:
//
//	routes := handler.Routes()
//	endpoints := security.CollectEndpoints(routes, func(method, path string) string {
//	    return fmt.Sprintf("%s %s", method, path)
//	})
func CollectEndpoints(routes []rest.Route, nameFunc func(method, path string) string) []Endpoint {
	endpoints := make([]Endpoint, 0, len(routes))

	for _, route := range routes {
		name := route.Path
		if nameFunc != nil {
			name = nameFunc(route.Method, route.Path)
		} else {
			name = fmt.Sprintf("%s - %s", route.Method, route.Path)
		}

		endpoints = append(endpoints, Endpoint{
			Method:    route.Method,
			Path:      route.Path,
			Hash:      HashPath(route.Method, route.Path),
			Name:      name,
			GroupName: PathToServiceName(route.Path),
		})
	}

	return endpoints
}

// RegisterRoutesAtStartup is a helper function to register routes with the permission service.
// Call this in main.go after creating your REST handler.
//
// By default, swagger paths (/swagger, /swagger/*) are ignored.
// You can override this by providing custom ignore patterns with WithIgnorePaths.
//
// Parameters:
//   - collector: Handler that implements Routes() method
//   - serviceCode: Service identifier
//   - registrationURL: Permission service endpoint for registration
//   - opts: Optional configuration options (WithAuthToken, WithHTTPClient, WithNameFunc, WithIgnorePaths)
//
// Example usage in main.go with auth token:
//
//	restHandler := handler.NewRestHandler(svcCtx)
//	security.RegisterRoutesAtStartup(
//	    restHandler,
//	    c.Security.ServiceCode,
//	    c.Security.RegistrationUrl,
//	    security.WithAuthToken(c.Security.Token),
//	)
//
// Example with custom HTTP client:
//
//	httpClient := httpc.New("route-registrar", httpc.WithAuthToken(token))
//	security.RegisterRoutesAtStartup(
//	    restHandler,
//	    c.Security.ServiceCode,
//	    c.Security.RegistrationUrl,
//	    security.WithHTTPClient(httpClient),
//	)
//
// Example with multiple options:
//
//	security.RegisterRoutesAtStartup(
//	    restHandler,
//	    c.Security.ServiceCode,
//	    c.Security.RegistrationUrl,
//	    security.WithAuthToken(c.Security.Token),
//	    security.WithNameFunc(func(method, path string) string {
//	        return fmt.Sprintf("[%s] %s", method, path)
//	    }),
//	)
//
// Example with ignored paths:
//
//	security.RegisterRoutesAtStartup(
//	    restHandler,
//	    c.Security.ServiceCode,
//	    c.Security.RegistrationUrl,
//	    security.WithAuthToken(c.Security.Token),
//	    security.WithIgnorePaths("/health", "/metrics", "/api/internal/*"),
//	)
func RegisterRoutesAtStartup(collector RouteCollector, serviceCode, baseUrl string, opts ...RegisterOption) {
	r := &registerConfig{
		baseUrl:         baseUrl,
		serviceCode:     serviceCode,
		registrationUrl: fmt.Sprintf("%s%s", baseUrl, RegistrationURL),
		httpClient:      NewHTTPClient(serviceCode), // default client
		ignorePaths:     []string{"/swagger", "/swagger/docs/:name", "/security/internal/webhook"},
	}
	for _, opt := range opts {
		opt(r)
	}

	// Filter out ignored paths
	var endpoints []Endpoint
	for _, endpoint := range CollectEndpoints(collector.Routes(), r.nameFunc) {
		if !shouldIgnorePath(endpoint.Path, r.ignorePaths) {
			endpoints = append(endpoints, endpoint)
		}
	}

	// Create initializer
	initializer := NewRouteRegistrarInitializer(NewRouteRegistrar(
		r.serviceCode,
		r.registrationUrl,
		endpoints,
		r.httpClient,
	))

	// Register routes
	initializer.Register(context.Background())
}

// NewHTTPClient creates a default HTTP client for route registration.
func NewHTTPClient(name string) httpc.Service {
	return httpc.New(name)
}

func NewHTTPClientWithAuth(auth authenticator.Authenticator, name, clientId, clientSecret string) httpc.Service {
	t, e := auth.SignInWithService(context.Background(), clientId, clientSecret)
	if e != nil {
		logx.Must(e)
	}
	return httpc.New(name, httpc.WithAuthToken(t.GetAccessToken()))
}
