package enum

// IncidentStatus maps to the Postgres incident_status ENUM type.
type IncidentStatus string

const (
	IncidentStatusOpen          IncidentStatus = "open"
	IncidentStatusInvestigating IncidentStatus = "investigating"
	IncidentStatusResolving     IncidentStatus = "resolving"
	IncidentStatusResolved      IncidentStatus = "resolved"
	IncidentStatusClosed        IncidentStatus = "closed"
)

// Severity represents incident/anomaly severity levels (maps to SMALLINT).
type Severity int

const (
	SeverityLow      Severity = 1
	SeverityMedium   Severity = 2
	SeverityHigh     Severity = 3
	SeverityCritical Severity = 4
)

// ResolutionCategory classifies the root cause of an incident.
type ResolutionCategory int

const (
	ResolutionCategoryUnknown  ResolutionCategory = 0
	ResolutionCategoryCodeBug  ResolutionCategory = 1
	ResolutionCategoryConfig   ResolutionCategory = 2
	ResolutionCategoryInfra    ResolutionCategory = 3
	ResolutionCategoryExternal ResolutionCategory = 4
)

// OutboxStatus maps to outbox_events.status (SMALLINT).
type OutboxStatus int16

const (
	OutboxStatusPending   OutboxStatus = 0
	OutboxStatusPublished OutboxStatus = 1
	OutboxStatusFailed    OutboxStatus = 2
)

// SagaStatus maps to saga_instances.status (SMALLINT).
type SagaStatus int16

const (
	SagaStatusStarted      SagaStatus = 0
	SagaStatusCompleted    SagaStatus = 1
	SagaStatusCompensating SagaStatus = 2
	SagaStatusFailed       SagaStatus = 3
)
