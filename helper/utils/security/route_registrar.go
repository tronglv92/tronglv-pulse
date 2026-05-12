package security

import (
	"context"
	"pulse/helper/utils/httpc"
	"fmt"
	"io"

	"github.com/zeromicro/go-zero/core/logx"
)

// Endpoint represents an API endpoint to be registered with the permission service.
type Endpoint struct {
	Method    string `json:"method"` // HTTP method (GET, POST, PUT, DELETE, etc.)
	Path      string `json:"path"`   // API path (e.g., "/api/users")
	Hash      string `json:"hash"`   // Hash of method:path (21 characters)
	Name      string `json:"name"`   // Human-readable name/description
	GroupName string `json:"group_name"`
}

// RouteRegistrationRequest is sent to the permission service to register endpoints.
type RouteRegistrationRequest struct {
	ServiceCode string     `json:"service_code"`
	Endpoints   []Endpoint `json:"endpoints"`
}

// RouteRegistrar handles registration of API endpoints with the permission service.
type RouteRegistrar interface {
	Register(ctx context.Context) error
}

type routeRegistrar struct {
	serviceCode string
	registryURL string
	endpoints   []Endpoint
	httpClient  httpc.Service
}

// NewRouteRegistrar creates a new route registrar.
//
// Parameters:
//   - serviceCode: The service identifier (from config)
//   - registryURL: The permission service endpoint URL for registration
//   - endpoints: List of endpoints to register
//   - httpClient: HTTP client for making requests
//
// Example:
//
//	endpoints := []security.Endpoint{
//	    {Method: "GET", Path: "/api/users", Hash: security.HashPath("GET", "/api/users"), Name: "List users"},
//	    {Method: "POST", Path: "/api/users", Hash: security.HashPath("POST", "/api/users"), Name: "Create user"},
//	}
//	registrar := security.NewRouteRegistrar("my-service", "http://permission-svc/api/endpoints/register", endpoints, httpClient)
func NewRouteRegistrar(serviceCode, registryURL string, endpoints []Endpoint, httpClient httpc.Service) RouteRegistrar {
	return &routeRegistrar{
		serviceCode: serviceCode,
		registryURL: registryURL,
		endpoints:   endpoints,
		httpClient:  httpClient,
	}
}

// Register sends all endpoints to the permission service for registration.
// The permission service will handle duplicate checking.
func (r *routeRegistrar) Register(ctx context.Context) error {
	if len(r.endpoints) == 0 {
		return nil
	}

	resp, err := r.httpClient.Post(ctx, r.registryURL, RouteRegistrationRequest{
		ServiceCode: r.serviceCode,
		Endpoints:   r.endpoints,
	})
	if err != nil {
		return fmt.Errorf("failed to register endpoints: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("permission service returned error status: %d", resp.StatusCode)
	}
	return nil
}

// RouteRegistrarInitializer implements core.Service to register routes at startup.
type RouteRegistrarInitializer struct {
	registrar RouteRegistrar
}

// NewRouteRegistrarInitializer creates a startup initializer for route registration.
func NewRouteRegistrarInitializer(registrar RouteRegistrar) *RouteRegistrarInitializer {
	return &RouteRegistrarInitializer{
		registrar: registrar,
	}
}

// Register implements core.Service interface.
// It registers all routes with the permission service at startup.
func (r *RouteRegistrarInitializer) Register(ctx context.Context) {
	if err := r.registrar.Register(ctx); err != nil {
		logx.Errorf("Failed to register routes with permission service: %v", err)
	}
}

// BuildEndpoint is a helper function to create an Endpoint with auto-generated hash.
//
// Example:
//
//	endpoint := security.BuildEndpoint("GET", "/api/users", "List all users")
func BuildEndpoint(method, path, name string) Endpoint {
	return Endpoint{
		Method: method,
		Path:   path,
		Hash:   HashPath(method, path),
		Name:   name,
	}
}
