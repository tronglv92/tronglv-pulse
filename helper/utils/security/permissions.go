package security

import (
	"context"
	"pulse/helper/utils/httpc"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/logx"
)

const RoleKey = "security:roles"

// RolePermissions represents the permissions for a single role.
type RolePermissions struct {
	Permissions []string `json:"permissions"`
}

// RolePermissionResponse represents the API response structure.
type RolePermissionResponse struct {
	Data struct {
		Roles map[string]RolePermissions `json:"roles"` // Role code -> permissions
	} `json:"data"`
}

type PermissionProvider interface {
	GetPermissions(ctx context.Context, userRoles []string) ([]string, error)
	Refresh(ctx context.Context) error
}

type permissionProvider struct {
	code        string
	baseUrl     string
	httpClient  httpc.Service
	cacheClient *collection.Cache
	staleCache  *StaleAwareCache
}

func NewPermissionProvider(code, baseUrl string, cacheClient, staleCacheClient *collection.Cache, httpClient httpc.Service) PermissionProvider {
	return &permissionProvider{
		code:        code,
		baseUrl:     baseUrl,
		cacheClient: cacheClient,
		httpClient:  httpClient,
		staleCache:  NewStaleAwareCache(cacheClient, staleCacheClient),
	}
}

func (p *permissionProvider) fetch(ctx context.Context) (map[string][]string, error) {
	url := fmt.Sprintf("%s/permission-svc/api/v1/services/roles?service_code=%s", p.baseUrl, p.code)

	// Use retry mechanism for resilience
	resp, err := httpc.WithRetry(ctx, func(ctx context.Context) (*http.Response, error) {
		return p.httpClient.Get(ctx, url)
	}, httpc.DefaultRetryConfig)

	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var apiResp RolePermissionResponse
	if err = json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	// Convert nested structure to flat map[string][]string
	roleMap := make(map[string][]string)
	for roleCode, rolePerms := range apiResp.Data.Roles {
		roleMap[roleCode] = rolePerms.Permissions
	}

	return roleMap, nil
}

func (p *permissionProvider) GetPermissions(ctx context.Context, userRoles []string) ([]string, error) {
	roles, isStale, err := p.staleCache.TakeWithStale(ctx, RoleKey, func(ctx context.Context) (any, error) {
		return p.fetch(ctx)
	})
	if err != nil {
		return nil, err
	}
	if isStale {
		logx.WithContext(ctx).Info("using stale permission cache")
	}

	roleMap := roles.(map[string][]string)

	permSet := make(map[string]struct{})
	for _, r := range userRoles {
		if ps, ok := roleMap[r]; ok {
			for _, v := range ps {
				permSet[v] = struct{}{}
			}
		}
	}

	perms := make([]string, 0, len(permSet))
	for v := range permSet {
		perms = append(perms, v)
	}
	return perms, nil
}

func (p *permissionProvider) Refresh(ctx context.Context) error {
	_, err := p.cacheClient.Take(RoleKey, func() (any, error) {
		return p.fetch(ctx)
	})
	return err
}
