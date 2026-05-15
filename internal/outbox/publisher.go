package outbox

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"pulse/internal/contract"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultPollInterval = 500 * time.Millisecond
	defaultBatchSize    = 100
	maxBackoff          = 30 * time.Second
)

// Publisher is the outbox drainer. It polls the outbox_events table for pending
// rows, publishes them to Kafka, and cleans up. It implements consumer.Listener.
type Publisher struct {
	repo         contract.OutboxRepo
	kafka        contract.KafkaPublisher
	dedup        *Deduplicator
	pollInterval time.Duration
	batchSize    int
	stopCh       chan struct{}
	stopped      atomic.Bool
	stopOnce     sync.Once
}

// NewPublisher creates a Publisher with default poll interval (500ms) and batch size (100).
func NewPublisher(repo contract.OutboxRepo, pub contract.KafkaPublisher, dedup *Deduplicator) *Publisher {
	return &Publisher{
		repo:         repo,
		kafka:        pub,
		dedup:        dedup,
		pollInterval: defaultPollInterval,
		batchSize:    defaultBatchSize,
		stopCh:       make(chan struct{}),
	}
}

// Start runs the blocking poll loop. It satisfies the consumer.Listener interface.
// Returns when Stop is called.
func (p *Publisher) Start() {
	logx.Info("[outbox-publisher] starting poll loop")

	backoff := p.pollInterval
	for {
		if p.isStopped() {
			return
		}

		ctx := context.Background()
		events, err := p.repo.FindPending(ctx, p.batchSize)
		if err != nil {
			logx.Errorf("[outbox-publisher] failed to find pending events: %v", err)
			p.sleep(backoff)
			backoff = minDuration(backoff*2, maxBackoff)
			continue
		}
		backoff = p.pollInterval // reset on successful poll

		for _, evt := range events {
			if p.isStopped() {
				return
			}

			dup, err := p.dedup.IsDuplicate(ctx, evt.ID)
			if err != nil {
				logx.Errorf("[outbox-publisher] dedup check failed for event %d: %v", evt.ID, err)
			}
			if dup {
				if err := p.repo.Delete(ctx, evt.ID); err != nil {
					logx.Errorf("[outbox-publisher] failed to delete duplicate event %d: %v", evt.ID, err)
				}
				continue
			}

			if err := p.kafka.Publish(ctx, evt.Topic, evt.AggregateID, evt.Payload); err != nil {
				logx.Errorf("[outbox-publisher] failed to publish event %d to %s: %v", evt.ID, evt.Topic, err)
				if markErr := p.repo.MarkFailed(ctx, evt.ID, evt.RetryCount+1); markErr != nil {
					logx.Errorf("[outbox-publisher] failed to mark event %d as failed: %v", evt.ID, markErr)
				}
				continue
			}

			if err := p.repo.MarkPublished(ctx, evt.ID); err != nil {
				logx.Errorf("[outbox-publisher] failed to mark event %d as published: %v", evt.ID, err)
			}
			if err := p.dedup.Mark(ctx, evt.ID); err != nil {
				logx.Errorf("[outbox-publisher] failed to mark dedup for event %d: %v", evt.ID, err)
			}
			if err := p.repo.Delete(ctx, evt.ID); err != nil {
				logx.Errorf("[outbox-publisher] failed to delete published event %d: %v", evt.ID, err)
			}
		}

		p.sleep(p.pollInterval)
	}
}

// Stop signals the poll loop to exit. Idempotent — safe to call multiple times.
func (p *Publisher) Stop() {
	p.stopOnce.Do(func() {
		logx.Info("[outbox-publisher] stopping")
		p.stopped.Store(true)
		close(p.stopCh)
	})
}

func (p *Publisher) isStopped() bool {
	select {
	case <-p.stopCh:
		return true
	default:
		return false
	}
}

// sleep waits for d but returns early if stopCh is closed.
func (p *Publisher) sleep(d time.Duration) {
	select {
	case <-p.stopCh:
	case <-time.After(d):
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
