package contract

import "context"

// AlertPayload is the channel-agnostic envelope for operational alerts.
// The concrete Notifier implementation maps these fields to the target
// channel format (Slack block kit, PagerDuty event, etc.).
type AlertPayload struct {
	TenantID     int64
	IncidentID   int64
	ServiceName  string
	ZScore       float64
	Severity     int
	RCASummary   string
	DashboardURL string
}

// Notifier delivers operational alerts to an external channel.
type Notifier interface {
	SendAlert(ctx context.Context, p AlertPayload) error
}
