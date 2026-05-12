package queue

import (
	"pulse/helper/utils/queue/kafka"
	"pulse/helper/utils/queue/rabbitmq"
)

func (c Config) GetQueueStack() string {
	return c.Stack
}

func (c Config) IsQueueStack(stack string) bool {
	return stack == c.GetQueueStack()
}

func (c Config) GetKafka() kafka.Config {
	return c.Kafka
}

func (c Config) GetRabbit() rabbitmq.Config {
	return c.Rabbit
}
