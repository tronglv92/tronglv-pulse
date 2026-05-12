package interceptor

import (
	"context"
	"pulse/helper/utils/authenticator"
	"pulse/helper/utils/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

type Authentication struct {
	enrich        bool
	authenticator authenticator.Authenticator
}

func NewAuthentication(authenticator authenticator.Authenticator, enrich bool) *Authentication {
	return &Authentication{
		enrich:        enrich,
		authenticator: authenticator,
	}
}

func (s *Authentication) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		authCtx, err := s.validate(ctx)
		if err != nil {
			return nil, err
		}
		return handler(authCtx, req)
	}
}

func (s *Authentication) Stream() grpc.StreamServerInterceptor {
	return func(svr any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		_, err := s.validate(stream.Context())
		if err != nil {
			return err
		}
		return handler(svr, stream)
	}
}

func (s *Authentication) validate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "Metadata not found")
	}

	token := md.Get("authorization")
	if len(token) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "Authorization token is missing")
	}

	tokenType, tokenData, err := authenticator.ExtractAuthToken(token[0])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Authorization token is invalid")
	}

	var mapClaims identity.Claims
	switch strings.ToUpper(tokenType) {
	case "BASIC":
		mapClaims, err = s.authenticator.SignInWithBasic(ctx, tokenData)
	case "BEARER":
		if s.enrich {
			mapClaims, err = s.authenticator.VerifyTokenWithPermissions(ctx, tokenData)
		} else {
			mapClaims, err = s.authenticator.VerifyToken(ctx, tokenData)
		}
	default:
		return nil, status.Errorf(codes.Unauthenticated, "Unsupported token type")
	}
	if err != nil {
		return nil, err
	}

	return identity.WithContext(ctx, mapClaims), nil
}
