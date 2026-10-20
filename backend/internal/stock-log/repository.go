package stocklog

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

type Repository interface {
	GetAllStockQuantities(ctx context.Context, f quantityFilter) ([]StockQuantity, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// เงื่อนไขที่ผ่านการตรวจจาก service แล้ว
type quantityFilter struct {
	BranchID string // ว่าง = ทุกสาขาที่ยังเปิดใช้งาน
	Search   string
	Category string // ว่าง = ทั้งหมด
	LowStock bool
	Active   string // activeTrue | activeFalse | activeAll
	Page     int
	PageSize int
}

// escapeLike กัน % และ _ ที่ user พิมพ์มาถูกตีความเป็น wildcard
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// GetAllStockQuantities คืนยอดคงเหลือของสินค้าแต่ละชนิดแยกตามสาขา ทีละหน้า
// ใช้ stocks x branches แล้ว LEFT JOIN stock_logs เพื่อให้สินค้าที่ไม่เคยมี log แสดงยอด 0
func (r *repository) GetAllStockQuantities(ctx context.Context, f quantityFilter) ([]StockQuantity, int64, error) {
	qty := "COALESCE(SUM(l.qty_change), 0)"

	base := r.db.WithContext(ctx).
		Table("stocks AS s").
		Joins("CROSS JOIN branches AS b").
		Joins("LEFT JOIN stock_logs AS l ON l.stock_id = s.id AND l.branch_id = b.id")

	if f.BranchID != "" {
		base = base.Where("b.id = ?", f.BranchID)
	} else {
		base = base.Where("b.is_active = TRUE")
	}

	switch f.Active {
	case activeTrue:
		base = base.Where("s.deleted_at IS NULL AND s.is_active = TRUE")
	case activeFalse:
		base = base.Where("s.deleted_at IS NULL AND s.is_active = FALSE")
	}

	if f.Category != "" {
		base = base.Where("s.category = ?", f.Category)
	}
	if f.Search != "" {
		base = base.Where("s.name ILIKE ?", "%"+escapeLike(f.Search)+"%")
	}

	base = base.Group("s.id, b.id")
	if f.LowStock {
		base = base.Having(qty + " <= s.min_qty")
	}

	// ชุดผลลัพธ์ก่อนแบ่งหน้า
	rows := base.Select(`
		s.id AS stock_id,
		s.name AS stock_name,
		s.category AS category,
		s.uses_per_unit AS uses_per_unit,
		s.min_qty AS min_qty,
		s.is_active AS is_active,
		b.id AS branch_id,
		b.name AS branch_name,
		` + qty + ` AS quantity`)

	var total int64
	if err := r.db.WithContext(ctx).Table("(?) AS t", rows).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var result []StockQuantity

	err := rows.
		Order("s.name ASC, b.name ASC, s.id ASC, b.id ASC").
		Limit(f.PageSize).
		Offset((f.Page - 1) * f.PageSize).
		Scan(&result).Error
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}
