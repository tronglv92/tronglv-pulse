package anthropic

// messagesRequest is the request body for the Anthropic Messages API.
type messagesRequest struct {
	Model     string    `json:"model"`
	MaxTokens int32     `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// messagesResponse is the response body from the Anthropic Messages API.
type messagesResponse struct {
	ID         string         `json:"id"`
	Model      string         `json:"model"`
	Content    []contentBlock `json:"content"`
	Usage      usage          `json:"usage"`
	StopReason string         `json:"stop_reason"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// apiError is the error response body from the Anthropic Messages API.
type apiError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}
