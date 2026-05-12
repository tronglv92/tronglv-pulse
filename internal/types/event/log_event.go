package event

import "time"

type RawLogEvent struct {
	TenantID    int64          `json:"tenant_id"`
	ServiceName string         `json:"service_name"`
	Level       string         `json:"level"`
	Message     string         `json:"message"`
	Fingerprint string         `json:"fingerprint"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Timestamp   time.Time      `json:"timestamp"`
}
