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
	GetStockQuantities(ctx context.Context, f QuantityFilter) ([]StockQuantity, error)
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

// GetStockQuantities คืนสินค้าพร้อมจำนวนคงเหลือ = SUM(qty_change) จาก stock_logs
// f.BranchID ว่าง = รวมทุกสาขา (service เป็นคนตัดสินว่าใครใช้แบบนี้ได้)
// ใช้ LEFT JOIN สินค้าที่ยังไม่มี log จะยังโผล่มาด้วยยอด 0
func (r *repository) GetStockQuantities(ctx context.Context, f QuantityFilter) ([]StockQuantity, error) {
	// เงื่อนไขสาขาต้องอยู่ใน ON ถ้าไปอยู่ใน WHERE จะตัดสินค้าที่ไม่มี log ทิ้ง
	joinSQL := "LEFT JOIN stock_logs l ON l.stock_id = s.id"
	var joinArgs []any
	if f.BranchID != "" {
		joinSQL += " AND l.branch_id = ?"
		joinArgs = append(joinArgs, f.BranchID)
	}

	q := r.db.WithContext(ctx).
		Table("stocks AS s").
		Select(`s.id AS stock_id, s.name, s.price, s.min_qty,
			COALESCE(SUM(l.qty_change), 0) AS quantity`).
		Joins(joinSQL, joinArgs...).
		Where("s.deleted_at IS NULL")

	if f.Search != "" {
		q = q.Where("s.name ILIKE ? ESCAPE '\\'", "%"+escapeLike(f.Search)+"%")
	}
	if f.Category != "" {
		q = q.Where("s.category = ?", f.Category)
	}

	q = q.Group("s.id")
	if f.LowStock {
		q = q.Having("COALESCE(SUM(l.qty_change), 0) <= s.min_qty")
	}

	var rows []StockQuantity
	if err := q.Order("s.name ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// escapeLike กัน % และ _ ที่ user พิมพ์มาถูกตีความเป็น wildcard
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
