package httpc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// RetryConfig defines the configuration for HTTP request retry logic.
type RetryConfig struct {
	MaxRetries     int           // Maximum number of retry attempts
	InitialBackoff time.Duration // Initial backoff duration
	MaxBackoff     time.Duration // Maximum backoff duration
	Multiplier     float64       // Backoff multiplier for exponential backoff
}

// DefaultRetryConfig provides default retry configuration:
// - 3 retries
// - 100ms initial backoff
// - 2s maximum backoff
// - 2.0 multiplier (exponential backoff)
var DefaultRetryConfig = RetryConfig{
	MaxRetries:     3,
	InitialBackoff: 100 * time.Millisecond,
	MaxBackoff:     2 * time.Second,
	Multiplier:     2.0,
}

// WithRetry wraps an HTTP request with exponential backoff retry logic.
// It retries on network errors and 5xx server errors.
// 4xx client errors are not retried as they indicate bad requests.
//
// Example usage:
//
//	resp, err := httpc.WithRetry(ctx, func(ctx context.Context) (*http.Response, error) {
//	    return httpClient.Get(ctx, "https://api.example.com/data")
//	}, httpc.DefaultRetryConfig)
func WithRetry(ctx context.Context, fn func(context.Context) (*http.Response, error), cfg RetryConfig) (*http.Response, error) {
	var lastErr error
	backoff := cfg.InitialBackoff

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Wait before retry (skip on first attempt)
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}

			// Calculate next backoff with exponential growth
			backoff = time.Duration(float64(backoff) * cfg.Multiplier)
			if backoff > cfg.MaxBackoff {
				backoff = cfg.MaxBackoff
			}
		}

		// Attempt the request
		resp, err := fn(ctx)

		// Success case: no error and status < 500
		if err == nil {
			if resp.StatusCode < 500 {
				return resp, nil
			}

			// 5xx error - close body and retry
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: server error", resp.StatusCode)
		} else {
			// Network error or other failure
			lastErr = err
		}

		// Log retry attempt
		if attempt < cfg.MaxRetries {
			logx.WithContext(ctx).Infof(
				"HTTP request failed (attempt %d/%d), retrying in %v: %v",
				attempt+1, cfg.MaxRetries+1, backoff, lastErr,
			)
		}
	}

	// All retries exhausted
	return nil, fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}
