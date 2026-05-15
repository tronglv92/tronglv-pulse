package llm

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"

	"pulse/internal/contract"
)

var _ contract.LLMClient = (*Fake)(nil)

// Fake is a deterministic test double for contract.LLMClient.
// Given identical inputs it always produces identical outputs,
// making it suitable for snapshot-style assertions.
type Fake struct{}

// NewFake creates a new deterministic LLM test fake.
func NewFake() *Fake { return &Fake{} }

// Generate returns a deterministic response derived from the prompts via FNV-32a.
func (f *Fake) Generate(_ context.Context, req contract.GenerateRequest) (*contract.GenerateResponse, error) {
	h := fnv.New32a()
	_, _ = h.Write([]byte(req.SystemPrompt))
	_, _ = h.Write([]byte(req.UserPrompt))
	sum := h.Sum32()

	model := req.Model
	if model == "" {
		model = "fake-model"
	}

	prompt := int(sum % 512)
	completion := int((sum >> 8) % 256)

	return &contract.GenerateResponse{
		Content:          fmt.Sprintf("fake-rca-%08x", sum),
		Model:            model,
		PromptTokens:     prompt,
		CompletionTokens: completion,
		TotalTokens:      prompt + completion,
	}, nil
}

// Embed returns a deterministic 1536-dim vector derived from the text via FNV-32a.
// Values are scattered into the [-1, 1] range using Knuth multiplicative hashing.
func (f *Fake) Embed(_ context.Context, text string) ([]float32, error) {
	h := fnv.New32a()
	_, _ = h.Write([]byte(text))
	seed := h.Sum32()

	const dims = 1536
	vec := make([]float32, dims)
	for i := range vec {
		// Knuth multiplicative scatter
		seed = seed*2654435761 + uint32(i)
		vec[i] = float32(int32(seed)) / float32(math.MaxInt32)
	}
	return vec, nil
}
