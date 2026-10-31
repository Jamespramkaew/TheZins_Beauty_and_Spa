package employee

import (
	"context"
	"errors"

	"backend/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindEmployeeById(ctx context.Context, id string) (*models.Employee, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindEmployeeById ค้นหาพนักงานตาม ID
// ไม่เจอ = (nil, nil) เหมือน stock.FindStockById ผู้เรียกต้องเช็ค nil เอง
// คืนพนักงานที่ไม่ active ด้วย ผู้เรียกต้องเช็ค IsActive และ BranchID เอง
// (เช่น stock-log ต้องเป็นพนักงาน active ของสาขาเดียวกับแถว log)
func (r *repository) FindEmployeeById(ctx context.Context, id string) (*models.Employee, error) {
	var employee models.Employee
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&employee).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &employee, nil
}
