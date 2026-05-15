package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pulse/helper/utils/queue/consumer"
	kafkahelper "pulse/helper/utils/queue/kafka"
	"pulse/internal/config"

	segkafka "github.com/segmentio/kafka-go"
)

// Publisher implements contract.KafkaPublisher by wrapping the helper kafka producer.
// It marshals values to JSON and publishes via the helper's SendCtx method.
type Publisher struct {
	producer *kafkahelper.Producer
}

// NewPublisher creates a Publisher with acks=all, batching defaults, and auto topic creation.
// Additional opts override or extend the defaults.
func NewPublisher(brokers []string, opts ...kafkahelper.PushOption) *Publisher {
	defaults := []kafkahelper.PushOption{
		kafkahelper.WithRequiredAck(segkafka.RequireAll),
		kafkahelper.WithBatchSize(100),
		kafkahelper.WithBatchTimeout(50 * time.Millisecond),
		kafkahelper.WithAllowAutoTopicCreation(),
	}
	combined := append(defaults, opts...)

	cfg := kafkahelper.Config{
		Brokers: brokers,
	}

	return &Publisher{
		producer: kafkahelper.NewProducer(cfg, combined...),
	}
}

// NewPublisherFromConfig creates a Publisher from a KafkaConfig.
func NewPublisherFromConfig(c config.KafkaConfig, opts ...kafkahelper.PushOption) *Publisher {
	return NewPublisher(c.Brokers, opts...)
}

// Publish marshals value to JSON and sends it to the given Kafka topic with the specified key.
func (p *Publisher) Publish(ctx context.Context, topic, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message for %s: %w", topic, err)
	}

	payload := &consumer.Payload{
		QueueName: topic,
		Message:   json.RawMessage(data),
		MetaData:  map[string]string{kafkahelper.Key: key},
	}

	if err := p.producer.SendCtx(ctx, payload); err != nil {
		return fmt.Errorf("failed to publish to %s: %w", topic, err)
	}

	return nil
}

// Close shuts down the underlying Kafka writer.
func (p *Publisher) Close() error {
	return p.producer.Close()
}
