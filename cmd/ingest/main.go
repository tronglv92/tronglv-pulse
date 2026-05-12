package main

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/ingest.yaml", "the config file")

func main() {
	flag.Parse()
	_ = godotenv.Load()

	fmt.Printf("Pulse Ingest server starting (config: %s)...\n", *configFile)

	svcGroup := service.NewServiceGroup()
	defer svcGroup.Stop()

	// TODO TASK-007: load config via internal/config.Load(configFile)
	// TODO TASK-009: svcCtx := registry.NewServiceContext(c)
	// TODO TASK-013: svcGroup.Add(server.NewHttpServer(c.Server, handler.NewIngestHandler(svcCtx)))

	svcGroup.Start()
}
