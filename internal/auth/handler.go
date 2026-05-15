package auth

import (
	"net/http"
	"strconv"
	"strings"

	"pulse/helper/utils/errors"
	"pulse/helper/utils/identity"
	"pulse/helper/utils/server/http/response"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Handler exposes auth REST endpoints.
type Handler struct {
	svc *AuthService
}

// NewHandler creates a new auth Handler.
func NewHandler(svc *AuthService) *Handler {
	return &Handler{svc: svc}
}

// Routes returns auth routes. Public routes have no middleware wrapper;
// protected routes are wrapped with authMw and tenantMw.
func (h *Handler) Routes(authMw, tenantMw rest.Middleware) []rest.Route {
	protect := func(handler http.HandlerFunc) http.HandlerFunc {
		return authMw(tenantMw(handler))
	}

	return []rest.Route{
		{Method: http.MethodPost, Path: "/v1/auth/login", Handler: h.login()},
		{Method: http.MethodPost, Path: "/v1/auth/refresh", Handler: h.refresh()},
		{Method: http.MethodPost, Path: "/v1/auth/keys", Handler: protect(h.createAPIKey())},
		{Method: http.MethodDelete, Path: "/v1/auth/keys/:id", Handler: protect(h.revokeAPIKey())},
		{Method: http.MethodGet, Path: "/v1/auth/keys", Handler: protect(h.listAPIKeys())},
	}
}

func (h *Handler) login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			response.Error(r.Context(), w, errors.NewBadRequest("INVALID_BODY", "Invalid request body."))
			return
		}

		result, err := h.svc.Login(r.Context(), req)
		if err != nil {
			response.Error(r.Context(), w, err)
			return
		}

		response.OkJson(r.Context(), w, result)
	}
}

func (h *Handler) refresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RefreshRequest
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			response.Error(r.Context(), w, errors.NewBadRequest("INVALID_BODY", "Invalid request body."))
			return
		}

		result, err := h.svc.Refresh(r.Context(), req)
		if err != nil {
			response.Error(r.Context(), w, err)
			return
		}

		response.OkJson(r.Context(), w, result)
	}
}

func (h *Handler) createAPIKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := TenantIDFromContext(r.Context())
		if !ok {
			response.Error(r.Context(), w, errors.NewUnauthorized("MISSING_TENANT", "Tenant context required."))
			return
		}

		claims := identity.FromContext(r.Context())
		actorID, _ := strconv.ParseInt(claims.GetId(), 10, 64)

		var req CreateAPIKeyRequest
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			response.Error(r.Context(), w, errors.NewBadRequest("INVALID_BODY", "Invalid request body."))
			return
		}

		result, err := h.svc.CreateAPIKey(r.Context(), tenantID, actorID, req)
		if err != nil {
			response.Error(r.Context(), w, err)
			return
		}

		response.OkJson(r.Context(), w, result)
	}
}

func (h *Handler) revokeAPIKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := TenantIDFromContext(r.Context())
		if !ok {
			response.Error(r.Context(), w, errors.NewUnauthorized("MISSING_TENANT", "Tenant context required."))
			return
		}

		claims := identity.FromContext(r.Context())
		actorID, _ := strconv.ParseInt(claims.GetId(), 10, 64)

		// Extract :id path param. go-zero puts path params in the URL path.
		// For "/v1/auth/keys/:id", extract the last segment.
		pathParts := strings.Split(r.URL.Path, "/")
		idStr := pathParts[len(pathParts)-1]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.Error(r.Context(), w, errors.NewBadRequest("INVALID_ID", "Invalid API key ID."))
			return
		}

		if err := h.svc.RevokeAPIKey(r.Context(), id, tenantID, actorID); err != nil {
			response.Error(r.Context(), w, err)
			return
		}

		response.OkMsg(r.Context(), w)
	}
}

func (h *Handler) listAPIKeys() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := TenantIDFromContext(r.Context())
		if !ok {
			response.Error(r.Context(), w, errors.NewUnauthorized("MISSING_TENANT", "Tenant context required."))
			return
		}

		result, err := h.svc.ListAPIKeys(r.Context(), tenantID)
		if err != nil {
			response.Error(r.Context(), w, err)
			return
		}

		response.OkJson(r.Context(), w, result)
	}
}
