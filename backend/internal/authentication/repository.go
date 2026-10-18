package authentication

import (
	"backend/internal/models"
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	UpdatePassword(ctx context.Context, userId string, newHashPassword string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repository) UpdatePassword(ctx context.Context, username string, newHashPassword string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("username = ?", username).
		Update("password", newHashPassword).Error
}
