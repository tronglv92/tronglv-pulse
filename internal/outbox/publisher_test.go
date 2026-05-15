package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"pulse/internal/types/entity"
)

// ---------------------------------------------------------------------------
// Mock: OutboxRepo
// ---------------------------------------------------------------------------

type mockOutboxRepo struct {
	mu              sync.Mutex
	findPendingFn   func(ctx context.Context, limit int) ([]*entity.OutboxEvent, error)
	markPublishedFn func(ctx context.Context, id int64) error
	markFailedFn    func(ctx context.Context, id int64, retryCount int16) error
	deleteFn        func(ctx context.Context, id int64) error

	// tracking
	markPublishedIDs []int64
	markFailedCalls  []markFailedCall
	deletedIDs       []int64
}

type markFailedCall struct {
	ID         int64
	RetryCount int16
}

func (m *mockOutboxRepo) FindPending(ctx context.Context, limit int) ([]*entity.OutboxEvent, error) {
	if m.findPendingFn != nil {
		return m.findPendingFn(ctx, limit)
	}
	return nil, nil
}

func (m *mockOutboxRepo) MarkPublished(ctx context.Context, id int64) error {
	m.mu.Lock()
	m.markPublishedIDs = append(m.markPublishedIDs, id)
	m.mu.Unlock()
	if m.markPublishedFn != nil {
		return m.markPublishedFn(ctx, id)
	}
	return nil
}

func (m *mockOutboxRepo) MarkFailed(ctx context.Context, id int64, retryCount int16) error {
	m.mu.Lock()
	m.markFailedCalls = append(m.markFailedCalls, markFailedCall{ID: id, RetryCount: retryCount})
	m.mu.Unlock()
	if m.markFailedFn != nil {
		return m.markFailedFn(ctx, id, retryCount)
	}
	return nil
}

func (m *mockOutboxRepo) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	m.deletedIDs = append(m.deletedIDs, id)
	m.mu.Unlock()
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Mock: KafkaPublisher
// ---------------------------------------------------------------------------

type mockKafkaPublisher struct {
	mu        sync.Mutex
	publishFn func(ctx context.Context, topic, key string, value any) error
	calls     []kafkaPublishCall
}

type kafkaPublishCall struct {
	Topic string
	Key   string
	Value any
}

func (m *mockKafkaPublisher) Publish(ctx context.Context, topic, key string, value any) error {
	m.mu.Lock()
	m.calls = append(m.calls, kafkaPublishCall{Topic: topic, Key: key, Value: value})
	m.mu.Unlock()
	if m.publishFn != nil {
		return m.publishFn(ctx, topic, key, value)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeEvent(id int64, topic, aggregateID string, retryCount int16) *entity.OutboxEvent {
	return &entity.OutboxEvent{
		ID:          id,
		AggregateID: aggregateID,
		EventType:   "test.event",
		Topic:       topic,
		Payload:     json.RawMessage(`{"test":true}`),
		RetryCount:  retryCount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// startPublisher creates and starts a Publisher with fast poll interval.
// Returns the publisher and a cleanup function.
func startPublisher(repo *mockOutboxRepo, kafka *mockKafkaPublisher, dedup *Deduplicator) *Publisher {
	p := NewPublisher(repo, kafka, dedup)
	p.pollInterval = time.Millisecond
	p.batchSize = 10
	go p.Start()
	return p
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestNewPublisher(t *testing.T) {
	repo := &mockOutboxRepo{}
	kafka := &mockKafkaPublisher{}
	dedup := NewDeduplicator(newMockCache())

	p := NewPublisher(repo, kafka, dedup)
	if p == nil {
		t.Fatal("expected non-nil Publisher")
	}
	if p.pollInterval != defaultPollInterval {
		t.Errorf("pollInterval = %v, want %v", p.pollInterval, defaultPollInterval)
	}
	if p.batchSize != defaultBatchSize {
		t.Errorf("batchSize = %d, want %d", p.batchSize, defaultBatchSize)
	}
}

func TestPublisher_StopIsIdempotent(t *testing.T) {
	repo := &mockOutboxRepo{}
	kafka := &mockKafkaPublisher{}
	dedup := NewDeduplicator(newMockCache())

	p := NewPublisher(repo, kafka, dedup)
	p.pollInterval = time.Millisecond
	go p.Start()

	// Multiple Stop calls should not panic.
	p.Stop()
	p.Stop()
	p.Stop()
}

func TestPublisher_EmptyBatch(t *testing.T) {
	var pollCount atomic.Int32
	done := make(chan struct{})

	repo := &mockOutboxRepo{
		findPendingFn: func(_ context.Context, _ int) ([]*entity.OutboxEvent, error) {
			if pollCount.Add(1) >= 2 {
				select {
				case done <- struct{}{}:
				default:
				}
			}
			return nil, nil // empty batch
		},
	}
	kafka := &mockKafkaPublisher{}
	dedup := NewDeduplicator(newMockCache())

	p := startPublisher(repo, kafka, dedup)
	defer p.Stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for empty polls")
	}

	kafka.mu.Lock()
	kafkaCalls := len(kafka.calls)
	kafka.mu.Unlock()
	if kafkaCalls != 0 {
		t.Errorf("expected 0 kafka calls, got %d", kafkaCalls)
	}
}

func TestPublisher_HappyPath_SingleEvent(t *testing.T) {
	evt := makeEvent(1, "logs.enriched", "agg-1", 0)
	var pollCount atomic.Int32
	done := make(chan struct{})

	repo := &mockOutboxRepo{
		findPendingFn: func(_ context.Context, _ int) ([]*entity.OutboxEvent, error) {
			n := pollCount.Add(1)
			if n == 1 {
				return []*entity.OutboxEvent{evt}, nil
			}
			select {
			case done <- struct{}{}:
			default:
			}
			return nil, nil
		},
	}
	kafka := &mockKafkaPublisher{}
	mc := newMockCache()
	dedup := NewDeduplicator(mc)

	p := startPublisher(repo, kafka, dedup)
	defer p.Stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}

	// Verify Kafka publish
	kafka.mu.Lock()
	if len(kafka.calls) != 1 {
		t.Fatalf("expected 1 kafka call, got %d", len(kafka.calls))
	}
	if kafka.calls[0].Topic != "logs.enriched" {
		t.Errorf("topic = %q, want %q", kafka.calls[0].Topic, "logs.enriched")
	}
	if kafka.calls[0].Key != "agg-1" {
		t.Errorf("key = %q, want %q", kafka.calls[0].Key, "agg-1")
	}
	kafka.mu.Unlock()

	// Verify MarkPublished
	repo.mu.Lock()
	if len(repo.markPublishedIDs) != 1 || repo.markPublishedIDs[0] != 1 {
		t.Errorf("markPublished = %v, want [1]", repo.markPublishedIDs)
	}
	// Verify Delete
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 1 {
		t.Errorf("deleted = %v, want [1]", repo.deletedIDs)
	}
	repo.mu.Unlock()

	// Verify dedup marked
	if _, ok := mc.store["evt:1"]; !ok {
		t.Error("expected dedup key evt:1 to be set")
	}
}

func TestPublisher_HappyPath_MultipleBatch(t *testing.T) {
	events := []*entity.OutboxEvent{
		makeEvent(10, "topic-a", "agg-10", 0),
		makeEvent(20, "topic-b", "agg-20", 0),
		makeEvent(30, "topic-c", "agg-30", 0),
	}
	var pollCount atomic.Int32
	done := make(chan struct{})

	repo := &mockOutboxRepo{
		findPendingFn: func(_ context.Context, _ int) ([]*entity.OutboxEvent, error) {
			n := pollCount.Add(1)
			if n == 1 {
				return events, nil
			}
			select {
			case done <- struct{}{}:
			default:
			}
			return nil, nil
		},
	}
	kafka := &mockKafkaPublisher{}
	dedup := NewDeduplicator(newMockCache())

	p := startPublisher(repo, kafka, dedup)
	defer p.Stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}

	kafka.mu.Lock()
	if len(kafka.calls) != 3 {
		t.Fatalf("expected 3 kafka calls, got %d", len(kafka.calls))
	}
	kafka.mu.Unlock()

	repo.mu.Lock()
	if len(repo.markPublishedIDs) != 3 {
		t.Errorf("expected 3 markPublished calls, got %d", len(repo.markPublishedIDs))
	}
	if len(repo.deletedIDs) != 3 {
		t.Errorf("expected 3 delete calls, got %d", len(repo.deletedIDs))
	}
	repo.mu.Unlock()
}

func TestPublisher_DuplicateEvent_SkipsPublish(t *testing.T) {
	evt := makeEvent(42, "logs.enriched", "agg-42", 0)
	var pollCount atomic.Int32
	done := make(chan struct{})

	repo := &mockOutboxRepo{
		findPendingFn: func(_ context.Context, _ int) ([]*entity.OutboxEvent, error) {
			n := pollCount.Add(1)
			if n == 1 {
				return []*entity.OutboxEvent{evt}, nil
			}
			select {
			case done <- struct{}{}:
			default:
			}
			return nil, nil
		},
	}
	kafka := &mockKafkaPublisher{}

	// Pre-mark as duplicate
	mc := newMockCache()
	mc.store["evt:42"] = []byte("1")
	dedup := NewDeduplicator(mc)

	p := startPublisher(repo, kafka, dedup)
	defer p.Stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}

	// Kafka should NOT be called for a duplicate
	kafka.mu.Lock()
	if len(kafka.calls) != 0 {
		t.Errorf("expected 0 kafka calls for dup, got %d", len(kafka.calls))
	}
	kafka.mu.Unlock()

	// MarkPublished should NOT be called
	repo.mu.Lock()
	if len(repo.markPublishedIDs) != 0 {
		t.Errorf("expected 0 markPublished for dup, got %d", len(repo.markPublishedIDs))
	}
	// Delete SHOULD be called to clean up the dup row
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 42 {
		t.Errorf("expected delete of dup id 42, got %v", repo.deletedIDs)
	}
	repo.mu.Unlock()
}

func TestPublisher_DedupCheckError_StillPublishes(t *testing.T) {
	evt := makeEvent(7, "logs.enriched", "agg-7", 0)
	var pollCount atomic.Int32
	done := make(chan struct{})

	repo := &mockOutboxRepo{
		findPendingFn: func(_ context.Context, _ int) ([]*entity.OutboxEvent, error) {
			n := pollCount.Add(1)
			if n == 1 {
				return []*entity.OutboxEvent{evt}, nil
			}
			select {
			case done <- struct{}{}:
			default:
			}
			return nil, nil
		},
	}
	kafka := &mockKafkaPublisher{}

	// Cache that returns error on Exists
	mc := &errorOnExistsCache{mockCache: newMockCache()}
	dedup := NewDeduplicator(mc)

	p := startPublisher(repo, kafka, dedup)
	defer p.Stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}

	// Should still publish despite dedup error
	kafka.mu.Lock()
	if len(kafka.calls) != 1 {
		t.Errorf("expected 1 kafka call despite dedup error, got %d", len(kafka.calls))
	}
	kafka.mu.Unlock()
}

// errorOnExistsCache wraps mockCache but always returns error on Exists.
type errorOnExistsCache struct {
	*mockCache
}

func (e *errorOnExistsCache) Exists(_ context.Context, _ ...string) (bool, error) {
	return false, errors.New("redis connection failed")
}

func TestPublisher_KafkaPublishError_MarksFailed(t *testing.T) {
	evt := makeEvent(5, "logs.enriched", "agg-5", 2) // retryCount=2
	var pollCount atomic.Int32
	done := make(chan struct{})

	repo := &mockOutboxRepo{
		findPendingFn: func(_ context.Context, _ int) ([]*entity.OutboxEvent, error) {
			n := pollCount.Add(1)
			if n == 1 {
				return []*entity.OutboxEvent{evt}, nil
			}
			select {
			case done <- struct{}{}:
			default:
			}
			return nil, nil
		},
	}
	kafka := &mockKafkaPublisher{
		publishFn: func(_ context.Context, _, _ string, _ any) error {
			return errors.New("kafka unavailable")
		},
	}
	dedup := NewDeduplicator(newMockCache())

	p := startPublisher(repo, kafka, dedup)
	defer p.Stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}

	// Should call MarkFailed with retryCount+1
	repo.mu.Lock()
	if len(repo.markFailedCalls) != 1 {
		t.Fatalf("expected 1 markFailed call, got %d", len(repo.markFailedCalls))
	}
	if repo.markFailedCalls[0].ID != 5 {
		t.Errorf("markFailed id = %d, want 5", repo.markFailedCalls[0].ID)
	}
	if repo.markFailedCalls[0].RetryCount != 3 {
		t.Errorf("markFailed retryCount = %d, want 3", repo.markFailedCalls[0].RetryCount)
	}
	// Should NOT be deleted or mark published
	if len(repo.deletedIDs) != 0 {
		t.Errorf("expected 0 deletes on kafka error, got %d", len(repo.deletedIDs))
	}
	if len(repo.markPublishedIDs) != 0 {
		t.Errorf("expected 0 markPublished on kafka error, got %d", len(repo.markPublishedIDs))
	}
	repo.mu.Unlock()
}

func TestPublisher_MarkPublishedError_Continues(t *testing.T) {
	events := []*entity.OutboxEvent{
		makeEvent(1, "t1", "a1", 0),
		makeEvent(2, "t2", "a2", 0),
	}
	var pollCount atomic.Int32
	done := make(chan struct{})

	repo := &mockOutboxRepo{
		findPendingFn: func(_ context.Context, _ int) ([]*entity.OutboxEvent, error) {
			n := pollCount.Add(1)
			if n == 1 {
				return events, nil
			}
			select {
			case done <- struct{}{}:
			default:
			}
			return nil, nil
		},
		markPublishedFn: func(_ context.Context, _ int64) error {
			return errors.New("mark published failed")
		},
	}
	kafka := &mockKafkaPublisher{}
	dedup := NewDeduplicator(newMockCache())

	p := startPublisher(repo, kafka, dedup)
	defer p.Stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}

	// Both events should be published despite MarkPublished errors
	kafka.mu.Lock()
	if len(kafka.calls) != 2 {
		t.Errorf("expected 2 kafka calls, got %d", len(kafka.calls))
	}
	kafka.mu.Unlock()
}

func TestPublisher_DeleteError_Continues(t *testing.T) {
	events := []*entity.OutboxEvent{
		makeEvent(1, "t1", "a1", 0),
		makeEvent(2, "t2", "a2", 0),
	}
	var pollCount atomic.Int32
	done := make(chan struct{})

	repo := &mockOutboxRepo{
		findPendingFn: func(_ context.Context, _ int) ([]*entity.OutboxEvent, error) {
			n := pollCount.Add(1)
			if n == 1 {
				return events, nil
			}
			select {
			case done <- struct{}{}:
			default:
			}
			return nil, nil
		},
		deleteFn: func(_ context.Context, _ int64) error {
			return errors.New("delete failed")
		},
	}
	kafka := &mockKafkaPublisher{}
	dedup := NewDeduplicator(newMockCache())

	p := startPublisher(repo, kafka, dedup)
	defer p.Stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}

	// Both events should be published despite Delete errors
	kafka.mu.Lock()
	if len(kafka.calls) != 2 {
		t.Errorf("expected 2 kafka calls, got %d", len(kafka.calls))
	}
	kafka.mu.Unlock()

	repo.mu.Lock()
	if len(repo.markPublishedIDs) != 2 {
		t.Errorf("expected 2 markPublished, got %d", len(repo.markPublishedIDs))
	}
	repo.mu.Unlock()
}

func TestPublisher_FindPendingError_Backoff(t *testing.T) {
	var pollCount atomic.Int32
	done := make(chan struct{})

	repo := &mockOutboxRepo{
		findPendingFn: func(_ context.Context, _ int) ([]*entity.OutboxEvent, error) {
			n := pollCount.Add(1)
			if n <= 2 {
				return nil, errors.New("db connection lost")
			}
			// 3rd call succeeds — confirms backoff recovery
			select {
			case done <- struct{}{}:
			default:
			}
			return nil, nil
		},
	}
	kafka := &mockKafkaPublisher{}
	dedup := NewDeduplicator(newMockCache())

	p := NewPublisher(repo, kafka, dedup)
	p.pollInterval = time.Millisecond
	p.batchSize = 10
	go p.Start()
	defer p.Stop()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for backoff recovery")
	}

	// Should have polled at least 3 times (2 errors + 1 success)
	if pollCount.Load() < 3 {
		t.Errorf("expected >= 3 polls, got %d", pollCount.Load())
	}

	// No kafka calls since we returned empty on success
	kafka.mu.Lock()
	if len(kafka.calls) != 0 {
		t.Errorf("expected 0 kafka calls, got %d", len(kafka.calls))
	}
	kafka.mu.Unlock()
}

func TestPublisher_StopDuringEventProcessing(t *testing.T) {
	stopTriggered := make(chan struct{})
	var pub *Publisher

	events := []*entity.OutboxEvent{
		makeEvent(1, "t1", "a1", 0),
		makeEvent(2, "t2", "a2", 0),
		makeEvent(3, "t3", "a3", 0),
	}
	var pollCount atomic.Int32

	repo := &mockOutboxRepo{
		findPendingFn: func(_ context.Context, _ int) ([]*entity.OutboxEvent, error) {
			n := pollCount.Add(1)
			if n == 1 {
				return events, nil
			}
			return nil, nil
		},
	}
	kafka := &mockKafkaPublisher{
		publishFn: func(_ context.Context, _, _ string, _ any) error {
			// Stop after the first publish
			select {
			case <-stopTriggered:
			default:
				close(stopTriggered)
				pub.Stop()
			}
			return nil
		},
	}
	dedup := NewDeduplicator(newMockCache())

	pub = NewPublisher(repo, kafka, dedup)
	pub.pollInterval = time.Millisecond
	pub.batchSize = 10

	done := make(chan struct{})
	go func() {
		pub.Start()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publisher did not exit promptly after Stop during processing")
	}

	// At least 1 event was published; remaining may or may not depending on timing
	kafka.mu.Lock()
	if len(kafka.calls) < 1 {
		t.Error("expected at least 1 kafka call before stop")
	}
	kafka.mu.Unlock()
}

func TestPublisher_StopDuringSleep(t *testing.T) {
	var pollCount atomic.Int32

	repo := &mockOutboxRepo{
		findPendingFn: func(_ context.Context, _ int) ([]*entity.OutboxEvent, error) {
			pollCount.Add(1)
			return nil, nil // empty — publisher will sleep
		},
	}
	kafka := &mockKafkaPublisher{}
	dedup := NewDeduplicator(newMockCache())

	p := NewPublisher(repo, kafka, dedup)
	// Use a long poll interval so we're likely sleeping when Stop is called
	p.pollInterval = 5 * time.Second
	p.batchSize = 10

	done := make(chan struct{})
	go func() {
		p.Start()
		close(done)
	}()

	// Wait for at least one poll to complete (publisher enters sleep)
	for pollCount.Load() < 1 {
		time.Sleep(time.Millisecond)
	}

	p.Stop()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("publisher did not exit promptly when stopped during sleep")
	}
}
