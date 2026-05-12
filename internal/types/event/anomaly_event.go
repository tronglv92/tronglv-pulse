package event

import "time"

type AnomalyDetectedEvent struct {
	TenantID    int64     `json:"tenant_id"`
	ServiceName string    `json:"service_name"`
	Fingerprint string    `json:"fingerprint"`
	ZScore      float64   `json:"z_score"`
	WindowSize  int       `json:"window_size"`
	LogCount    int       `json:"log_count"`
	DetectedAt  time.Time `json:"detected_at"`
}
