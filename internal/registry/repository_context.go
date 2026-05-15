package registry

import (
	"pulse/internal/auth"
	"pulse/internal/contract"
	"pulse/internal/outbox"

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
	GetUserRepo() contract.UserRepo
	GetAPIKeyRepo() contract.APIKeyRepo
	GetOutboxRepo() contract.OutboxRepo
	GetSagaRepo() contract.SagaRepo
	GetAuditLogRepo() contract.AuditLogRepo
	GetTxManager() contract.TxManager
	GetOutboxAppender() contract.OutboxAppender
}

type repositoryContext struct {
	db           *gorm.DB
	outboxRepo   *outbox.Repository
	tenantRepo   contract.TenantRepo
	userRepo     contract.UserRepo
	apiKeyRepo   contract.APIKeyRepo
	auditLogRepo contract.AuditLogRepo
}

// NewRepositoryContext creates the context from an open GORM connection.
// Replace nil returns with concrete repo constructors as each is implemented.
func NewRepositoryContext(db *gorm.DB) RepositoryContext {
	return &repositoryContext{
		db:           db,
		outboxRepo:   outbox.NewRepository(db),
		tenantRepo:   auth.NewTenantRepository(db),
		userRepo:     auth.NewUserRepository(db),
		apiKeyRepo:   auth.NewAPIKeyRepository(db),
		auditLogRepo: auth.NewAuditLogRepository(db),
	}
}

func (r *repositoryContext) GetLogRepo() contract.LogRepo              { return nil }
func (r *repositoryContext) GetIncidentRepo() contract.IncidentRepo     { return nil }
func (r *repositoryContext) GetEmbeddingRepo() contract.EmbeddingRepo   { return nil }
func (r *repositoryContext) GetLLMCallRepo() contract.LLMCallRepo       { return nil }
func (r *repositoryContext) GetTenantRepo() contract.TenantRepo         { return r.tenantRepo }
func (r *repositoryContext) GetUserRepo() contract.UserRepo             { return r.userRepo }
func (r *repositoryContext) GetAPIKeyRepo() contract.APIKeyRepo         { return r.apiKeyRepo }
func (r *repositoryContext) GetOutboxRepo() contract.OutboxRepo         { return r.outboxRepo }
func (r *repositoryContext) GetSagaRepo() contract.SagaRepo             { return nil }
func (r *repositoryContext) GetAuditLogRepo() contract.AuditLogRepo     { return r.auditLogRepo }
func (r *repositoryContext) GetTxManager() contract.TxManager           { return nil }
func (r *repositoryContext) GetOutboxAppender() contract.OutboxAppender { return r.outboxRepo }
