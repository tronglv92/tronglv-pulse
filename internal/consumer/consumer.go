package consumer

import (
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const restartBackoff = time.Second

// Listener is the minimal interface a Kafka consumer queue must satisfy.
// Both queue.MessageQueue and service.Service from go-zero satisfy this.
type Listener interface {
	Start()
	Stop()
}

// Registration pairs a human-readable name with a Listener for logging.
type Registration struct {
	Name     string
	Listener Listener
}

// Supervisor manages a set of Listeners with per-goroutine panic recovery
// and automatic restart. It implements service.Service so it can be added
// to a go-zero ServiceGroup.
type Supervisor struct {
	listeners []Registration
	stopCh    chan struct{}
	stopped   atomic.Bool
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

// NewSupervisor creates a Supervisor with no listeners.
// Use Add to register listeners before calling Start.
func NewSupervisor(regs ...Registration) *Supervisor {
	return &Supervisor{
		listeners: regs,
		stopCh:    make(chan struct{}),
	}
}

// Add appends a Registration. Must be called before Start.
func (s *Supervisor) Add(reg Registration) {
	s.listeners = append(s.listeners, reg)
}

// Start launches one goroutine per registration and blocks until Stop is called
// (or all goroutines exit). This satisfies the service.Starter interface.
func (s *Supervisor) Start() {
	for i := range s.listeners {
		reg := s.listeners[i]
		s.wg.Add(1)
		go s.runListener(reg)
	}
	s.wg.Wait()
}

// Stop gracefully shuts down all listeners. Idempotent — safe to call multiple times.
// This satisfies the service.Stopper interface.
func (s *Supervisor) Stop() {
	s.stopOnce.Do(func() {
		s.stopped.Store(true)
		for i := range s.listeners {
			s.listeners[i].Listener.Stop()
		}
		close(s.stopCh)
	})
}

// runListener is the restart loop for a single listener.
func (s *Supervisor) runListener(reg Registration) {
	defer s.wg.Done()

	for {
		// Check if we should exit before (re)starting.
		select {
		case <-s.stopCh:
			return
		default:
		}

		logx.Infof("[supervisor] starting listener %q", reg.Name)
		s.runWithRecovery(reg)

		// After Start returns (normally or via panic recovery), check if stopped.
		if s.stopped.Load() {
			return
		}

		logx.Infof("[supervisor] listener %q exited, restarting after backoff", reg.Name)

		// Interruptible backoff sleep.
		select {
		case <-s.stopCh:
			return
		case <-time.After(restartBackoff):
		}
	}
}

// runWithRecovery calls reg.Listener.Start() and recovers from panics.
func (s *Supervisor) runWithRecovery(reg Registration) {
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("[supervisor] listener %q panicked: %v\n%s", reg.Name, r, debug.Stack())
		}
	}()
	reg.Listener.Start()
}
