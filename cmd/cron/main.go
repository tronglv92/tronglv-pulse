package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/joho/godotenv"
)

var configFile = flag.String("f", "etc/cron.yaml", "the config file")

func main() {
	flag.Parse()
	_ = godotenv.Load()
	_ = context.Background()

	fmt.Printf("Pulse Cron service starting (config: %s)...\n", *configFile)

	// TODO TASK-007: c := config.Load(configFile)
	// TODO TASK-009: ctx := registry.NewCronContext(c)
	// TODO TASK-037: cronSvc := server.NewCron(c.Server.GetCronJob())
	// TODO TASK-037: cronSvc.Register(context.Background(), cron.RegisterCommands(ctx))
}
