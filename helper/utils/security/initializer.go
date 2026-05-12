package security

import (
	"context"
	"os"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// PermissionInitializer implements the core.Service interface to initialize
// permission cache at service startup with retry logic.
type PermissionInitializer struct {
	provider PermissionProvider
}

// NewPermissionInitializer creates a new permission initializer service.
//
// This initializer will be called during service startup (via the Providers() pattern)
// to pre-load the permission cache. It implements exponential backoff retry logic
// to handle temporary permission service outages.
func NewPermissionInitializer(provider PermissionProvider) *PermissionInitializer {
	return &PermissionInitializer{
		provider: provider,
	}
}

// Register implements the core.Service interface.
// It is called during service initialization to pre-load the permission cache.
//
// Retry behavior:
//   - Max retries: 5 attempts
//   - Backoff: Exponential (1s, 2s, 4s, 8s, 16s)
//   - Total time: ~31 seconds
//   - On failure: Exits with error code 1 (fail-fast)
func (p *PermissionInitializer) Register() {
	ctx := context.Background()
	maxRetries := 5

	logx.Info("Initializing permission cache...")

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := p.provider.Refresh(ctx)
		if err == nil {
			logx.Info("Permission cache initialized successfully")
			return
		}

		// Calculate exponential backoff
		backoff := time.Duration(1<<attempt) * time.Second

		if attempt < maxRetries-1 {
			logx.Errorf("Failed to initialize permissions (attempt %d/%d): %v, retrying in %v", attempt+1, maxRetries, err, backoff)
			time.Sleep(backoff)
		} else {
			// Final attempt failed
			logx.Severef("Failed to initialize permissions after %d attempts: %v", maxRetries, err)
			logx.Severe("Permission service is required for startup. Exiting.")
			os.Exit(1)
		}
	}
}
