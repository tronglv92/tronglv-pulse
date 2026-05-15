package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"pulse/helper/utils/errors"
	"pulse/helper/utils/identity"
	"pulse/internal/contract"
	"pulse/internal/types/entity"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 30 * 24 * time.Hour
	apiKeyPrefix         = "pk_"
	apiKeyRandBytes      = 24 // 24 bytes = 48 hex chars → "pk_<48hex>"
)

// AuthService handles authentication logic: login, token refresh, and API key CRUD.
type AuthService struct {
	userRepo   contract.UserRepo
	tenantRepo contract.TenantRepo
	apiKeyRepo contract.APIKeyRepo
	auditRepo  contract.AuditLogRepo
	jwtSecret  string
}

// NewAuthService creates a new AuthService with all required dependencies.
func NewAuthService(
	userRepo contract.UserRepo,
	tenantRepo contract.TenantRepo,
	apiKeyRepo contract.APIKeyRepo,
	auditRepo contract.AuditLogRepo,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		apiKeyRepo: apiKeyRepo,
		auditRepo:  auditRepo,
		jwtSecret:  jwtSecret,
	}
}

// Login authenticates a user by email/password and returns a JWT token pair.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.NewBadRequest("INVALID_INPUT", "Email and password are required.")
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.NewUnauthorized("INVALID_CREDENTIALS", "Invalid email or password.")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.NewUnauthorized("INVALID_CREDENTIALS", "Invalid email or password.")
	}

	if !user.IsActive {
		return nil, errors.NewUnauthorized("ACCOUNT_DISABLED", "Account is disabled.")
	}

	tenant, err := s.tenantRepo.FindByID(ctx, user.TenantID)
	if err != nil {
		return nil, errors.NewInternalServer("TENANT_NOT_FOUND", "Associated tenant not found.")
	}

	tokens, err := s.generateTokenPair(user, tenant)
	if err != nil {
		return nil, errors.NewInternalServer("TOKEN_GENERATION_FAILED", "Failed to generate tokens.")
	}

	// Best-effort audit log — don't fail login on audit write error.
	_ = s.auditRepo.Append(ctx, &entity.AuditLog{
		TenantID:  user.TenantID,
		ActorID:   &user.ID,
		ActorType: "user",
		Action:    "auth.login",
		Resource:  "user",
		ResourceID: &user.ID,
	})

	return tokens, nil
}

// Refresh validates a refresh token and returns a new token pair.
func (s *AuthService) Refresh(ctx context.Context, req RefreshRequest) (*TokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, errors.NewBadRequest("INVALID_INPUT", "Refresh token is required.")
	}

	claims, err := identity.FromToken(req.RefreshToken, nil, []byte(s.jwtSecret))
	if err != nil {
		return nil, errors.NewUnauthorized("INVALID_TOKEN", "Invalid or expired refresh token.")
	}

	if claims.GetKind() != "refresh" {
		return nil, errors.NewUnauthorized("INVALID_TOKEN_KIND", "Token is not a refresh token.")
	}

	userID, err := strconv.ParseInt(claims.GetId(), 10, 64)
	if err != nil {
		return nil, errors.NewUnauthorized("INVALID_TOKEN", "Invalid token subject.")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.NewUnauthorized("USER_NOT_FOUND", "User not found.")
	}

	if !user.IsActive {
		return nil, errors.NewUnauthorized("ACCOUNT_DISABLED", "Account is disabled.")
	}

	tenant, err := s.tenantRepo.FindByID(ctx, user.TenantID)
	if err != nil {
		return nil, errors.NewInternalServer("TENANT_NOT_FOUND", "Associated tenant not found.")
	}

	tokens, err := s.generateTokenPair(user, tenant)
	if err != nil {
		return nil, errors.NewInternalServer("TOKEN_GENERATION_FAILED", "Failed to generate tokens.")
	}

	return tokens, nil
}

// CreateAPIKey generates a new API key for the tenant, stores the SHA-256 hash,
// and returns the raw key (shown only once).
func (s *AuthService) CreateAPIKey(ctx context.Context, tenantID, actorID int64, req CreateAPIKeyRequest) (*APIKeyResponse, error) {
	if req.Name == "" {
		return nil, errors.NewBadRequest("INVALID_INPUT", "API key name is required.")
	}

	rawBytes := make([]byte, apiKeyRandBytes)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	rawKey := apiKeyPrefix + hex.EncodeToString(rawBytes)

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	apiKey := &entity.APIKey{
		TenantID:  tenantID,
		KeyHash:   keyHash,
		Name:      req.Name,
		ExpiresAt: req.ExpiresAt,
	}

	if err := s.apiKeyRepo.Save(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to save api key: %w", err)
	}

	_ = s.auditRepo.Append(ctx, &entity.AuditLog{
		TenantID:   tenantID,
		ActorID:    &actorID,
		ActorType:  "user",
		Action:     "api_key.create",
		Resource:   "api_key",
		ResourceID: &apiKey.ID,
	})

	return &APIKeyResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		RawKey:    rawKey,
		CreatedAt: apiKey.CreatedAt,
		ExpiresAt: apiKey.ExpiresAt,
	}, nil
}

// RevokeAPIKey soft-deletes an API key by ID, scoped to the tenant.
func (s *AuthService) RevokeAPIKey(ctx context.Context, id, tenantID, actorID int64) error {
	if err := s.apiKeyRepo.Delete(ctx, id, tenantID); err != nil {
		return errors.NewNotFound("API_KEY_NOT_FOUND", "API key not found.")
	}

	_ = s.auditRepo.Append(ctx, &entity.AuditLog{
		TenantID:   tenantID,
		ActorID:    &actorID,
		ActorType:  "user",
		Action:     "api_key.revoke",
		Resource:   "api_key",
		ResourceID: &id,
	})

	return nil
}

// ListAPIKeys returns all active API keys for the tenant (never exposing the hash).
func (s *AuthService) ListAPIKeys(ctx context.Context, tenantID int64) (*APIKeyListResponse, error) {
	keys, err := s.apiKeyRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}

	items := make([]APIKeyItem, len(keys))
	for i, k := range keys {
		items[i] = APIKeyItem{
			ID:         k.ID,
			Name:       k.Name,
			CreatedAt:  k.CreatedAt,
			ExpiresAt:  k.ExpiresAt,
			LastUsedAt: k.LastUsedAt,
		}
	}

	return &APIKeyListResponse{Keys: items}, nil
}

// generateTokenPair creates HS256-signed access and refresh JWTs.
func (s *AuthService) generateTokenPair(user *entity.User, tenant *entity.Tenant) (*TokenResponse, error) {
	now := time.Now()

	accessToken, err := s.signToken(user, tenant, "user", now, accessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken, err := s.signToken(user, tenant, "refresh", now, refreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenDuration.Seconds()),
	}, nil
}

func (s *AuthService) signToken(user *entity.User, tenant *entity.Tenant, kind string, now time.Time, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"eid":  strconv.FormatInt(user.ID, 10),
		"kind": kind,
		"iss":  "pulse",
		"iat":  now.Unix(),
		"exp":  now.Add(duration).Unix(),
		"attributes": map[string]string{
			"tenant_id": strconv.FormatInt(tenant.ID, 10),
			"email":     user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
