package authenticator

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"pulse/helper/utils/httpc"
	"pulse/helper/utils/identity"
	pb "pulse/helper/utils/internal/authenticator"
	"google.golang.org/grpc"
)

// AuthTransport is the gRPC transport interface for the authenticator service.
type AuthTransport interface {
	SignInWithService(ctx context.Context, in *pb.SignInWithServiceRequest, opts ...grpc.CallOption) (*pb.OAuthTokenResponse, error)
	VerifyToken(ctx context.Context, in *pb.VerifyTokenRequest, opts ...grpc.CallOption) (*pb.VerifyTokenResponse, error)
}

// Authenticator authenticates requests and verifies tokens.
type Authenticator interface {
	SignInWithBasic(ctx context.Context, credentials string) (identity.Claims, error)
	VerifyToken(ctx context.Context, token string) (identity.Claims, error)
	VerifyTokenWithPermissions(ctx context.Context, token string) (identity.Claims, error)
	SignInWithService(ctx context.Context, clientId, clientSecret string) (*pb.OAuthTokenResponse, error)
}

// TokenManager manages access tokens for service-to-service communication.
type TokenManager interface {
	GetHTTPClient(ctx context.Context, serviceCode string) (httpc.Service, error)
}

type defaultAuthenticator struct {
	transport AuthTransport
}

// NewAuthenticator creates a new Authenticator using the given gRPC transport.
func NewAuthenticator(transport AuthTransport) Authenticator {
	return &defaultAuthenticator{transport: transport}
}

func (a *defaultAuthenticator) SignInWithBasic(ctx context.Context, credentials string) (identity.Claims, error) {
	decoded, err := base64.StdEncoding.DecodeString(credentials)
	if err != nil {
		return nil, fmt.Errorf("invalid basic auth credentials: %w", err)
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid basic auth format, expected clientId:clientSecret")
	}
	resp, err := a.transport.SignInWithService(ctx, &pb.SignInWithServiceRequest{
		ClientId:     parts[0],
		ClientSecret: parts[1],
	})
	if err != nil {
		return nil, fmt.Errorf("sign in with basic: %w", err)
	}
	return a.VerifyToken(ctx, resp.GetAccessToken())
}

func (a *defaultAuthenticator) VerifyToken(ctx context.Context, token string) (identity.Claims, error) {
	resp, err := a.transport.VerifyToken(ctx, &pb.VerifyTokenRequest{Token: token})
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}
	return resp, nil
}

func (a *defaultAuthenticator) VerifyTokenWithPermissions(ctx context.Context, token string) (identity.Claims, error) {
	resp, err := a.transport.VerifyToken(ctx, &pb.VerifyTokenRequest{Token: token, Enrich: true})
	if err != nil {
		return nil, fmt.Errorf("verify token with permissions: %w", err)
	}
	return resp, nil
}

func (a *defaultAuthenticator) SignInWithService(ctx context.Context, clientId, clientSecret string) (*pb.OAuthTokenResponse, error) {
	return a.transport.SignInWithService(ctx, &pb.SignInWithServiceRequest{
		ClientId:     clientId,
		ClientSecret: clientSecret,
	})
}

// ExtractAuthToken parses an Authorization header value into token type and data.
// Supports "Bearer <token>" and "Basic <credentials>" formats.
func ExtractAuthToken(header string) (tokenType, tokenData string, err error) {
	if header == "" {
		return "", "", fmt.Errorf("authorization header is empty")
	}
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "", "", fmt.Errorf("invalid authorization header format, expected '<type> <token>'")
	}
	return parts[0], parts[1], nil
}
