package registry

import (
	"pulse/internal/contract"

	"gorm.io/gorm"
)

// RepositoryContext exposes all persistence interfaces.
// All getters return nil until TASK-011+ wires concrete implementations.
type RepositoryContext interface {
	GetLogRepo() contract.LogRepo
	GetIncidentRepo() contract.IncidentRepo
	GetEmbeddingRepo() contract.EmbeddingRepo
	GetLLMCallRepo() contract.LLMCallRepo
	GetTenantRepo() contract.TenantRepo
	GetAPIKeyRepo() contract.APIKeyRepo
	GetOutboxRepo() contract.OutboxRepo
	GetSagaRepo() contract.SagaRepo
	GetAuditLogRepo() contract.AuditLogRepo
	GetTxManager() contract.TxManager
	GetOutboxAppender() contract.OutboxAppender
}

type repositoryContext struct {
	db *gorm.DB
}

// NewRepositoryContext creates the context from an open GORM connection.
// Replace nil returns with concrete repo constructors as each is implemented.
func NewRepositoryContext(db *gorm.DB) RepositoryContext {
	return &repositoryContext{db: db}
}

func (r *repositoryContext) GetLogRepo() contract.LogRepo              { return nil }
func (r *repositoryContext) GetIncidentRepo() contract.IncidentRepo     { return nil }
func (r *repositoryContext) GetEmbeddingRepo() contract.EmbeddingRepo   { return nil }
func (r *repositoryContext) GetLLMCallRepo() contract.LLMCallRepo       { return nil }
func (r *repositoryContext) GetTenantRepo() contract.TenantRepo         { return nil }
func (r *repositoryContext) GetAPIKeyRepo() contract.APIKeyRepo         { return nil }
func (r *repositoryContext) GetOutboxRepo() contract.OutboxRepo         { return nil }
func (r *repositoryContext) GetSagaRepo() contract.SagaRepo             { return nil }
func (r *repositoryContext) GetAuditLogRepo() contract.AuditLogRepo     { return nil }
func (r *repositoryContext) GetTxManager() contract.TxManager           { return nil }
func (r *repositoryContext) GetOutboxAppender() contract.OutboxAppender { return nil }
