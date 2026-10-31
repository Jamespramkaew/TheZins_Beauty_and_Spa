package stock

import (
	"context"
	"strings"

	"backend/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetAllStock(ctx context.Context) ([]models.Stock, error)
	GetStockContainersByStockID(ctx context.Context, stockId string, branchId string, status string, page int, pageSize int) ([]models.StockContainer, int64, error)
	FindStockByName(ctx context.Context, name string) (*models.Stock, error)
	FindStockById(ctx context.Context, id string) (*models.Stock, error)
	CreateStock(ctx context.Context, stock models.Stock) (models.Stock, error)
	UpdateStock(ctx context.Context, stock models.Stock) (models.Stock, error)
	CheckStockHasRelation(ctx context.Context, stockId string) (bool, error)
	SoftDeleteStock(ctx context.Context, stockId string) error
	HardDeleteStock(ctx context.Context, stockId string) error
	FindStockByIdIncludingDeleted(ctx context.Context, id string) (*models.Stock, error)
	RestoreStock(ctx context.Context, stockId string) (models.Stock, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAllStock(ctx context.Context) ([]models.Stock, error) {
	var stocks []models.Stock
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&stocks).Error; err != nil {
		return nil, err
	}
	return stocks, nil
}

// GetStockContainersByStockID คืนกระปุกของสินค้า 1 ชนิดทีละหน้า เรียงจากเปิดล่าสุดก่อน
// พร้อมจำนวนแถวทั้งหมดที่ตรงเงื่อนไข (ไม่นับ limit/offset) เอาไปคำนวณจำนวนหน้า
// branchId ว่าง = ทุกสาขา (service เป็นคนตัดสินว่าใครใช้แบบนี้ได้)
// status, page, pageSize ต้องผ่านการตรวจจาก service แล้ว
func (r *repository) GetStockContainersByStockID(ctx context.Context, stockId string, branchId string, status string, page int, pageSize int) ([]models.StockContainer, int64, error) {
	filtered := r.db.WithContext(ctx).
		Model(&models.StockContainer{}).
		Where("stock_id = ?", stockId)

	if branchId != "" {
		filtered = filtered.Where("branch_id = ?", branchId)
	}

	switch status {
	case StatusOpen:
		filtered = filtered.Where("closed_at IS NULL")
	case StatusClosed:
		filtered = filtered.Where("closed_at IS NOT NULL")
	}

	// Session ทำให้ใช้ query เดียวกันซ้ำได้ทั้งตอน Count และ Find (ไม่งั้นเงื่อนไขของ Count จะติดไปด้วย)
	filtered = filtered.Session(&gorm.Session{})

	var total int64
	if err := filtered.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var containers []models.StockContainer
	// id เป็นตัวตัดสินเมื่อ opened_at เท่ากัน ไม่งั้นข้ามหน้าแล้วแถวอาจซ้ำหรือตกหล่น
	err := filtered.
		Preload("Branch").
		Order("opened_at DESC, id ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&containers).Error
	if err != nil {
		return nil, 0, err
	}
	return containers, total, nil
}

// escapeLike กัน % และ _ ที่ user พิมพ์มาถูกตีความเป็น wildcard
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// FindStockByName ค้นหาสินค้าตามชื่อ (รวมทั้งที่ถูกลบด้วย soft delete)
func (r *repository) FindStockByName(ctx context.Context, name string) (*models.Stock, error) {
	var stock models.Stock
	// เรียงตัวที่ยังไม่ถูกลบขึ้นก่อน ไม่งั้นถ้ามีชื่อเดียวกันทั้งที่ลบแล้วและยังใช้งานอยู่ อาจได้ตัวที่ลบแล้วมาแทน
	err := r.db.WithContext(ctx).Unscoped().Where("name = ?", name).
		Order("(deleted_at IS NULL) DESC").First(&stock).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &stock, nil
}

// CreateStock สร้างสินค้าใหม่
func (r *repository) CreateStock(ctx context.Context, stock models.Stock) (models.Stock, error) {
	if err := r.db.WithContext(ctx).Create(&stock).Error; err != nil {
		return models.Stock{}, err
	}
	return stock, nil
}

// FindStockById ค้นหาสินค้าตาม ID
func (r *repository) FindStockById(ctx context.Context, id string) (*models.Stock, error) {
	var stock models.Stock
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&stock).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &stock, nil
}

// UpdateStock อัปเดตข้อมูลสินค้า
func (r *repository) UpdateStock(ctx context.Context, stock models.Stock) (models.Stock, error) {
	if err := r.db.WithContext(ctx).Save(&stock).Error; err != nil {
		return models.Stock{}, err
	}
	return stock, nil
}

// CheckStockHasRelation เช็คว่า Stock ถูกอ้างอิงจากตารางอื่นหรือไม่
// ต้องครบทุกตารางที่มี FK stock_id ชี้มาที่ stocks ไม่งั้น hard delete จะติด FK แล้วตอบ 500
// Relations: OrderItem, CourseItem, StockLog, StockContainer
func (r *repository) CheckStockHasRelation(ctx context.Context, stockId string) (bool, error) {
	relatedModels := []any{
		&models.OrderItem{}, // เคยขายแล้ว (V1 ยังไม่ตัดสต็อกตอนขาย จึงอาจไม่มี stock_logs)
		&models.CourseItem{},
		&models.StockLog{},
		&models.StockContainer{},
	}

	for _, m := range relatedModels {
		var count int64
		err := r.db.WithContext(ctx).Model(m).Where("stock_id = ?", stockId).Count(&count).Error
		if err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}

	return false, nil
}

// SoftDeleteStock ทำ Soft Delete (set DeletedAt)
func (r *repository) SoftDeleteStock(ctx context.Context, stockId string) error {
	return r.db.WithContext(ctx).Delete(&models.Stock{}, "id = ?", stockId).Error
}

// HardDeleteStock ลบ Stock ออกจากฐานข้อมูลจริงๆ
func (r *repository) HardDeleteStock(ctx context.Context, stockId string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&models.Stock{}, "id = ?", stockId).Error
}

// FindStockByIdIncludingDeleted ค้นหาสินค้าตาม ID รวมที่ถูก soft delete
func (r *repository) FindStockByIdIncludingDeleted(ctx context.Context, id string) (*models.Stock, error) {
	var stock models.Stock
	err := r.db.WithContext(ctx).Unscoped().Where("id = ?", id).First(&stock).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &stock, nil
}

// RestoreStock กู้คืนสินค้าที่ถูก soft delete (เซ็ต deleted_at เป็น NULL) แล้วคืนข้อมูลล่าสุด
func (r *repository) RestoreStock(ctx context.Context, stockId string) (models.Stock, error) {
	var stock models.Stock
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&models.Stock{}).
			Where("id = ? AND deleted_at IS NOT NULL", stockId).
			Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return tx.First(&stock, "id = ?", stockId).Error
	})
	if err != nil {
		return models.Stock{}, err
	}
	return stock, nil
}
