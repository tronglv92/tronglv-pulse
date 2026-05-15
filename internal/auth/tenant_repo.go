package auth

import (
	"context"
	"fmt"

	"pulse/internal/contract"
	"pulse/internal/types/entity"

	"gorm.io/gorm"
)

var _ contract.TenantRepo = (*TenantRepository)(nil)

// TenantRepository implements contract.TenantRepo using GORM.
type TenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository creates a new TenantRepository.
func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Save(ctx context.Context, e *entity.Tenant) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return fmt.Errorf("failed to save tenant: %w", err)
	}
	return nil
}

func (r *TenantRepository) FindByID(ctx context.Context, id int64) (*entity.Tenant, error) {
	var t entity.Tenant
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error; err != nil {
		return nil, fmt.Errorf("failed to find tenant by id: %w", err)
	}
	return &t, nil
}

func (r *TenantRepository) FindBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	var t entity.Tenant
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&t).Error; err != nil {
		return nil, fmt.Errorf("failed to find tenant by slug: %w", err)
	}
	return &t, nil
}

func (r *TenantRepository) Update(ctx context.Context, e *entity.Tenant) error {
	if err := r.db.WithContext(ctx).Save(e).Error; err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	return nil
}
