package example

import (
	"context"
	"backend/database"
	"gorm.io/gorm"
)

type Repository interface {
	GetAllExampleMessage(ctx context.Context, limit, offset int) ([]database.Example, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetAllExampleMessage(ctx context.Context, limit, offset int) ([]database.Example, int64, error) {
	var example []database.Example
	var totalCount int64

	if err := r.db.WithContext(ctx).
		Model(&database.Example{}).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&example).Error

	if err != nil {
		return nil, 0, err
	}

	return example, totalCount, nil
}