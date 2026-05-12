package server

import (
	"pulse/helper/utils/localize"
	"pulse/helper/utils/recovery"
	"pulse/helper/utils/server/grpc/interceptor"
	"pulse/helper/utils/server/http/middleware"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func init() {
	for _, v := range Providers() {
		v.Register()
	}
}

func NewHttpServer(c Config, h RestHandler, opts ...rest.RunOption) *rest.Server {
	srv := rest.MustNewServer(c.GetHttp(), opts...)
	srv.Use(recovery.RestMiddleware(c.GetEnvironment()))
	srv.Use(middleware.TraceMiddleware(c.GetEnvironment(), c.GetName()))
	srv.Use(localize.RestMiddleware(localize.NewWithDefault()))
	h.Register(srv)
	return srv
}

func NewGrpcServer(c Config, h GrpcHandler, opts ...grpc.ServerOption) *zrpc.RpcServer {
	s := zrpc.MustNewServer(c.GetGrpc(), func(grpcServer *grpc.Server) {
		h.Register(grpcServer)
		if h.Reflection() {
			reflection.Register(grpcServer)
		}
	})
	s.AddOptions(opts...)
	s.AddUnaryInterceptors(
		interceptor.UnaryLogInterceptor(c.GetEnvironment()),
		interceptor.UnaryRecoverInterceptor(c.GetEnvironment()),
	)
	h.Interceptors(s)
	return s
}
