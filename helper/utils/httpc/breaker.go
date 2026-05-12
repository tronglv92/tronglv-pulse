package httpc

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/rest/httpc"
)

// BreakerService wraps an httpc.Service with circuit breaker protection.
// The circuit breaker trips open when too many requests fail, preventing
// further requests to a failing service and allowing it time to recover.
//
// The breaker only trips on network errors, not HTTP errors (4xx/5xx),
// because HTTP errors indicate the service is responding even if with errors.
type BreakerService struct {
	inner Service
	brk   breaker.Breaker
}

// NewBreakerService creates a new Service with circuit breaker protection.
//
// Parameters:
//   - inner: The underlying httpc.Service to wrap
//   - name: Circuit breaker name for metrics and logging
//
// Example:
//
//	httpClient := httpc.New("api-client")
//	protectedClient := httpc.NewBreakerService(httpClient, "core-service")
func NewBreakerService(inner Service, name string) Service {
	return &BreakerService{
		inner: inner,
		brk:   breaker.NewBreaker(breaker.WithName(name)),
	}
}

// acceptable determines if an error should trip the circuit breaker.
// Only network-level errors trip the breaker; HTTP errors (even 5xx) are considered acceptable
// because they indicate the service is reachable and responding.
func acceptable(err error) bool {
	return err == nil
}

// Head implements Service.Head with circuit breaker protection.
func (s *BreakerService) Head(ctx context.Context, url string, opts ...Option) (*http.Response, error) {
	var resp *http.Response
	err := s.brk.DoWithAcceptable(func() error {
		var e error
		resp, e = s.inner.Head(ctx, url, opts...)
		return e
	}, acceptable)
	return resp, err
}

// Get implements Service.Get with circuit breaker protection.
func (s *BreakerService) Get(ctx context.Context, url string, opts ...Option) (*http.Response, error) {
	var resp *http.Response
	err := s.brk.DoWithAcceptable(func() error {
		var e error
		resp, e = s.inner.Get(ctx, url, opts...)
		return e
	}, acceptable)
	return resp, err
}

// Post implements Service.Post with circuit breaker protection.
func (s *BreakerService) Post(ctx context.Context, url string, payload any, opts ...Option) (*http.Response, error) {
	var resp *http.Response
	err := s.brk.DoWithAcceptable(func() error {
		var e error
		resp, e = s.inner.Post(ctx, url, payload, opts...)
		return e
	}, acceptable)
	return resp, err
}

// Put implements Service.Put with circuit breaker protection.
func (s *BreakerService) Put(ctx context.Context, url string, payload any, opts ...Option) (*http.Response, error) {
	var resp *http.Response
	err := s.brk.DoWithAcceptable(func() error {
		var e error
		resp, e = s.inner.Put(ctx, url, payload, opts...)
		return e
	}, acceptable)
	return resp, err
}

// Patch implements Service.Patch with circuit breaker protection.
func (s *BreakerService) Patch(ctx context.Context, url string, payload any, opts ...Option) (*http.Response, error) {
	var resp *http.Response
	err := s.brk.DoWithAcceptable(func() error {
		var e error
		resp, e = s.inner.Patch(ctx, url, payload, opts...)
		return e
	}, acceptable)
	return resp, err
}

// Delete implements Service.Delete with circuit breaker protection.
func (s *BreakerService) Delete(ctx context.Context, url string, payload any, opts ...Option) (*http.Response, error) {
	var resp *http.Response
	err := s.brk.DoWithAcceptable(func() error {
		var e error
		resp, e = s.inner.Delete(ctx, url, payload, opts...)
		return e
	}, acceptable)
	return resp, err
}

// PostForm implements Service.PostForm with circuit breaker protection.
func (s *BreakerService) PostForm(ctx context.Context, url string, payload map[string]string, opts ...Option) (*http.Response, error) {
	var resp *http.Response
	err := s.brk.DoWithAcceptable(func() error {
		var e error
		resp, e = s.inner.PostForm(ctx, url, payload, opts...)
		return e
	}, acceptable)
	return resp, err
}

// Do implements httpc.Service.Do with circuit breaker protection.
func (s *BreakerService) Do(ctx context.Context, method, url string, data any) (*http.Response, error) {
	var resp *http.Response
	err := s.brk.DoWithAcceptable(func() error {
		var e error
		resp, e = s.inner.Do(ctx, method, url, data)
		return e
	}, acceptable)
	return resp, err
}

// DoRequest implements httpc.Service.DoRequest with circuit breaker protection.
func (s *BreakerService) DoRequest(r *http.Request) (*http.Response, error) {
	var resp *http.Response
	err := s.brk.DoWithAcceptable(func() error {
		var e error
		resp, e = s.inner.DoRequest(r)
		return e
	}, acceptable)
	return resp, err
}

// Ensure BreakerService implements both Service and httpc.Service
var _ Service = (*BreakerService)(nil)
var _ httpc.Service = (*BreakerService)(nil)
