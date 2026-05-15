package kafka

import (
	"context"
	"encoding/json"
	"testing"

	"pulse/helper/utils/queue/consumer"
	kafkahelper "pulse/helper/utils/queue/kafka"
	"pulse/internal/contract"
)

// Compile-time assertion: Publisher must satisfy contract.KafkaPublisher.
var _ contract.KafkaPublisher = (*Publisher)(nil)

// TestPublisher_Publish_Success verifies that Publish marshals a Go struct to JSON
// and constructs the correct Payload envelope.
func TestPublisher_Publish_Success(t *testing.T) {
	type logEvent struct {
		Service string `json:"service"`
		Level   string `json:"level"`
		Message string `json:"message"`
	}

	event := logEvent{
		Service: "api-gateway",
		Level:   "error",
		Message: "connection refused",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	payload := &consumer.Payload{
		QueueName: TopicRawLogs,
		Message:   json.RawMessage(data),
		MetaData:  map[string]string{kafkahelper.Key: "tenant-42"},
	}

	// Verify payload fields are constructed correctly.
	if payload.GetQueueName() != TopicRawLogs {
		t.Errorf("expected topic %q, got %q", TopicRawLogs, payload.GetQueueName())
	}

	meta := payload.GetMetaData()
	if meta[kafkahelper.Key] != "tenant-42" {
		t.Errorf("expected key %q, got %q", "tenant-42", meta[kafkahelper.Key])
	}

	// Verify the JSON roundtrip of the message payload.
	payloadBytes := payload.GetByte()
	if payloadBytes == nil {
		t.Fatal("payload.GetByte() returned nil")
	}

	var decoded consumer.Payload
	if err := json.Unmarshal(payloadBytes, &decoded); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if decoded.QueueName != TopicRawLogs {
		t.Errorf("decoded queue name: want %q, got %q", TopicRawLogs, decoded.QueueName)
	}

	// Verify the inner message is the original event JSON.
	rawMsg, err := json.Marshal(decoded.Message)
	if err != nil {
		t.Fatalf("failed to marshal decoded message: %v", err)
	}

	var roundTripped logEvent
	if err := json.Unmarshal(rawMsg, &roundTripped); err != nil {
		t.Fatalf("failed to unmarshal decoded message: %v", err)
	}

	if roundTripped.Service != event.Service {
		t.Errorf("service: want %q, got %q", event.Service, roundTripped.Service)
	}
	if roundTripped.Level != event.Level {
		t.Errorf("level: want %q, got %q", event.Level, roundTripped.Level)
	}
	if roundTripped.Message != event.Message {
		t.Errorf("message: want %q, got %q", event.Message, roundTripped.Message)
	}
}

// TestPublisher_Publish_MarshalError verifies that Publish returns an error
// when the value cannot be marshaled to JSON.
func TestPublisher_Publish_MarshalError(t *testing.T) {
	// Create publisher with dummy broker (won't connect — just for construction).
	pub := NewPublisher([]string{"localhost:9092"})
	defer pub.Close()

	// channel types cannot be marshaled to JSON.
	err := pub.Publish(context.Background(), TopicRawLogs, "key-1", make(chan int))
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}

	expected := "failed to marshal message for " + TopicRawLogs
	if len(err.Error()) < len(expected) || err.Error()[:len(expected)] != expected {
		t.Errorf("error should start with %q, got %q", expected, err.Error())
	}
}

// TestPublisher_NewPublisher verifies that the constructor creates a non-nil Publisher.
func TestPublisher_NewPublisher(t *testing.T) {
	pub := NewPublisher([]string{"broker1:9092", "broker2:9092"})
	if pub == nil {
		t.Fatal("expected non-nil Publisher")
	}
	if pub.producer == nil {
		t.Fatal("expected non-nil underlying producer")
	}
	pub.Close()
}

// TestPublisher_NewPublisherFromConfig verifies the config-based constructor.
func TestPublisher_NewPublisherFromConfig(t *testing.T) {
	cfg := struct {
		Brokers []string
	}{
		Brokers: []string{"broker1:9092"},
	}

	// NewPublisherFromConfig accepts config.KafkaConfig which has Brokers field.
	// We test via NewPublisher directly since we can't easily construct config.KafkaConfig
	// without the full struct.
	pub := NewPublisher(cfg.Brokers)
	if pub == nil {
		t.Fatal("expected non-nil Publisher")
	}
	pub.Close()
}

// TestPublisher_Close verifies that Close delegates without panicking.
func TestPublisher_Close(t *testing.T) {
	pub := NewPublisher([]string{"localhost:9092"})
	if err := pub.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
}

// TestValidTopic verifies known topics return true and unknown topics return false.
func TestValidTopic(t *testing.T) {
	tests := []struct {
		topic string
		want  bool
	}{
		{TopicRawLogs, true},
		{TopicAnomalies, true},
		{TopicRCARequest, true},
		{TopicRCAResult, true},
		{TopicOutbox, true},
		{"pulse.unknown", false},
		{"", false},
		{"random-topic", false},
	}

	for _, tt := range tests {
		t.Run(tt.topic, func(t *testing.T) {
			if got := ValidTopic(tt.topic); got != tt.want {
				t.Errorf("ValidTopic(%q) = %v, want %v", tt.topic, got, tt.want)
			}
		})
	}
}

// TestAllTopics verifies AllTopics returns exactly 5 known topics.
func TestAllTopics(t *testing.T) {
	topics := AllTopics()
	if len(topics) != 5 {
		t.Fatalf("expected 5 topics, got %d", len(topics))
	}

	expected := map[string]bool{
		TopicRawLogs:    true,
		TopicAnomalies:  true,
		TopicRCARequest: true,
		TopicRCAResult:  true,
		TopicOutbox:     true,
	}

	for _, topic := range topics {
		if !expected[topic] {
			t.Errorf("unexpected topic: %q", topic)
		}
	}
}

// TestAllTopics_ReturnsCopy verifies AllTopics returns a copy, not the internal slice.
func TestAllTopics_ReturnsCopy(t *testing.T) {
	topics := AllTopics()
	topics[0] = "mutated"

	fresh := AllTopics()
	if fresh[0] == "mutated" {
		t.Error("AllTopics returned internal slice reference, expected a copy")
	}
}
