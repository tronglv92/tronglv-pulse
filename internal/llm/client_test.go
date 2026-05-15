package llm

import (
	"context"
	"strings"
	"testing"
	"time"

	"pulse/internal/contract"
	"pulse/internal/types/define/constant"
)

// mockCache implements contract.Cache for testing.
type mockCache struct {
	store map[string][]byte
}

func newMockCache() *mockCache {
	return &mockCache{store: make(map[string][]byte)}
}

func (m *mockCache) Get(_ context.Context, key string) ([]byte, error) {
	v, ok := m.store[key]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (m *mockCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	m.store[key] = value
	return nil
}

func (m *mockCache) Del(_ context.Context, keys ...string) error {
	for _, k := range keys {
		delete(m.store, k)
	}
	return nil
}

func (m *mockCache) Exists(_ context.Context, keys ...string) (bool, error) {
	for _, k := range keys {
		if _, ok := m.store[k]; !ok {
			return false, nil
		}
	}
	return true, nil
}

func TestClient_Generate_BudgetDisabled(t *testing.T) {
	mc := newMockCache()
	mc.store[constant.RedisKeyLLMDisabled] = []byte("1")

	c := New(NewFake(), NewFake(), mc)
	_, err := c.Generate(context.Background(), contract.GenerateRequest{
		UserPrompt: "test",
	})
	if err == nil {
		t.Fatal("expected error when budget is disabled")
	}
	if !strings.Contains(err.Error(), "budget exhausted") {
		t.Errorf("expected budget exhausted error, got: %v", err)
	}
}

func TestClient_Generate_BudgetAllowed(t *testing.T) {
	mc := newMockCache()
	c := New(NewFake(), NewFake(), mc)

	resp, err := c.Generate(context.Background(), contract.GenerateRequest{
		UserPrompt: "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestClient_Embed_BudgetDisabled(t *testing.T) {
	mc := newMockCache()
	mc.store[constant.RedisKeyLLMDisabled] = []byte("1")

	c := New(NewFake(), NewFake(), mc)
	_, err := c.Embed(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error when budget is disabled")
	}
	if !strings.Contains(err.Error(), "budget exhausted") {
		t.Errorf("expected budget exhausted error, got: %v", err)
	}
}

func TestClient_Embed_BudgetAllowed(t *testing.T) {
	mc := newMockCache()
	c := New(NewFake(), NewFake(), mc)

	vec, err := c.Embed(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vec) != 1536 {
		t.Errorf("expected 1536 dimensions, got %d", len(vec))
	}
}
