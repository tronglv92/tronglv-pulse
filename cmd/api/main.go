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

	svcGroup := service.NewServiceGroup()
	defer svcGroup.Stop()

	svcGroup.Add(handler.NewHealthServer(c.Name, c.Host, c.Port))

	_ = registry.NewServiceContext(c) // TODO TASK-023: wire HTTP + gRPC handlers

	svcGroup.Start()
}
