package server

import (
	"pulse/helper/utils/security"
	"pulse/helper/utils/server/core"
	"pulse/helper/utils/toolkit/filex"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/gateway"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type RestHandler interface {
	Register(svr *rest.Server)
}

type GrpcHandler interface {
	Register(svr *grpc.Server)
	Interceptors(rpc *zrpc.RpcServer)
	Reflection() bool
}

func Providers() []core.Service {
	return []core.Service{
		filex.NewMimeType(),
	}
}

type Config struct {
	Env      string              `json:",default=pro,optional"`
	Language string              `json:"lang,default=vi"`
	Cron     CronJobConf         `json:"cronjob,optional"`
	Http     rest.RestConf       `json:"http,optional"`
	Grpc     zrpc.RpcServerConf  `json:"grpc,optional"`
	Gateway  gateway.GatewayConf `json:"gateway,optional"`
	Security security.Config     `json:"security,optional"`
}

func (c Config) GetEnvironment() string {
	return c.Env
}

func (c Config) IsProduction() bool {
	return c.IsEnvironment(service.ProMode)
}

func (c Config) IsEnvironment(env string) bool {
	return c.GetEnvironment() == env
}

func (c Config) GetLang() string {
	return c.Language
}

func (c Config) GetName() string {
	if len(c.Http.Name) > 0 {
		return c.Http.Name
	}
	return c.Grpc.Name
}

func (c Config) GetCronJob() CronJobConf {
	return c.Cron
}

func (c Config) GetHttp() rest.RestConf {
	return c.Http
}

func (c Config) GetGrpc() zrpc.RpcServerConf {
	return c.Grpc
}

func (c Config) GetGateway() gateway.GatewayConf {
	return c.Gateway
}

func (c Config) GetSecurity() security.Config { return c.Security }
