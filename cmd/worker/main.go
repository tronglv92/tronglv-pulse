package main

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/worker.yaml", "the config file")

func main() {
	flag.Parse()
	_ = godotenv.Load()

	fmt.Printf("Pulse Worker starting (config: %s)...\n", *configFile)

	svcGroup := service.NewServiceGroup()
	defer svcGroup.Stop()

	// TODO TASK-007: load config via internal/config.Load(configFile)
	// TODO TASK-009: ctx := registry.NewConsumerContext(c)
	// TODO TASK-022: wire Kafka consumers (enricher, rca, notify) via consumer.NewHandler(ctx)
	// TODO TASK-040: svcGroup.Add(outbox.NewPublisher(ctx)) — outbox drain goroutine

	svcGroup.Start()
}
