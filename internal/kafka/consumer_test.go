package kafka

import (
	"testing"

	"pulse/helper/utils/queue/consumer"
	kafkahelper "pulse/helper/utils/queue/kafka"
	"pulse/internal/config"
)

// TestConsumeHandlerIsAlias verifies that ConsumeHandler is a true type alias
// for consumer.MessageHandler — not a separate interface.
func TestConsumeHandlerIsAlias(t *testing.T) {
	// Compile-time assertion: a consumer.MessageHandler value must be
	// assignable to ConsumeHandler without conversion.
	var _ ConsumeHandler = (consumer.MessageHandler)(nil)
}

func TestToHelperConfig(t *testing.T) {
	cfg := config.KafkaConfig{
		Brokers: []string{"broker-1:9092", "broker-2:9092"},
	}
	topic := "pulse.logs.raw"
	group := "pulse-enricher"

	got := toHelperConfig(cfg, topic, group)

	// Verify mapped fields.
	if len(got.Brokers) != 2 || got.Brokers[0] != "broker-1:9092" || got.Brokers[1] != "broker-2:9092" {
		t.Fatalf("Brokers mismatch: got %v", got.Brokers)
	}
	if got.Topic != topic {
		t.Fatalf("Topic: got %q, want %q", got.Topic, topic)
	}
	if got.Group != group {
		t.Fatalf("Group: got %q, want %q", got.Group, group)
	}

	// Verify unmapped fields stay at zero/default values.
	want := kafkahelper.Config{}
	if got.Offset != want.Offset {
		t.Fatalf("Offset should be zero value, got %q", got.Offset)
	}
	if got.Consumers != want.Consumers {
		t.Fatalf("Consumers should be zero value, got %d", got.Consumers)
	}
	if got.Processors != want.Processors {
		t.Fatalf("Processors should be zero value, got %d", got.Processors)
	}
}
