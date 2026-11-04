package employee

import (
	"context"
	"errors"

	"backend/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindEmployeeById(ctx context.Context, id string) (*models.Employee, error)
	GetActiveEmployees(ctx context.Context, branchId string) ([]models.Employee, error)
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

// GetActiveEmployees คืนพนักงานที่ active เรียงตามชื่อ
// branchId ว่าง = ทุกสาขา
func (r *repository) GetActiveEmployees(ctx context.Context, branchId string) ([]models.Employee, error) {
	employees := []models.Employee{}
	q := r.db.WithContext(ctx).Where("is_active = TRUE")
	if branchId != "" {
		q = q.Where("branch_id = ?", branchId)
	}
	if err := q.Order("name ASC, id ASC").Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}
