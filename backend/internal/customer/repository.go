package customer

import (
	"context"

	"backend/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	Search(ctx context.Context, q string, limit int) ([]models.Customer, error)
	GetByID(ctx context.Context, id string) (*models.Customer, error)
	Create(ctx context.Context, customer *models.Customer) error
	Update(ctx context.Context, id string, fields map[string]any) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Search ค้นลูกค้าจาก HN, ชื่อเต็ม หรือเบอร์โทร (ไม่สนตัวพิมพ์เล็กใหญ่)
// q ว่าง = คืนลูกค้าที่สร้างล่าสุด
func (r *repository) Search(ctx context.Context, q string, limit int) ([]models.Customer, error) {
	var customers []models.Customer

	query := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit)
	if q != "" {
		like := "%" + q + "%"
		query = query.Where(
			"hn_no ILIKE ? OR CONCAT(fname, ' ', lname) ILIKE ? OR tel ILIKE ?",
			like, like, like,
		)
	}

	if err := query.Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*models.Customer, error) {
	var customer models.Customer
	if err := r.db.WithContext(ctx).First(&customer, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *repository) Create(ctx context.Context, customer *models.Customer) error {
	return r.db.WithContext(ctx).Create(customer).Error
}

// Update แก้เฉพาะคอลัมน์ที่อยู่ใน fields
// ใช้ map แทน struct เพื่อให้แก้เป็น "" หรือ NULL ได้
func (r *repository) Update(ctx context.Context, id string, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&models.Customer{}).Where("id = ?", id).Updates(fields).Error
}
