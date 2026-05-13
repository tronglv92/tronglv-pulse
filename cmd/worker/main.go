package main

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/service"
	"pulse/internal/config"
	"pulse/internal/registry"
)

var configFile = flag.String("f", "etc/worker.yaml", "the config file")

func main() {
	flag.Parse()
	_ = godotenv.Load()

	c := config.Load[config.WorkerConfig](configFile)
	fmt.Printf("Pulse Worker starting (name: %s, brokers: %v)\n",
		c.Name, c.Kafka.Brokers)

	svcGroup := service.NewServiceGroup()
	defer svcGroup.Stop()

	_ = registry.NewConsumerContext(c) // TODO TASK-016: wire Kafka consumers + outbox drainer

	svcGroup.Start()
}
