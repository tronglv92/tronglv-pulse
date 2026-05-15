package main

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/service"

	"pulse/internal/config"
	"pulse/internal/consumer"
	"pulse/internal/kafka"
	"pulse/internal/outbox"
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

	ctx := registry.NewConsumerContext(c)

	kafkaPub := kafka.NewPublisherFromConfig(c.Kafka)
	dedup := outbox.NewDeduplicator(ctx.GetCache())
	pub := outbox.NewPublisher(ctx.GetOutboxRepo(), kafkaPub, dedup)

	sup := consumer.NewSupervisor()
	sup.Add(consumer.Registration{Name: "outbox-publisher", Listener: pub})
	// TODO: register domain consumers (enricher, rca, notify) via sup.Add(...)
	svcGroup.Add(sup)

	svcGroup.Start()
}
