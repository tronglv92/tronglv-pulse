package kafka

import "pulse/internal/types/define/constant"

// Re-export topic constants for package-local convenience.
const (
	TopicRawLogs    = constant.TopicRawLogs    // "pulse.logs.raw"
	TopicAnomalies  = constant.TopicAnomalies  // "pulse.anomalies"
	TopicRCARequest = constant.TopicRCARequest // "pulse.rca.request"
	TopicRCAResult  = constant.TopicRCAResult  // "pulse.rca.result"
	TopicOutbox     = constant.TopicOutbox     // "pulse.outbox"
)

var allTopics = []string{
	TopicRawLogs,
	TopicAnomalies,
	TopicRCARequest,
	TopicRCAResult,
	TopicOutbox,
}

// AllTopics returns a slice of all known Kafka topics.
func AllTopics() []string {
	out := make([]string, len(allTopics))
	copy(out, allTopics)
	return out
}

// ValidTopic returns true if topic is one of the known Pulse topics.
func ValidTopic(topic string) bool {
	for _, t := range allTopics {
		if t == topic {
			return true
		}
	}
	return false
}
