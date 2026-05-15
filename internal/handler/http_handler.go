package handler

import (
	"pulse/internal/auth"
	"pulse/internal/registry"

	"github.com/zeromicro/go-zero/rest"
)

// RestHandler registers all domain REST routes on the go-zero server.
type RestHandler struct {
	svcCtx registry.ServiceContext
}

// NewRestHandler creates a RestHandler from the root DI context.
func NewRestHandler(svcCtx registry.ServiceContext) *RestHandler {
	return &RestHandler{svcCtx: svcCtx}
}

// Register wires domain handlers and adds their routes to the server.
func (h *RestHandler) Register(srv *rest.Server) {
	authSvc := auth.NewAuthService(
		h.svcCtx.GetUserRepo(),
		h.svcCtx.GetTenantRepo(),
		h.svcCtx.GetAPIKeyRepo(),
		h.svcCtx.GetAuditLogRepo(),
		h.svcCtx.GetConfig().Auth.JwtSecret,
	)

	authHandler := auth.NewHandler(authSvc)
	authMw := h.svcCtx.GetAuthMiddleware()
	tenantMw := h.svcCtx.GetTenantMiddleware()

	srv.AddRoutes(authHandler.Routes(authMw, tenantMw))
}
