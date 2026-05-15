package auth

import (
	"context"
	"fmt"

	"pulse/internal/contract"
	"pulse/internal/types/entity"

	"gorm.io/gorm"
)

var _ contract.AuditLogRepo = (*AuditLogRepository)(nil)

// AuditLogRepository implements contract.AuditLogRepo using GORM.
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new AuditLogRepository.
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Append(ctx context.Context, e *entity.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return fmt.Errorf("failed to append audit log: %w", err)
	}
	return nil
}

func (r *AuditLogRepository) FindByResource(ctx context.Context, tenantID int64, resource string, resourceID int64) ([]*entity.AuditLog, error) {
	var logs []*entity.AuditLog
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND resource = ? AND resource_id = ?", tenantID, resource, resourceID).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to find audit logs by resource: %w", err)
	}
	return logs, nil
}
