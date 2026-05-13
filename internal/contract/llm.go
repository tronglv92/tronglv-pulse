package contract

import "context"

// GenerateRequest bundles prompt parameters for a chat-completion call.
type GenerateRequest struct {
	SystemPrompt string
	UserPrompt   string
	// Model overrides the client default if non-empty (e.g., "claude-sonnet-4-6").
	Model     string
	MaxTokens int32
}

// GenerateResponse carries the assistant's message and billing metadata.
type GenerateResponse struct {
	Content          string
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// LLMClient abstracts text generation and vector embedding.
// All implementations must check the llm:disabled Redis flag before making
// any network call (see budget guard in CLAUDE.md).
type LLMClient interface {
	// Generate calls the chat-completion API and returns the assistant message.
	Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
	// Embed returns a 1536-dim float32 slice.
	// Callers convert to pgvector.Vector before storing in ErrorEmbedding.
	Embed(ctx context.Context, text string) ([]float32, error)
}
