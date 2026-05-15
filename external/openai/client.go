package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"pulse/helper/utils/httpc"
)

const (
	embeddingModel = "text-embedding-3-small"
	expectedDims   = 1536
	baseURL        = "https://api.openai.com"
)

// Client is a thin HTTP wrapper around the OpenAI Embeddings API.
// It only implements the Embed portion of contract.LLMClient.
type Client struct {
	http httpc.Service
}

// New creates an OpenAI embeddings client with the given API key.
func New(apiKey string) *Client {
	return &Client{
		http: httpc.NewWithClientAndBaseURL(
			"openai",
			httpc.Client(10),
			baseURL,
			httpc.WithAuthToken(apiKey),
			httpc.WithJsonContentType(),
		),
	}
}

// Embed calls the OpenAI Embeddings API and returns a 1536-dim float32 vector.
func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	body := embeddingRequest{
		Input:          text,
		Model:          embeddingModel,
		EncodingFormat: "float",
	}

	resp, err := c.http.Post(ctx, "/v1/embeddings", body)
	if err != nil {
		return nil, fmt.Errorf("failed to call openai embeddings API: %w", err)
	}

	raw, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read openai response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var ae apiError
		if err := json.Unmarshal(raw, &ae); err == nil && ae.Error.Message != "" {
			return nil, fmt.Errorf("openai API error %d: %s — %s", resp.StatusCode, ae.Error.Type, ae.Error.Message)
		}
		return nil, fmt.Errorf("openai API error %d: %s", resp.StatusCode, string(raw))
	}

	var er embeddingResponse
	if err := json.Unmarshal(raw, &er); err != nil {
		return nil, fmt.Errorf("failed to unmarshal openai response: %w", err)
	}

	if len(er.Data) == 0 {
		return nil, fmt.Errorf("openai: empty embedding response")
	}

	vec := er.Data[0].Embedding
	if len(vec) != expectedDims {
		return nil, fmt.Errorf("openai: expected %d dimensions, got %d", expectedDims, len(vec))
	}

	return vec, nil
}
