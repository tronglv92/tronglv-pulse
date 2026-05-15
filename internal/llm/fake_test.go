package llm

import (
	"context"
	"testing"

	"pulse/internal/contract"
)

func TestFake_Generate_Deterministic(t *testing.T) {
	f := NewFake()
	req := contract.GenerateRequest{
		SystemPrompt: "you are a helpful assistant",
		UserPrompt:   "explain the error",
	}

	r1, err := f.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r2, err := f.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r1.Content != r2.Content {
		t.Errorf("expected deterministic content, got %q and %q", r1.Content, r2.Content)
	}
	if r1.PromptTokens != r2.PromptTokens {
		t.Errorf("expected deterministic prompt tokens, got %d and %d", r1.PromptTokens, r2.PromptTokens)
	}
	if r1.CompletionTokens != r2.CompletionTokens {
		t.Errorf("expected deterministic completion tokens, got %d and %d", r1.CompletionTokens, r2.CompletionTokens)
	}
}

func TestFake_Generate_Variation(t *testing.T) {
	f := NewFake()
	r1, _ := f.Generate(context.Background(), contract.GenerateRequest{
		UserPrompt: "input A",
	})
	r2, _ := f.Generate(context.Background(), contract.GenerateRequest{
		UserPrompt: "input B",
	})

	if r1.Content == r2.Content {
		t.Error("expected different content for different inputs")
	}
}

func TestFake_Generate_ModelOverride(t *testing.T) {
	f := NewFake()

	r1, _ := f.Generate(context.Background(), contract.GenerateRequest{
		UserPrompt: "test",
	})
	if r1.Model != "fake-model" {
		t.Errorf("expected default model 'fake-model', got %q", r1.Model)
	}

	r2, _ := f.Generate(context.Background(), contract.GenerateRequest{
		UserPrompt: "test",
		Model:      "claude-sonnet-4-6",
	})
	if r2.Model != "claude-sonnet-4-6" {
		t.Errorf("expected model 'claude-sonnet-4-6', got %q", r2.Model)
	}
}

func TestFake_Embed_Deterministic(t *testing.T) {
	f := NewFake()
	v1, err := f.Embed(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v2, err := f.Embed(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(v1) != 1536 {
		t.Fatalf("expected 1536 dimensions, got %d", len(v1))
	}
	for i := range v1 {
		if v1[i] != v2[i] {
			t.Fatalf("embedding mismatch at index %d: %f != %f", i, v1[i], v2[i])
		}
	}
}

func TestFake_Embed_Variation(t *testing.T) {
	f := NewFake()
	v1, _ := f.Embed(context.Background(), "text A")
	v2, _ := f.Embed(context.Background(), "text B")

	same := true
	for i := range v1 {
		if v1[i] != v2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("expected different vectors for different inputs")
	}
}

func TestFake_Embed_Dimensions(t *testing.T) {
	f := NewFake()
	vec, err := f.Embed(context.Background(), "any text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vec) != 1536 {
		t.Errorf("expected 1536 dimensions, got %d", len(vec))
	}
}
