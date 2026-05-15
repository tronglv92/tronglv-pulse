package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"pulse/helper/utils/httpc"
	"pulse/internal/contract"
)

const (
	defaultMaxTokens = int32(4096)
	apiVersion       = "2023-06-01"
	baseURL          = "https://api.anthropic.com"
	maxAttempts      = 2
)

// Client is a thin HTTP wrapper around the Anthropic Messages API.
// It only implements the Generate portion of contract.LLMClient.
type Client struct {
	http         httpc.Service
	defaultModel string
}

// New creates an Anthropic client with the given API key and default model.
func New(apiKey, defaultModel string) *Client {
	return &Client{
		http: httpc.NewWithClientAndBaseURL(
			"anthropic",
			httpc.Client(20),
			baseURL,
			httpc.WithHeaders(map[string]string{
				"x-api-key":         apiKey,
				"anthropic-version": apiVersion,
				"Content-Type":      "application/json",
			}),
		),
		defaultModel: defaultModel,
	}
}

// Generate calls the Anthropic Messages API and returns the assistant message.
func (c *Client) Generate(ctx context.Context, req contract.GenerateRequest) (*contract.GenerateResponse, error) {
	model := req.Model
	if model == "" {
		model = c.defaultModel
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	}

	body := messagesRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    req.SystemPrompt,
		Messages: []message{
			{Role: "user", Content: req.UserPrompt},
		},
	}

	// Retry once on JSON parse failure.
	for attempt := 0; attempt < maxAttempts; attempt++ {
		resp, err := c.http.Post(ctx, "/v1/messages", body)
		if err != nil {
			return nil, fmt.Errorf("failed to call anthropic messages API: %w", err)
		}

		raw, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read anthropic response body: %w", err)
		}

		if resp.StatusCode >= 400 {
			var ae apiError
			if err := json.Unmarshal(raw, &ae); err == nil && ae.Error.Message != "" {
				return nil, fmt.Errorf("anthropic API error %d: %s — %s", resp.StatusCode, ae.Error.Type, ae.Error.Message)
			}
			return nil, fmt.Errorf("anthropic API error %d: %s", resp.StatusCode, string(raw))
		}

		var mr messagesResponse
		if err := json.Unmarshal(raw, &mr); err != nil {
			if attempt < maxAttempts-1 {
				continue
			}
			return nil, fmt.Errorf("failed to unmarshal anthropic response: %w", err)
		}

		var parts []string
		for _, block := range mr.Content {
			if block.Type == "text" {
				parts = append(parts, block.Text)
			}
		}

		return &contract.GenerateResponse{
			Content:          strings.Join(parts, ""),
			Model:            mr.Model,
			PromptTokens:     mr.Usage.InputTokens,
			CompletionTokens: mr.Usage.OutputTokens,
			TotalTokens:      mr.Usage.InputTokens + mr.Usage.OutputTokens,
		}, nil
	}

	// Unreachable, but the compiler doesn't know that.
	return nil, fmt.Errorf("anthropic: exhausted retry attempts")
}
