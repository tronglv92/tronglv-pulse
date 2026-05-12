# Route Registration Integration Guide

This guide shows how to automatically register your API routes with the permission service at startup.

## Step 1: Update Your RestHandler

Add a `Routes()` method to your REST handler that returns all routes.

**File**: `internal/handler/http_handler.go`

```go
package handler

import (
	outboundticketpb "pmc-inapp-outbound-ticket-svc/api/outbound_ticket"
	"pmc-inapp-outbound-ticket-svc/internal/registry"

	"dev.azure.com/PHARMACITYJSC/PharmacityLibrary/_git/pmc-inapp-helper.git/v2/server/http/handler"
	"github.com/zeromicro/go-zero/rest"
)

type RestHandler struct {
	svc registry.AppContext
}

func NewRestHandler(svc registry.AppContext) RestHandler {
	return RestHandler{svc: svc}
}

func (h RestHandler) Register(svr *rest.Server) {
	handler.RegisterSwaggerHandler(svr)

	svr.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{
				h.svc.GetAuthMiddleware(),
			},
			h.combine(
				outboundticketpb.RegisterOutboundTicketHTTPServer(svr, NewOutboundTicketHandler(h.svc)),
			)...,
		),
	)
}

// Routes returns all routes for permission service registration
// ADD THIS METHOD ↓
func (h RestHandler) Routes() []rest.Route {
	return h.combine(
		outboundticketpb.RegisterOutboundTicketHTTPServer(nil, NewOutboundTicketHandler(h.svc)),
	)
}

func (h RestHandler) combine(slices ...[]rest.Route) []rest.Route {
	var result []rest.Route
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}
```

## Step 2: Update main.go

Add route registration in your `main.go` file, similar to how incident is configured.

**File**: `cmd/main/main.go`

```go
package main

import (
	"context"
	"dev.azure.com/PHARMACITYJSC/PharmacityLibrary/_git/pmc-inapp-helper.git/v2/client/incident"
	"dev.azure.com/PHARMACITYJSC/PharmacityLibrary/_git/pmc-inapp-helper.git/v2/security"  // ADD THIS
	"dev.azure.com/PHARMACITYJSC/PharmacityLibrary/_git/pmc-inapp-helper.git/v2/server"
	"dev.azure.com/PHARMACITYJSC/PharmacityLibrary/_git/pmc-inapp-helper.git/v2/server/http/response"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/service"
	"pmc-inapp-outbound-ticket-svc/internal/config"
	_ "pmc-inapp-outbound-ticket-svc/internal/encoding/excel"
	"pmc-inapp-outbound-ticket-svc/internal/handler"
	"pmc-inapp-outbound-ticket-svc/internal/registry"
)

var configFile = flag.String("f", "etc/app.yaml", "the config file")

func main() {
	c := config.Load(configFile)
	configureGlobalHandlers(c)

	// Create service context
	svcCtx := registry.NewServiceContext(c)

	// Create REST handler
	restHandler := handler.NewRestHandler(svcCtx)

	// Register routes with permission service at startup
	// ADD THIS ↓
	security.RegisterRoutesAtStartup(
		restHandler,
		c.Security.ServiceCode,
		c.Security.RegistrationUrl,
		nil, // use default naming: "METHOD /path"
	)

	svcGroup := service.NewServiceGroup()
	svcGroup.Add(server.NewGrpcServer(c.Server,
		handler.NewGrpcHandler(svcCtx),
	))
	svcGroup.Add(server.NewHttpServer(c.Server,
		restHandler,  // Use the handler we already created
	))
	defer svcGroup.Stop()

	fmt.Printf("Starting server at %s:%d...\n", c.Server.Http.Host, c.Server.Http.Port)
	svcGroup.Start()
}

func configureGlobalHandlers(c config.Config) {
	response.SetIncidentHandler(func(ctx context.Context, err error) {
		_ = incident.New(c.GetInternal()).NotifyError(ctx, err)
	})
}
```

## Step 3: Ensure Config Has Registration URL

Make sure your `etc/app.yaml` has the security configuration:

```yaml
security:
  service_code: "pmc-inapp-outbound-ticket-svc"
  registration_url: "http://pmc-inapp-permission-svc:8000/api/v1/endpoints/register"
  permission_url: "http://pmc-inapp-permission-svc:8000"
  permission_cache_ttl: 6
  permission_stale_ttl: 24
```

## Custom Naming (Optional)

If you want custom names for your endpoints, provide a naming function:

```go
security.RegisterRoutesAtStartup(
	restHandler,
	c.Security.ServiceCode,
	c.Security.RegistrationUrl,
	func(method, path string) string {
		// Custom naming logic
		return fmt.Sprintf("[%s] %s", method, path)
	},
)
```

## How It Works

1. **Startup**: When service starts, `RegisterRoutesAtStartup()` is called
2. **Route Collection**: Calls `restHandler.Routes()` to get all routes
3. **Conversion**: Converts `[]rest.Route` to `[]security.Endpoint` with auto-hashing
4. **Registration**: Sends all endpoints to permission service

**Example Request to Permission Service:**
```json
POST /api/v1/endpoints/register

{
  "service_code": "pmc-inapp-outbound-ticket-svc",
  "endpoints": [
    {
      "method": "GET",
      "path": "/api/v1/outbound-tickets",
      "hash": "f8e7d6c5b4a39210a8c7e2f1d9b3c4e5",
      "name": "GET /api/v1/outbound-tickets"
    },
    {
      "method": "POST",
      "path": "/api/v1/outbound-tickets",
      "hash": "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6",
      "name": "POST /api/v1/outbound-tickets"
    }
  ]
}
```

## Notes

- Route registration is **non-blocking** - if it fails, service still starts
- Duplicate checking is handled by permission service
- Routes are identified by hash of `METHOD:PATH`
- Registration happens once at startup
- All routes from protobuf-generated handlers are automatically included
