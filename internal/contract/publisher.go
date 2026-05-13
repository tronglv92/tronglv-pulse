package contract

import "context"

// KafkaPublisher publishes a single message to a Kafka topic.
// value is marshaled to JSON by the concrete implementation;
// key is the partition key (e.g., fingerprint or tenantID as string).
type KafkaPublisher interface {
	Publish(ctx context.Context, topic, key string, value any) error
}
