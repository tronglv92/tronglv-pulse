package slack

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"pulse/helper/utils/httpc"
	"pulse/internal/contract"
)

var _ contract.Notifier = (*Client)(nil)

var retryConfig = httpc.RetryConfig{
	MaxRetries:     5,
	InitialBackoff: 500 * time.Millisecond,
	MaxBackoff:     10 * time.Second,
	Multiplier:     2.0,
}

// Client sends alert notifications to a Slack channel via Incoming Webhook.
type Client struct {
	http       httpc.Service
	webhookURL string
}

// New creates a Slack notifier that posts to the given webhook URL.
func New(webhookURL string) *Client {
	return &Client{
		http: httpc.NewWithClient(
			"slack",
			httpc.Client(5),
			httpc.WithJsonContentType(),
		),
		webhookURL: webhookURL,
	}
}

// SendAlert formats an AlertPayload as a Slack Block Kit message and posts it.
func (c *Client) SendAlert(ctx context.Context, p contract.AlertPayload) error {
	emoji := severityEmoji(p.Severity)

	payload := webhookPayload{
		Text: fmt.Sprintf("%s Anomaly Alert — %s (z=%.2f)", emoji, p.ServiceName, p.ZScore),
		Blocks: []block{
			{
				Type: "header",
				Text: &textObject{
					Type: "plain_text",
					Text: fmt.Sprintf("%s Anomaly Alert — %s", emoji, p.ServiceName),
				},
			},
			{
				Type: "section",
				Fields: []textObject{
					{Type: "mrkdwn", Text: fmt.Sprintf("*Service:*\n%s", p.ServiceName)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Z-Score:*\n%.2f", p.ZScore)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Severity:*\n%s", severityLabel(p.Severity))},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Incident:*\n#%d", p.IncidentID)},
				},
			},
			{
				Type: "section",
				Text: &textObject{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*RCA Summary:*\n%s", p.RCASummary),
				},
			},
			{
				Type: "section",
				Text: &textObject{
					Type: "mrkdwn",
					Text: fmt.Sprintf("<%s|View Dashboard>", p.DashboardURL),
				},
			},
		},
	}

	resp, err := httpc.WithRetry(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.http.Post(ctx, c.webhookURL, payload)
	}, retryConfig)
	if err != nil {
		return fmt.Errorf("failed to send slack alert: %w", err)
	}

	raw, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read slack response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack webhook error %d: %s", resp.StatusCode, string(raw))
	}

	return nil
}

func severityEmoji(severity int) string {
	switch {
	case severity >= 3:
		return "\xF0\x9F\x94\xB4" // red circle
	case severity == 2:
		return "\xF0\x9F\x9F\xA0" // orange circle
	default:
		return "\xF0\x9F\x9F\xA1" // yellow circle
	}
}

func severityLabel(severity int) string {
	switch {
	case severity >= 3:
		return "Critical"
	case severity == 2:
		return "Warning"
	default:
		return "Info"
	}
}
