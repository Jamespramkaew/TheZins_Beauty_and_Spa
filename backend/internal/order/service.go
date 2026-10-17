package order

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"backend/internal/currentuser"
	"backend/internal/models"

	"github.com/shopspring/decimal"
)

// error ที่ handler ใช้เลือก HTTP status
var (
	ErrInvalidInput = errors.New("invalid input")                       // 400
	ErrForbidden    = errors.New("only branch users can create orders") // 403
)

// bangkok = เวลาไทย (UTC+7 ไม่มีเวลาออมแสง) ใช้หา "วันนี้" และเดือนในเลขบิล
var bangkok = time.FixedZone("Asia/Bangkok", 7*60*60)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// OrderDetail = บิลที่ส่งกลับให้ client: ข้อมูลบิล + รายการ + การจ่าย + ยอดที่คำนวณสด (ไม่เก็บในตาราง)
type OrderDetail struct {
	models.Order
	Items       []models.OrderItem `json:"items"`
	Payments    []models.Payment   `json:"payments"`
	Total       decimal.Decimal    `json:"total"`       // Σ(qty × unit_price − discount)
	Paid        decimal.Decimal    `json:"paid"`        // Σ amount
	Outstanding decimal.Decimal    `json:"outstanding"` // total − paid
	Status      string             `json:"status"`      // paid / partial / unpaid
}

type Service interface {
	Create(ctx context.Context, user *currentuser.CurrentUser, req CreateOrder) (*OrderDetail, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) Create(ctx context.Context, user *currentuser.CurrentUser, req CreateOrder) (*OrderDetail, error) {
	// 1) ใครออกบิล: ต้องเป็นบัญชีสาขา branch มาจากผู้ใช้ ไม่ได้มาจาก body
	if user.Role != "branch" || user.BranchID == nil {
		return nil, ErrForbidden
	}

	now := time.Now().In(bangkok)
	today := dateOnly(now)

	// 2) ลูกค้า
	if !uuidPattern.MatchString(req.CustomerID) {
		return nil, invalid("customer_id is required")
	}
	exists, err := s.repository.CustomerExists(ctx, req.CustomerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, invalid("customer not found")
	}

	// 3) วันที่ขาย: ไม่ส่งมา = วันนี้, ย้อนหลังได้, ห้ามเป็นวันในอนาคต
	orderDate := today
	if req.OrderDate != nil && strings.TrimSpace(*req.OrderDate) != "" {
		d, err := time.Parse(time.DateOnly, strings.TrimSpace(*req.OrderDate))
		if err != nil {
			return nil, invalid("order_date must be YYYY-MM-DD")
		}
		if d.After(today) {
			return nil, invalid("order_date cannot be in the future")
		}
		orderDate = d
	}

	// 4) ช่องทาง
	var source *string
	if req.Source != nil && strings.TrimSpace(*req.Source) != "" {
		v := strings.TrimSpace(*req.Source)
		if v != "phone" && v != "walk-in" && v != "page" {
			return nil, invalid("source must be phone, walk-in or page")
		}
		source = &v
	}

	// 5) รายการ: ตรวจรูปแบบก่อน แล้วค่อยดึงคอร์ส/สินค้าทีเดียว
	if len(req.Items) == 0 {
		return nil, invalid("items must have at least 1 item")
	}
	var courseIDs, stockIDs []string
	for i, it := range req.Items {
		switch it.ItemType {
		case "course":
			if it.CourseID == nil || !uuidPattern.MatchString(*it.CourseID) || it.StockID != nil {
				return nil, invalid("items[%d]: course item needs course_id only", i)
			}
			if it.Qty != nil && *it.Qty != 1 {
				return nil, invalid("items[%d]: course qty must be 1", i)
			}
			courseIDs = append(courseIDs, *it.CourseID)
		case "stock":
			if it.StockID == nil || !uuidPattern.MatchString(*it.StockID) || it.CourseID != nil {
				return nil, invalid("items[%d]: stock item needs stock_id only", i)
			}
			if it.Qty != nil && *it.Qty < 1 {
				return nil, invalid("items[%d]: qty must be at least 1", i)
			}
			stockIDs = append(stockIDs, *it.StockID)
		default:
			return nil, invalid("items[%d]: item_type must be course or stock", i)
		}
	}

	courses, err := s.repository.FindCourses(ctx, courseIDs)
	if err != nil {
		return nil, err
	}
	stocks, err := s.repository.FindStocks(ctx, stockIDs)
	if err != nil {
		return nil, err
	}

	// 6) snapshot ราคา จำนวนครั้ง วันหมดอายุ ณ ตอนขาย แล้วรวมยอด
	items := make([]models.OrderItem, 0, len(req.Items))
	total := decimal.Zero
	for i, it := range req.Items {
		qty := 1
		if it.Qty != nil {
			qty = *it.Qty
		}
		item := models.OrderItem{ItemType: it.ItemType, Qty: qty, Discount: it.Discount}

		if it.ItemType == "course" {
			c, ok := courses[*it.CourseID]
			if !ok {
				return nil, invalid("items[%d]: course not found", i)
			}
			if !c.IsActive {
				return nil, invalid("items[%d]: course is not active", i)
			}
			item.CourseID = &c.ID
			item.UnitPrice = c.Price
			sessions := c.MaxSession
			item.TotalSession = &sessions
			if c.ValidDay != nil {
				expires := orderDate.AddDate(0, 0, *c.ValidDay)
				item.ExpiresAt = &expires
			}
		} else {
			st, ok := stocks[*it.StockID]
			if !ok {
				return nil, invalid("items[%d]: stock not found", i)
			}
			if !st.IsActive || !st.IsSellable {
				return nil, invalid("items[%d]: stock is not for sale", i)
			}
			item.StockID = &st.ID
			item.UnitPrice = st.Price
		}

		gross := item.UnitPrice.Mul(decimal.NewFromInt(int64(qty)))
		if !isMoney(it.Discount) || it.Discount.IsNegative() || it.Discount.GreaterThan(gross) {
			return nil, invalid("items[%d]: discount must be between 0 and %s with at most 2 decimals", i, gross.StringFixed(2))
		}
		total = total.Add(gross.Sub(it.Discount))
		items = append(items, item)
	}

	// 7) การจ่าย: วิธีจ่ายถูกต้อง, ยอด > 0, รวมแล้วห้ามเกินยอดบิล
	payments := make([]models.Payment, 0, len(req.Payments))
	paid := decimal.Zero
	for i, p := range req.Payments {
		if p.Method != "cash" && p.Method != "transfer" && p.Method != "card" {
			return nil, invalid("payments[%d]: method must be cash, transfer or card", i)
		}
		if !isMoney(p.Amount) || !p.Amount.IsPositive() {
			return nil, invalid("payments[%d]: amount must be more than 0 with at most 2 decimals", i)
		}
		ref := strings.TrimSpace(p.Reference)
		if utf8.RuneCountInString(ref) > 100 {
			return nil, invalid("payments[%d]: reference must be at most 100 characters", i)
		}
		paid = paid.Add(p.Amount)
		payments = append(payments, models.Payment{Method: p.Method, Amount: p.Amount, Reference: ref})
	}
	if paid.GreaterThan(total) {
		return nil, invalid("total payments %s exceed order total %s", paid.StringFixed(2), total.StringFixed(2))
	}

	// 8) บันทึก (เลขบิลออกใน repository ภายใต้การล็อกสาขา)
	order := models.Order{
		CustomerID: req.CustomerID,
		BranchID:   *user.BranchID,
		Source:     source,
		OrderDate:  orderDate,
	}
	if err := s.repository.Create(ctx, now.Format("0601"), &order, items, payments); err != nil {
		return nil, err
	}

	return buildDetail(order, items, payments), nil
}

// buildDetail รวมบิลกับรายการและการจ่าย แล้วคำนวณยอดและสถานะ
func buildDetail(order models.Order, items []models.OrderItem, payments []models.Payment) *OrderDetail {
	total := decimal.Zero
	for _, it := range items {
		total = total.Add(it.UnitPrice.Mul(decimal.NewFromInt(int64(it.Qty))).Sub(it.Discount))
	}
	paid := decimal.Zero
	for _, p := range payments {
		paid = paid.Add(p.Amount)
	}

	status := "unpaid"
	if paid.GreaterThanOrEqual(total) {
		status = "paid"
	} else if paid.IsPositive() {
		status = "partial"
	}

	return &OrderDetail{
		Order:       order,
		Items:       items,
		Payments:    payments,
		Total:       total,
		Paid:        paid,
		Outstanding: total.Sub(paid),
		Status:      status,
	}
}

// dateOnly ตัดเวลาออก เหลือแค่วันที่ (ตามปฏิทินของ t) เก็บเป็นเที่ยงคืน UTC
// ใช้ UTC เพราะ PostgreSQL แปลงค่าเวลาเป็นคอลัมน์ date ตาม timezone ของ session (UTC)
// ถ้าส่งเที่ยงคืนเวลาไทยไป จะถูกบันทึกเป็นวันก่อนหน้า
func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// isMoney = ทศนิยมไม่เกิน 2 ตำแหน่ง (คอลัมน์เงินเป็น numeric(12,2))
func isMoney(d decimal.Decimal) bool {
	return d.Equal(d.Round(2))
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrInvalidInput}, args...)...)
}
