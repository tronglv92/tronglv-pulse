package contract

import (
	"context"

	"pulse/internal/types/entity"
)

// Step is a single compensatable unit in a saga.
type Step interface {
	// Name returns a stable identifier stored in saga_instances.current_step.
	Name() string
	// Execute runs the forward action. A non-nil error triggers compensation.
	Execute(ctx context.Context, payload []byte) error
	// Compensate undoes the forward action. Called in reverse step order on failure.
	Compensate(ctx context.Context, payload []byte) error
	// CanSkip returns true when Execute is a safe no-op (idempotent re-run guard).
	CanSkip(ctx context.Context, payload []byte) (bool, error)
}

// Saga is an ordered sequence of Steps that executes as a distributed transaction.
type Saga interface {
	// Name returns the saga type string stored in saga_instances.saga_type.
	Name() string
	Steps() []Step
}

// SagaRepository is the orchestrator's persistence view — richer than SagaRepo.
// SagaRepo (repository.go) is the registry-level CRUD; this is the saga-executor view.
type SagaRepository interface {
	Load(ctx context.Context, id int64) (*entity.SagaInstance, error)
	Save(ctx context.Context, e *entity.SagaInstance) error
	// AdvanceStep records successful execution of a step.
	AdvanceStep(ctx context.Context, id int64, stepName string) error
	// BeginCompensation transitions the saga to compensating status and records the error.
	BeginCompensation(ctx context.Context, id int64, failedStep string, errMsg string) error
}
