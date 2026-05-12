package event

import "time"

type RCACompletedEvent struct {
	TenantID    int64     `json:"tenant_id"`
	IncidentID  int64     `json:"incident_id"`
	Fingerprint string    `json:"fingerprint"`
	Summary     string    `json:"summary"`
	Model       string    `json:"model"`
	CompletedAt time.Time `json:"completed_at"`
}
