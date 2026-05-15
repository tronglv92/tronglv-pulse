package kafka

import (
	"fmt"

	"pulse/helper/utils/queue/consumer"
	kafkahelper "pulse/helper/utils/queue/kafka"
	"pulse/internal/config"

	"github.com/zeromicro/go-zero/core/queue"
)

// ConsumeHandler is a type alias for consumer.MessageHandler.
// Any consumer.MessageHandler is automatically a ConsumeHandler — no adapter needed.
type ConsumeHandler = consumer.MessageHandler

// NewListener creates a Kafka consumer queue for the given topic and consumer group.
// It returns the queue and any error from the helper layer.
func NewListener(cfg config.KafkaConfig, topic, group string,
	handler ConsumeHandler, opts ...kafkahelper.QueueOption,
) (queue.MessageQueue, error) {
	helperCfg := toHelperConfig(cfg, topic, group)
	q, err := kafkahelper.NewQueue(helperCfg, handler, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka listener for %s/%s: %w", topic, group, err)
	}
	return q, nil
}

// MustNewListener is like NewListener but panics on error.
func MustNewListener(cfg config.KafkaConfig, topic, group string,
	handler ConsumeHandler, opts ...kafkahelper.QueueOption,
) queue.MessageQueue {
	q, err := NewListener(cfg, topic, group, handler, opts...)
	if err != nil {
		panic(err)
	}
	return q
}

// toHelperConfig maps domain KafkaConfig + topic/group into the helper's Config.
// All other fields (Offset, Consumers, Processors, etc.) use the helper's defaults.
// Callers customize via QueueOption functions.
func toHelperConfig(cfg config.KafkaConfig, topic, group string) kafkahelper.Config {
	return kafkahelper.Config{
		Brokers: cfg.Brokers,
		Topic:   topic,
		Group:   group,
	}
}
