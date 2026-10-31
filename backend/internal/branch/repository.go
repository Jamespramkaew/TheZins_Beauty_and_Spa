package branch

import (
	"context"
	"errors"

	"backend/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindBranchById(ctx context.Context, id string) (*models.Branch, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindBranchById ค้นหาสาขาตาม ID
// ไม่เจอ = (nil, nil) เหมือน stock.FindStockById ผู้เรียกต้องเช็ค nil เอง
// คืนสาขาที่ปิดใช้งานด้วย (is_active = false) ผู้เรียกเป็นคนตัดสินว่ายอมรับไหม
func (r *repository) FindBranchById(ctx context.Context, id string) (*models.Branch, error) {
	var branch models.Branch
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&branch).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &branch, nil
}
