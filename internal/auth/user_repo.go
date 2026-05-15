package auth

import (
	"context"
	"fmt"

	"pulse/internal/contract"
	"pulse/internal/types/entity"

	"gorm.io/gorm"
)

var _ contract.UserRepo = (*UserRepository)(nil)

// UserRepository implements contract.UserRepo using GORM.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return &u, nil
}
