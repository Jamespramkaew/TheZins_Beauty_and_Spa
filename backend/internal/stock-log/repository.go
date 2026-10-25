package stocklog

import (
	"backend/internal/models"
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	GetAllStockQuantities(ctx context.Context, f quantityFilter) ([]StockQuantity, int64, error)
	GetStockLogs(ctx context.Context, f stockLogFilter) ([]models.StockLog, int64, error)
	CreateStockLog(ctx context.Context, log *models.StockLog) error
	TransferStockLog(ctx context.Context, data TransferStockLogRequest, logType string) error
	GetStockQuantity(ctx context.Context, stockId string, branchId string) (int, error)
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
		` + qty + ` AS quantity`).Session(&gorm.Session{}) // ใช้ซ้ำได้ทั้งตอนนับและตอนดึงข้อมูล โดยเงื่อนไขของอีกฝั่งไม่ติดมา

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

// stockLogFilter = เงื่อนไขของ GET /api/stock-logs ที่ผ่านการตรวจจาก service แล้ว
// ค่าว่าง/nil = ไม่กรองช่องนั้น
type stockLogFilter struct {
	StockID     string
	BranchID    string
	LogType     string
	Direction   string     // directionIn | directionOut | ว่าง
	From        *time.Time // created_at >= From
	Until       *time.Time // created_at < Until (service บวกให้ครบวันแล้ว)
	OrderItemID string
	UsageID     string
	Page        int
	PageSize    int
}

// GetStockLogs คืนประวัติ stock_logs ทีละหน้า เรียงใหม่สุดก่อน พร้อมจำนวนแถวทั้งหมดที่ตรงเงื่อนไข
func (r *repository) GetStockLogs(ctx context.Context, f stockLogFilter) ([]models.StockLog, int64, error) {
	filtered := r.db.WithContext(ctx).Model(&models.StockLog{})

	if f.StockID != "" {
		filtered = filtered.Where("stock_id = ?", f.StockID)
	}
	if f.BranchID != "" {
		filtered = filtered.Where("branch_id = ?", f.BranchID)
	}
	if f.LogType != "" {
		filtered = filtered.Where("log_type = ?", f.LogType)
	}
	switch f.Direction {
	case directionIn:
		filtered = filtered.Where("qty_change > 0")
	case directionOut:
		filtered = filtered.Where("qty_change < 0")
	}
	if f.From != nil {
		filtered = filtered.Where("created_at >= ?", *f.From)
	}
	if f.Until != nil {
		filtered = filtered.Where("created_at < ?", *f.Until)
	}
	if f.OrderItemID != "" {
		filtered = filtered.Where("order_item_id = ?", f.OrderItemID)
	}
	if f.UsageID != "" {
		filtered = filtered.Where("usage_id = ?", f.UsageID)
	}

	// Session ทำให้ใช้ query เดียวกันซ้ำได้ทั้งตอน Count และ Find
	filtered = filtered.Session(&gorm.Session{})

	var total int64
	if err := filtered.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []models.StockLog
	// Unscoped: สินค้าที่ถูก soft delete ต้องยังแสดงชื่อในประวัติได้ (log ห้ามหาย)
	withDeleted := func(db *gorm.DB) *gorm.DB { return db.Unscoped() }

	// id เป็นตัวตัดสินเมื่อ created_at เท่ากัน ไม่งั้นข้ามหน้าแล้วแถวอาจซ้ำหรือตกหล่น
	err := filtered.
		Preload("Stock", withDeleted).
		Preload("Branch").
		Preload("Employee").
		Order("created_at DESC, id DESC").
		Limit(f.PageSize).
		Offset((f.Page - 1) * f.PageSize).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// CreateStockLog บันทึก stock log 1 แถว
// ถ้าเป็นการตัดออก (qty_change < 0) จะล็อกแถวสินค้าแล้วเช็คยอดคงเหลือใน transaction เดียวกับการบันทึก
// คนอื่นที่ตัดสินค้าตัวเดียวกันพร้อมกันต้องรอจนเรา commit จึงเห็นยอดล่าสุดเสมอ สต็อกไม่ติดลบ
func (r *repository) CreateStockLog(ctx context.Context, log *models.StockLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if log.QtyChange < 0 {
			if err := lockStock(tx, log.StockID); err != nil {
				return err
			}
			if err := checkEnoughStock(tx, log.StockID, log.BranchID, -log.QtyChange); err != nil {
				return err
			}
		}
		return tx.Create(log).Error
	})
}

// lockStock ล็อกแถวสินค้า (SELECT ... FOR UPDATE) ไว้จนกว่า transaction จะจบ
func lockStock(tx *gorm.DB, stockId string) error {
	var stock models.Stock
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&stock, "id = ?", stockId).Error
}

// checkEnoughStock เช็คว่าสาขานี้มีสินค้าพอให้ตัดออก need ชิ้น
// ต้องเรียกหลัง lockStock ใน transaction เดียวกัน ยอดที่อ่านได้จึงเป็นยอดล่าสุด
func checkEnoughStock(tx *gorm.DB, stockId, branchId string, need int) error {
	var available int
	if err := tx.Model(&models.StockLog{}).
		Where("stock_id = ? AND branch_id = ?", stockId, branchId).
		Select("COALESCE(SUM(qty_change), 0)").
		Scan(&available).Error; err != nil {
		return err
	}
	if available < need {
		return invalid("insufficient stock: available %d, requested %d", available, need)
	}
	return nil
}

// TransferStockLog ลง 2 แถวใน transaction เดียว: สาขาต้นทาง (-qty) และสาขาปลายทาง (+qty)
// ถ้าแถวใดแถวหนึ่งพัง จะ rollback ทั้งคู่
func (r *repository) TransferStockLog(ctx context.Context, data TransferStockLogRequest, logType string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		note := data.Note // แต่ละแถวต้องมี pointer ของตัวเอง

		out := &models.StockLog{
			LogType:    logType,
			StockID:    data.StockID,
			BranchID:   data.FromBranchId,
			EmployeeID: data.EmployeeID,
			QtyChange:  -data.Qty,
			Note:       &note,
		}
		if err := tx.Create(out).Error; err != nil {
			return err
		}

		inNote := data.Note
		in := &models.StockLog{
			LogType:    logType,
			StockID:    data.StockID,
			BranchID:   data.ToBranchId,
			EmployeeID: data.EmployeeID,
			QtyChange:  data.Qty,
			Note:       &inNote,
		}
		return tx.Create(in).Error
	})
}

// GetStockQuantity คืนยอดคงเหลือของสินค้า 1 ชนิดใน 1 สาขา
func (r *repository) GetStockQuantity(ctx context.Context, stockId string, branchId string) (int, error) {
	var quantity int
	err := r.db.WithContext(ctx).
		Model(&models.StockLog{}).
		Where("stock_id = ? AND branch_id = ?", stockId, branchId).
		Select("COALESCE(SUM(qty_change), 0)").
		Scan(&quantity).Error
	if err != nil {
		return 0, err
	}
	return quantity, nil
}
