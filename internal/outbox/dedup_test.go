package outbox

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// mockCache implements contract.Cache for testing.
type mockCache struct {
	store    map[string][]byte
	setKey   string
	setVal   []byte
	setTTL   time.Duration
	setCalls int
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

func (m *mockCache) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	m.store[key] = value
	m.setKey = key
	m.setVal = value
	m.setTTL = ttl
	m.setCalls++
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

func TestDeduplicator_IsDuplicate_True(t *testing.T) {
	mc := newMockCache()
	mc.store["evt:42"] = []byte("1")

	d := NewDeduplicator(mc)
	dup, err := d.IsDuplicate(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dup {
		t.Fatal("expected IsDuplicate to return true")
	}
}

func TestDeduplicator_IsDuplicate_False(t *testing.T) {
	mc := newMockCache()

	d := NewDeduplicator(mc)
	dup, err := d.IsDuplicate(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dup {
		t.Fatal("expected IsDuplicate to return false")
	}
}

func TestDeduplicator_Mark(t *testing.T) {
	mc := newMockCache()
	d := NewDeduplicator(mc)

	if err := d.Mark(context.Background(), 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mc.setKey != "evt:42" {
		t.Errorf("expected key evt:42, got %s", mc.setKey)
	}
	if string(mc.setVal) != "1" {
		t.Errorf("expected value '1', got %q", mc.setVal)
	}
	if mc.setTTL != time.Hour {
		t.Errorf("expected TTL 1h, got %v", mc.setTTL)
	}
}

func TestDeduplicator_KeyFormat(t *testing.T) {
	tests := []struct {
		id   int64
		want string
	}{
		{1, "evt:1"},
		{42, "evt:42"},
		{999999, "evt:999999"},
		{0, "evt:0"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("id=%d", tt.id), func(t *testing.T) {
			got := dedupKey(tt.id)
			if got != tt.want {
				t.Errorf("dedupKey(%d) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
