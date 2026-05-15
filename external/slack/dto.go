package slack

// webhookPayload is the request body for a Slack Incoming Webhook.
type webhookPayload struct {
	Text   string  `json:"text"`
	Blocks []block `json:"blocks"`
}

type block struct {
	Type   string       `json:"type"`
	Text   *textObject  `json:"text,omitempty"`
	Fields []textObject `json:"fields,omitempty"`
}

type textObject struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
