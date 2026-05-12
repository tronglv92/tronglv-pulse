package interceptor

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
)

func UnaryLogInterceptor(env string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log := logx.WithContext(ctx)
		log.Infof("gRPC method called: %s, request: %v", info.FullMethod, req)
		resp, err := handler(ctx, req)
		if err != nil {
			log.Errorf("gRPC method failed: %s, error: %v", info.FullMethod, err)
		}
		return resp, err
	}
}
