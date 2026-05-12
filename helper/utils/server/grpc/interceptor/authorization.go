package interceptor

import (
	"context"
	"pulse/helper/utils/identity"
	pmcpb "pulse/helper/utils/pmcpb/protobuf"
	"pulse/helper/utils/toolkit/stringx"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"strings"
)

type (
	Authorization struct {
		ignoreEmptyCode bool
	}

	Option func(*Authorization)
)

func WithIgnoreEmptyCode(ignoreEmptyCode bool) Option {
	return func(s *Authorization) {
		s.ignoreEmptyCode = ignoreEmptyCode
	}
}

func NewAuthorization(opts ...Option) *Authorization {
	r := &Authorization{
		ignoreEmptyCode: true,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (s *Authorization) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := s.validate(ctx, info.FullMethod); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (s *Authorization) Stream() grpc.StreamServerInterceptor {
	return func(svr any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := s.validate(stream.Context(), info.FullMethod); err != nil {
			return err
		}
		return handler(svr, stream)
	}
}

func (s *Authorization) validate(ctx context.Context, fullMethod string) error {
	parts := strings.Split(fullMethod, "/")
	if len(parts) != 3 {
		return status.Error(codes.NotFound, fmt.Sprintf("Invalid method parts: %s", fullMethod))
	}
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(parts[1]))
	if err != nil {
		return err
	}

	methodDesc := desc.(protoreflect.ServiceDescriptor).Methods().ByName(protoreflect.Name(parts[2]))
	if methodDesc == nil {
		return status.Error(codes.NotFound, fmt.Sprintf("Invalid method: %s", fullMethod))
	}

	methodOptions := methodDesc.Options().(*descriptorpb.MethodOptions)
	permissionCode := proto.GetExtension(methodOptions, pmcpb.E_PermissionCode)
	if stringx.IsNilOrEmptyString(permissionCode) {
		if s.ignoreEmptyCode {
			return nil
		}
		return status.Error(codes.NotFound, fmt.Sprintf("Permission code not found for method: %s", fullMethod))
	}
	
	jwtMap, err := identity.MapClaimsFromContext(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	if !jwtMap.HasPermission(permissionCode.(string)) {
		return status.Errorf(codes.PermissionDenied, "Missing permission")
	}
	return nil
}
