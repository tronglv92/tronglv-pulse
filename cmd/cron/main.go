package main

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"pulse/internal/config"
	"pulse/internal/registry"
)

var configFile = flag.String("f", "etc/cron.yaml", "the config file")

func main() {
	flag.Parse()
	_ = godotenv.Load()

	c := config.Load[config.CronConfig](configFile)
	fmt.Printf("Pulse Cron starting (name: %s, budget: $%.2f/day)\n",
		c.Name, c.LLM.DailyBudgetUsd)

	_ = registry.NewCronContext(c) // TODO TASK-038: register cron commands
}
