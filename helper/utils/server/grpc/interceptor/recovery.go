package interceptor

import (
	"context"
	"pulse/helper/utils/errors"
	"pulse/helper/utils/recovery"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/service"
	"google.golang.org/grpc"
	"runtime/debug"
)

func UnaryRecoverInterceptor(env string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer handleCrash(func(r any) {
			err = toPanicError(ctx, env, r)
		})

		return handler(ctx, req)
	}
}

func handleCrash(handler func(any)) {
	if r := recover(); r != nil {
		handler(r)
	}
}

func toPanicError(ctx context.Context, env string, r any) error {
	logc.Errorf(ctx, "%+v\n\n%s", r, debug.Stack())

	var metadata = make(map[string]string)
	if env != service.ProMode {
		metadata["stack"] = recovery.SprintStack()
	}
	return errors.NewInternalServer("PANIC", "An unexpected server error occurred.").WithMetadata(metadata).GRPCStatus().Err()
}
