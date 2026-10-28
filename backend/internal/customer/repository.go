package customer

import (
	"context"
	"time"

	"backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	Search(ctx context.Context, q string, page, pageSize int) ([]models.Customer, int64, error)
	GetByID(ctx context.Context, id string) (*models.Customer, error)
	Create(ctx context.Context, customer *models.Customer) error
	Update(ctx context.Context, id string, fields map[string]any) error
	ChangeTier(ctx context.Context, customerID, newTier, changedBy string) (*models.Customer, error)
	ListTierChanges(ctx context.Context, customerID string) ([]TierChangeRow, error)
}

// TierChangeRow 1 แถวในประวัติการเปลี่ยนระดับ + username ของบัญชีที่เปลี่ยน
type TierChangeRow struct {
	ID                string    `json:"id"`
	NewTier           string    `json:"new_tier"`
	ChangedBy         string    `json:"changed_by"`
	ChangedByUsername string    `json:"changed_by_username"`
	ChangedAt         time.Time `json:"changed_at"`
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Search ค้นลูกค้าจาก HN, ชื่อเต็ม หรือเบอร์โทร (ไม่สนตัวพิมพ์เล็กใหญ่) ทีละหน้า
// q ว่าง = คืนลูกค้าที่สร้างล่าสุด · คืนจำนวนที่ตรงเงื่อนไขทั้งหมด (total) ไว้ทำ meta
func (r *repository) Search(ctx context.Context, q string, page, pageSize int) ([]models.Customer, int64, error) {
	filtered := r.db.WithContext(ctx).Model(&models.Customer{})
	if q != "" {
		like := "%" + q + "%"
		filtered = filtered.Where(
			"hn_no ILIKE ? OR CONCAT(fname, ' ', lname) ILIKE ? OR tel ILIKE ?",
			like, like, like,
		)
	}

	// Session ทำให้ใช้ query เดียวกันซ้ำได้ทั้งตอน Count และ Find
	filtered = filtered.Session(&gorm.Session{})

	var total int64
	if err := filtered.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var customers []models.Customer
	// id เป็นตัวตัดสินเมื่อ created_at เท่ากัน ไม่งั้นข้ามหน้าแล้วแถวอาจซ้ำหรือตกหล่น
	err := filtered.
		Order("created_at DESC, id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&customers).Error
	if err != nil {
		return nil, 0, err
	}
	return customers, total, nil
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

// ChangeTier เปลี่ยนระดับลูกค้าและบันทึกประวัติใน transaction เดียว
// ล็อกแถวลูกค้า (SELECT ... FOR UPDATE) ถ้ามีคนเปลี่ยนพร้อมกัน อีกคนต้องรอแล้วจะเห็นระดับล่าสุด
func (r *repository) ChangeTier(ctx context.Context, customerID, newTier, changedBy string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&customer, "id = ?", customerID).Error; err != nil {
			return err
		}

		// เช็คตรงนี้ (ใต้การล็อก) ไม่ใช่ใน service เพราะต้องเป็นระดับล่าสุด ณ ตอนบันทึก
		if customer.Tier == newTier {
			return invalid("customer is already %s", newTier)
		}

		if err := tx.Model(&customer).Update("tier", newTier).Error; err != nil {
			return err
		}
		customer.Tier = newTier // ให้ค่าที่ตอบกลับเป็นระดับใหม่

		return tx.Create(&models.TierChange{
			CustomerID: customerID,
			NewTier:    newTier,
			ChangedBy:  changedBy,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// ListTierChanges ประวัติการเปลี่ยนระดับของลูกค้าคนนี้ ใหม่ → เก่า
func (r *repository) ListTierChanges(ctx context.Context, customerID string) ([]TierChangeRow, error) {
	rows := []TierChangeRow{}
	err := r.db.WithContext(ctx).Table("tier_changes AS t").
		Select("t.id, t.new_tier, t.changed_by, u.username AS changed_by_username, t.changed_at").
		Joins("JOIN users AS u ON u.id = t.changed_by").
		Where("t.customer_id = ?", customerID).
		Order("t.changed_at DESC").
		Scan(&rows).Error
	return rows, err
}
