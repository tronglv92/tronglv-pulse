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

var configFile = flag.String("f", "etc/ingest.yaml", "the config file")

func main() {
	flag.Parse()
	_ = godotenv.Load()

	c := config.Load[config.IngestConfig](configFile)
	fmt.Printf("Pulse Ingest server starting (name: %s, http: %s:%d)\n",
		c.Name, c.Host, c.Port)

	svcGroup := service.NewServiceGroup()
	defer svcGroup.Stop()

	svcGroup.Add(handler.NewHealthServer(c.Name, c.Host, c.Port))

	_ = registry.NewIngestContext(c) // TODO TASK-032: wire ingestion HTTP server

	svcGroup.Start()
}
