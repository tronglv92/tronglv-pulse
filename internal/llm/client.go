package llm

import (
	"context"
	"fmt"

	"pulse/internal/contract"
	"pulse/internal/types/define/constant"
)

// Generator is the subset of contract.LLMClient that produces text.
type Generator interface {
	Generate(ctx context.Context, req contract.GenerateRequest) (*contract.GenerateResponse, error)
}

// Embedder is the subset of contract.LLMClient that produces vectors.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

var _ contract.LLMClient = (*Client)(nil)

// Client composes a Generator and an Embedder into a full contract.LLMClient.
// Every call is gated by the Redis budget guard (llm:disabled key).
type Client struct {
	gen   Generator
	embed Embedder
	cache contract.Cache
}

// New creates a composite LLM client with budget guard.
func New(gen Generator, embed Embedder, cache contract.Cache) *Client {
	return &Client{gen: gen, embed: embed, cache: cache}
}

// Generate checks the budget guard then delegates to the Generator.
func (c *Client) Generate(ctx context.Context, req contract.GenerateRequest) (*contract.GenerateResponse, error) {
	if err := c.checkBudget(ctx); err != nil {
		return nil, err
	}
	return c.gen.Generate(ctx, req)
}

// Embed checks the budget guard then delegates to the Embedder.
func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	if err := c.checkBudget(ctx); err != nil {
		return nil, err
	}
	return c.embed.Embed(ctx, text)
}

// checkBudget returns an error when the LLM daily budget has been exhausted.
// On Redis errors it fails open (returns nil) to avoid blocking calls when
// the cache is temporarily unavailable.
func (c *Client) checkBudget(ctx context.Context) error {
	disabled, err := c.cache.Exists(ctx, constant.RedisKeyLLMDisabled)
	if err != nil {
		// Fail open — Redis is unreachable, allow the call.
		return nil
	}
	if disabled {
		return fmt.Errorf("llm: daily budget exhausted (llm:disabled flag is set)")
	}
	return nil
}
