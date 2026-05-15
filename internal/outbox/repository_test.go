package outbox

import (
	"testing"

	"pulse/internal/contract"
	"pulse/internal/types/entity"
)

// Compile-time interface assertions.
var (
	_ contract.OutboxRepo     = (*Repository)(nil)
	_ contract.OutboxAppender = (*Repository)(nil)
)

func TestNewRepository(t *testing.T) {
	r := NewRepository(nil)
	if r == nil {
		t.Fatal("expected non-nil Repository")
	}
}

func TestOutboxEvent_TableName(t *testing.T) {
	e := entity.OutboxEvent{}
	if e.TableName() != "outbox_events" {
		t.Errorf("TableName() = %q, want %q", e.TableName(), "outbox_events")
	}
}
