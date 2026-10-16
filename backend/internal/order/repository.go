package order

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	CustomerExists(ctx context.Context, id string) (bool, error)
	FindCourses(ctx context.Context, ids []string) (map[string]models.Course, error)
	FindStocks(ctx context.Context, ids []string) (map[string]models.Stock, error)
	Create(ctx context.Context, yymm string, order *models.Order, items []models.OrderItem, payments []models.Payment) error
	List(ctx context.Context, f ListFilter) ([]OrderRow, error)
	GetByID(ctx context.Context, id string) (*models.Order, []models.OrderItem, []models.Payment, error)
}

type repository struct {
	db *gorm.DB
}

// ListFilter เงื่อนไขค้นหาบิล ช่องที่เป็น nil = ไม่กรองข้อนั้น
type ListFilter struct {
	BranchID   *string    // nil = ทุกสาขา (admin)
	CustomerID *string    // บิลของลูกค้าคนนี้
	From       *time.Time // order_date ตั้งแต่วันนี้ (รวมวันนั้น)
	To         *time.Time // order_date ถึงวันนี้ (รวมวันนั้น)
	Status     *string    // paid / partial / unpaid
	Limit      int
}

// OrderRow 1 แถวในรายการบิล: ข้อมูลบิล + รหัสสาขา + ลูกค้า + ยอดรวม/ยอดจ่ายที่รวมมาจาก SQL
type OrderRow struct {
	models.Order `gorm:"embedded"`
	BranchCode   string          `json:"branch_code"`
	CustomerHnNo string          `json:"customer_hn_no"`
	CustomerName string          `json:"customer_name"`
	Total        decimal.Decimal `json:"total"`
	Paid         decimal.Decimal `json:"paid"`
	Outstanding  decimal.Decimal `gorm:"-" json:"outstanding"` // service เติมให้
	Status       string          `gorm:"-" json:"status"`      // service เติมให้
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CustomerExists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Customer{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// FindCourses ดึงคอร์สตาม id คืนเป็น map[id]คอร์ส (คอร์สที่ถูกลบแบบ soft delete จะไม่มีใน map)
func (r *repository) FindCourses(ctx context.Context, ids []string) (map[string]models.Course, error) {
	result := map[string]models.Course{}
	if len(ids) == 0 {
		return result, nil
	}
	var courses []models.Course
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&courses).Error; err != nil {
		return nil, err
	}
	for _, c := range courses {
		result[c.ID] = c
	}
	return result, nil
}

// FindStocks ดึงสินค้าตาม id คืนเป็น map[id]สินค้า
func (r *repository) FindStocks(ctx context.Context, ids []string) (map[string]models.Stock, error) {
	result := map[string]models.Stock{}
	if len(ids) == 0 {
		return result, nil
	}
	var stocks []models.Stock
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&stocks).Error; err != nil {
		return nil, err
	}
	for _, s := range stocks {
		result[s.ID] = s
	}
	return result, nil
}

// Create บันทึกบิล + รายการ + การจ่าย ใน transaction เดียว และออกเลขบิลให้
// เลขบิล = รหัสสาขา-YYMM-เลขรัน 4 หลักของสาขานั้นในเดือนนั้น เช่น S1-2610-0001
// อะไรผิดกลางทาง ทุกอย่างถูกยกเลิก (rollback) ไม่มีบิลครึ่งๆ กลางๆ ค้างในฐานข้อมูล
func (r *repository) Create(ctx context.Context, yymm string, order *models.Order, items []models.OrderItem, payments []models.Payment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) ล็อกแถวสาขา (SELECT ... FOR UPDATE)
		// คนอื่นที่ออกบิลสาขาเดียวกันพร้อมกันต้องรอจนเรา commit เลขจึงไม่ซ้ำ
		var branch models.Branch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&branch, "id = ?", order.BranchID).Error; err != nil {
			return err
		}

		// 2) หาเลขล่าสุดของสาขานี้ในเดือนนี้ แล้ว +1
		prefix := fmt.Sprintf("%s-%s-", branch.Code, yymm)
		var last models.Order
		err := tx.Where("order_no LIKE ?", prefix+"%").Order("order_no DESC").First(&last).Error
		next := 1
		if err == nil {
			n, convErr := strconv.Atoi(strings.TrimPrefix(last.OrderNo, prefix))
			if convErr != nil {
				return convErr
			}
			next = n + 1
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		order.OrderNo = fmt.Sprintf("%s%04d", prefix, next)

		// 3) บันทึกบิล (ได้ order.ID กลับมา)
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 4) บันทึกรายการ: ผูกกับบิล และให้เลขบรรทัด 1..n ตามลำดับที่ส่งมา
		for i := range items {
			items[i].OrderID = order.ID
			items[i].LineNo = i + 1
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		// 5) บันทึกการจ่าย (ถ้ามี)
		if len(payments) > 0 {
			for i := range payments {
				payments[i].OrderID = order.ID
			}
			if err := tx.Create(&payments).Error; err != nil {
				return err
			}
		}

		return nil // return nil = commit
	})
}

// List รายการบิลตามเงื่อนไข เรียงจากวันขายล่าสุด
// ยอดรวมและยอดจ่ายรวมด้วย SQL (SUM) จึงกรองตามสถานะได้ในคำสั่งเดียว
func (r *repository) List(ctx context.Context, f ListFilter) ([]OrderRow, error) {
	itemTotals := r.db.Model(&models.OrderItem{}).
		Select("order_id, SUM(qty * unit_price - discount) AS total").Group("order_id")
	paidTotals := r.db.Model(&models.Payment{}).
		Select("order_id, SUM(amount) AS paid").Group("order_id")

	q := r.db.WithContext(ctx).Table("orders AS o").
		Select("o.*, b.code AS branch_code, c.hn_no AS customer_hn_no, "+
			"TRIM(CONCAT(c.fname, ' ', c.lname)) AS customer_name, t.total, COALESCE(p.paid, 0) AS paid").
		Joins("JOIN branches AS b ON b.id = o.branch_id").
		Joins("JOIN customers AS c ON c.id = o.customer_id").
		Joins("JOIN (?) AS t ON t.order_id = o.id", itemTotals).
		Joins("LEFT JOIN (?) AS p ON p.order_id = o.id", paidTotals)

	if f.BranchID != nil {
		q = q.Where("o.branch_id = ?", *f.BranchID)
	}
	if f.CustomerID != nil {
		q = q.Where("o.customer_id = ?", *f.CustomerID)
	}
	if f.From != nil {
		q = q.Where("o.order_date >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("o.order_date <= ?", *f.To)
	}
	if f.Status != nil {
		// เงื่อนไขต้องตรงกับ statusOf ใน service
		switch *f.Status {
		case "paid":
			q = q.Where("COALESCE(p.paid, 0) >= t.total")
		case "partial":
			q = q.Where("COALESCE(p.paid, 0) > 0 AND COALESCE(p.paid, 0) < t.total")
		case "unpaid":
			q = q.Where("COALESCE(p.paid, 0) = 0 AND t.total > 0")
		}
	}

	rows := []OrderRow{}
	err := q.Order("o.order_date DESC, o.order_no DESC").Limit(f.Limit).Scan(&rows).Error
	return rows, err
}

// GetByID บิล 1 ใบ พร้อมลูกค้า สาขา รายการ (เรียงตาม line_no พร้อมชื่อคอร์ส/สินค้า) และการจ่าย (เรียงตามเวลา)
func (r *repository) GetByID(ctx context.Context, id string) (*models.Order, []models.OrderItem, []models.Payment, error) {
	db := r.db.WithContext(ctx)

	var order models.Order
	if err := db.Preload("Customer").Preload("Branch").First(&order, "id = ?", id).Error; err != nil {
		return nil, nil, nil, err
	}

	// Unscoped = ดึงคอร์ส/สินค้าที่ถูกลบแบบ soft delete ไปแล้วด้วย บิลเก่าจะได้ยังเห็นชื่อ
	unscoped := func(tx *gorm.DB) *gorm.DB { return tx.Unscoped() }
	items := []models.OrderItem{}
	if err := db.Preload("Course", unscoped).Preload("Stock", unscoped).
		Where("order_id = ?", id).Order("line_no").Find(&items).Error; err != nil {
		return nil, nil, nil, err
	}

	payments := []models.Payment{}
	if err := db.Where("order_id = ?", id).Order("paid_at").Find(&payments).Error; err != nil {
		return nil, nil, nil, err
	}

	return &order, items, payments, nil
}
