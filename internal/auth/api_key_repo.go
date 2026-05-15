package auth

import (
	"context"
	"fmt"
	"time"

	"pulse/internal/contract"
	"pulse/internal/types/entity"

	"gorm.io/gorm"
)

var _ contract.APIKeyRepo = (*APIKeyRepository)(nil)

// APIKeyRepository implements contract.APIKeyRepo using GORM.
type APIKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new APIKeyRepository.
func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Save(ctx context.Context, e *entity.APIKey) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return fmt.Errorf("failed to save api key: %w", err)
	}
	return nil
}

func (r *APIKeyRepository) FindByHash(ctx context.Context, keyHash string) (*entity.APIKey, error) {
	var k entity.APIKey
	if err := r.db.WithContext(ctx).Where("key_hash = ?", keyHash).First(&k).Error; err != nil {
		return nil, fmt.Errorf("failed to find api key by hash: %w", err)
	}
	return &k, nil
}

func (r *APIKeyRepository) ListByTenant(ctx context.Context, tenantID int64) ([]*entity.APIKey, error) {
	var keys []*entity.APIKey
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("failed to list api keys by tenant: %w", err)
	}
	return keys, nil
}

func (r *APIKeyRepository) Delete(ctx context.Context, id, tenantID int64) error {
	result := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.APIKey{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete api key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *APIKeyRepository) TouchLastUsed(ctx context.Context, id int64) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&entity.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error; err != nil {
		return fmt.Errorf("failed to touch api key last_used_at: %w", err)
	}
	return nil
}
