package security

type Config struct {
	JwksUrl            string   `json:"jwks_url,default=http://pmc-inapp-authentication-svc:8000/.well-known/jwks.json"`
	JWKSCacheTtl       int      `json:"jwks_cache_ttl,default=24"` //hours
	PermissionUrl      string   `json:"permission_url,default=http://pmc-inapp-permission-svc:8000"`
	PermissionCacheTtl int      `json:"permission_cache_ttl,default=6"`  //hours
	PermissionStaleTtl int      `json:"permission_stale_ttl,default=24"` //hours - stale cache fallback TTL
	RegistrationUrl    string   `json:"registration_url,optional"`       //route registration endpoint
	ServiceCode        string   `json:"service_code"`
	AllowedAudience    []string `json:"allowed_audience,optional"`
}
