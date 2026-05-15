package service

import (
	"pulse/helper/utils/toolkit/downloader"
	"pulse/internal/contract"
)

// ServiceFactoryContext is the subset of registry.ServiceContext / registry.CronContext
// that the ServiceFactory needs. Using this narrow interface prevents an import cycle.
type ServiceFactoryContext interface {
	GetDownloader() downloader.Downloader
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

// ServiceFactory vends service instances via lazy oncex initialization.
// Run `pmctl gen service` after adding or removing a service.
type ServiceFactory struct {
	ctx ServiceFactoryContext
}

func NewServiceFactory(ctx ServiceFactoryContext) *ServiceFactory {
	return &ServiceFactory{ctx: ctx}
}
