package main

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/service"
	"pulse/internal/config"
	"pulse/internal/handler"
	"pulse/internal/registry"
)

var configFile = flag.String("f", "etc/api.yaml", "the config file")

func main() {
	flag.Parse()
	_ = godotenv.Load()

	c := config.Load[config.APIConfig](configFile)
	fmt.Printf("Pulse API server starting (name: %s, http: %s:%d, grpc: %s)\n",
		c.Name, c.Host, c.Port, c.Grpc.ListenOn)

	svcCtx := registry.NewServiceContext(c)

	srv := handler.NewHealthServer(c.Name, c.Host, c.Port)
	handler.NewRestHandler(svcCtx).Register(srv)

	svcGroup := service.NewServiceGroup()
	defer svcGroup.Stop()

	svcGroup.Add(srv)
	svcGroup.Start()
}
