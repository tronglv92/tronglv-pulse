package queue

import (
	"context"
	"pulse/helper/utils/queue/consumer"
	"pulse/helper/utils/queue/kafka"
	"pulse/helper/utils/queue/rabbitmq"
	"fmt"
)

const (
	RabbitDriver string = "rabbitmq"
	KafkaDriver  string = "kafka"
)

type (
	Producer interface {
		Send(payload consumer.MessageContext) error
		SendCtx(ctx context.Context, payload consumer.MessageContext) error
		Close() error
	}

	Config struct {
		Stack  string          `json:"stack,optional"`
		Rabbit rabbitmq.Config `json:"rabbitmq,optional"`
		Kafka  kafka.Config    `json:"kafka,optional"`
	}

	options struct {
		KafkaOptions []kafka.PushOption `json:"kafka,optional"`
	}

	Option func(options *options)
)

func New(c Config, opts ...Option) (Producer, error) {
	var t options
	for _, opt := range opts {
		opt(&t)
	}
	switch c.Stack {
	case RabbitDriver:
		return rabbitmq.NewProducer(c.Rabbit), nil
	case KafkaDriver:
		return kafka.NewProducer(c.Kafka, t.KafkaOptions...), nil
	}
	return nil, fmt.Errorf("the queue driver does not support")
}

func Must(cfg Config, opts ...Option) Producer {
	q, err := New(cfg, opts...)
	if err != nil {
		panic(err)
	}
	return q
}

func WithKafkaOptions(opts ...kafka.PushOption) Option {
	return func(options *options) {
		options.KafkaOptions = opts
	}
}
