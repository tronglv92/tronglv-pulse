package consumer

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/service"
)

// Compile-time assertion: Supervisor implements service.Service.
var _ service.Service = (*Supervisor)(nil)

// --- mock listener ---

type mockListener struct {
	startCount atomic.Int32
	stopCount  atomic.Int32
	started    chan struct{} // closed on first Start call
	block      chan struct{} // Start blocks until this is closed
	panicOnce  atomic.Bool  // if true, panic on the first Start call
}

func newMockListener() *mockListener {
	return &mockListener{
		started: make(chan struct{}),
		block:   make(chan struct{}),
	}
}

func (m *mockListener) Start() {
	n := m.startCount.Add(1)

	// Signal that Start was called (only on first invocation).
	if n == 1 {
		close(m.started)
	}

	// Panic on first call if configured.
	if m.panicOnce.CompareAndSwap(true, false) {
		panic("boom")
	}

	// Block until released.
	<-m.block
}

func (m *mockListener) Stop() {
	m.stopCount.Add(1)
	// Release Start so the goroutine can exit.
	select {
	case <-m.block:
		// already closed
	default:
		close(m.block)
	}
}

// --- tests ---

func TestSupervisor_StartsAllListeners(t *testing.T) {
	l1 := newMockListener()
	l2 := newMockListener()

	sup := NewSupervisor(
		Registration{Name: "l1", Listener: l1},
		Registration{Name: "l2", Listener: l2},
	)

	done := make(chan struct{})
	go func() {
		sup.Start()
		close(done)
	}()

	// Wait for both listeners to start.
	waitOrTimeout(t, l1.started, 2*time.Second)
	waitOrTimeout(t, l2.started, 2*time.Second)

	sup.Stop()
	waitOrTimeout(t, done, 2*time.Second)

	if got := l1.startCount.Load(); got < 1 {
		t.Fatalf("l1 startCount: got %d, want >= 1", got)
	}
	if got := l2.startCount.Load(); got < 1 {
		t.Fatalf("l2 startCount: got %d, want >= 1", got)
	}
}

func TestSupervisor_StopIsIdempotent(t *testing.T) {
	l := newMockListener()
	sup := NewSupervisor(Registration{Name: "l", Listener: l})

	done := make(chan struct{})
	go func() {
		sup.Start()
		close(done)
	}()

	waitOrTimeout(t, l.started, 2*time.Second)

	// Call Stop 3 times — must not panic.
	sup.Stop()
	sup.Stop()
	sup.Stop()

	waitOrTimeout(t, done, 2*time.Second)
}

func TestSupervisor_PanicRecoveryAndRestart(t *testing.T) {
	l := newMockListener()
	l.panicOnce.Store(true) // first Start panics

	sup := NewSupervisor(Registration{Name: "panicker", Listener: l})

	done := make(chan struct{})
	go func() {
		sup.Start()
		close(done)
	}()

	// Wait for the listener to have been started at least twice (panic + restart).
	deadline := time.After(5 * time.Second)
	for {
		if l.startCount.Load() >= 2 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for restart; startCount=%d", l.startCount.Load())
		case <-time.After(50 * time.Millisecond):
		}
	}

	sup.Stop()
	waitOrTimeout(t, done, 2*time.Second)

	if got := l.startCount.Load(); got < 2 {
		t.Fatalf("startCount: got %d, want >= 2", got)
	}
}

func TestSupervisor_StopDuringRestart(t *testing.T) {
	// Listener that returns immediately (simulating a crash), causing backoff.
	immediateReturn := &immediateListener{started: make(chan struct{})}

	sup := NewSupervisor(Registration{Name: "fast-exit", Listener: immediateReturn})

	done := make(chan struct{})
	go func() {
		sup.Start()
		close(done)
	}()

	// Wait for at least one start.
	waitOrTimeout(t, immediateReturn.started, 2*time.Second)

	// Give time to enter the backoff sleep, then stop.
	time.Sleep(100 * time.Millisecond)
	sup.Stop()

	// Must exit promptly (well under the 1s backoff).
	waitOrTimeout(t, done, 2*time.Second)
}

func TestSupervisor_ZeroListeners(t *testing.T) {
	sup := NewSupervisor()

	done := make(chan struct{})
	go func() {
		sup.Start()
		close(done)
	}()

	// With zero listeners, Start should return immediately.
	waitOrTimeout(t, done, 2*time.Second)

	// Stop on an already-finished supervisor must not panic.
	sup.Stop()
}

func TestSupervisor_Add(t *testing.T) {
	l := newMockListener()
	sup := NewSupervisor()
	sup.Add(Registration{Name: "added", Listener: l})

	done := make(chan struct{})
	go func() {
		sup.Start()
		close(done)
	}()

	waitOrTimeout(t, l.started, 2*time.Second)
	sup.Stop()
	waitOrTimeout(t, done, 2*time.Second)

	if got := l.startCount.Load(); got < 1 {
		t.Fatalf("added listener startCount: got %d, want >= 1", got)
	}
}

// --- helpers ---

// immediateListener returns from Start immediately (no blocking).
type immediateListener struct {
	mu      sync.Mutex
	started chan struct{}
	closed  bool
}

func (l *immediateListener) Start() {
	l.mu.Lock()
	if !l.closed {
		l.closed = true
		close(l.started)
	}
	l.mu.Unlock()
}

func (l *immediateListener) Stop() {}

func waitOrTimeout(t *testing.T, ch <-chan struct{}, timeout time.Duration) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatal("timed out waiting for channel")
	}
}
