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
	"backend/package"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// error ที่ส่งกลับให้ client (_package.SendError แปลง Code เป็น HTTP status)
var (
	ErrForbidden  = _package.NewAppError("Not allowed", _package.CodeForbidden, "")                                    // 403
	ErrBranchOnly = _package.NewAppError("Not allowed", _package.CodeForbidden, "only branch users can create orders") // 403
	ErrNotFound   = _package.NewAppError("Order not found", _package.CodeNotFound, "")                                 // 404
)

// listLimit = จำนวนบิลสูงสุดต่อการเรียก GET /api/orders (V1 ยังไม่มี pagination)
const listLimit = 100

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
	List(ctx context.Context, user *currentuser.CurrentUser, q ListOrdersQuery) ([]OrderRow, error)
	GetByID(ctx context.Context, user *currentuser.CurrentUser, id string) (*OrderDetail, error)
	AddPayment(ctx context.Context, user *currentuser.CurrentUser, orderID string, req CreatePayment) (*OrderDetail, error)
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
		return nil, ErrBranchOnly
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
		payment, err := newPayment(fmt.Sprintf("payments[%d]", i), p)
		if err != nil {
			return nil, err
		}
		paid = paid.Add(payment.Amount)
		payments = append(payments, payment)
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

func (s *service) List(ctx context.Context, user *currentuser.CurrentUser, q ListOrdersQuery) ([]OrderRow, error) {
	f := ListFilter{Limit: listLimit}

	// สาขาเห็นแค่บิลของตัวเอง admin เห็นทุกสาขา
	if user.Role != "admin" {
		if user.BranchID == nil {
			return nil, ErrForbidden
		}
		f.BranchID = user.BranchID
	}

	if q.CustomerID != "" {
		if !uuidPattern.MatchString(q.CustomerID) {
			return nil, invalid("customer_id must be a UUID")
		}
		f.CustomerID = &q.CustomerID
	}
	if q.From != "" {
		d, err := time.Parse(time.DateOnly, q.From)
		if err != nil {
			return nil, invalid("from must be YYYY-MM-DD")
		}
		f.From = &d
	}
	if q.To != "" {
		d, err := time.Parse(time.DateOnly, q.To)
		if err != nil {
			return nil, invalid("to must be YYYY-MM-DD")
		}
		f.To = &d
	}
	if f.From != nil && f.To != nil && f.From.After(*f.To) {
		return nil, invalid("from must not be after to")
	}
	if q.Status != "" {
		if q.Status != "paid" && q.Status != "partial" && q.Status != "unpaid" {
			return nil, invalid("status must be paid, partial or unpaid")
		}
		f.Status = &q.Status
	}

	rows, err := s.repository.List(ctx, f)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].Outstanding = rows[i].Total.Sub(rows[i].Paid)
		rows[i].Status = statusOf(rows[i].Total, rows[i].Paid)
	}
	return rows, nil
}

func (s *service) GetByID(ctx context.Context, user *currentuser.CurrentUser, id string) (*OrderDetail, error) {
	if !uuidPattern.MatchString(id) {
		return nil, ErrNotFound
	}

	order, items, payments, err := s.repository.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// บิลของสาขาอื่น ตอบเหมือนไม่มีบิลนี้ (ไม่บอกว่ามีอยู่จริง)
	if user.Role != "admin" && (user.BranchID == nil || order.BranchID != *user.BranchID) {
		return nil, ErrNotFound
	}

	return buildDetail(*order, items, payments), nil
}

// AddPayment บันทึกการจ่ายเพิ่มให้บิล (สาขาเจ้าของบิล หรือ admin) แล้วส่งบิลที่อัปเดตแล้วกลับ
// การจ่ายที่บันทึกแล้วแก้หรือลบไม่ได้ (ไม่มี API ให้แก้)
func (s *service) AddPayment(ctx context.Context, user *currentuser.CurrentUser, orderID string, req CreatePayment) (*OrderDetail, error) {
	// ใช้ GetByID เช็คสิทธิ์: บิลไม่มีจริงหรือเป็นของสาขาอื่น → 404, admin ผ่านทุกบิล
	if _, err := s.GetByID(ctx, user, orderID); err != nil {
		return nil, err
	}

	payment, err := newPayment("payment", req)
	if err != nil {
		return nil, err
	}

	// เช็คยอดค้างจ่ายและบันทึกใน transaction เดียวภายใต้การล็อกบิล (ดู repository.AddPayment)
	if err := s.repository.AddPayment(ctx, orderID, &payment); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, user, orderID)
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

	return &OrderDetail{
		Order:       order,
		Items:       items,
		Payments:    payments,
		Total:       total,
		Paid:        paid,
		Outstanding: total.Sub(paid),
		Status:      statusOf(total, paid),
	}
}

// statusOf สถานะการจ่ายของบิล (เงื่อนไขต้องตรงกับการกรอง status ใน repository.List)
func statusOf(total, paid decimal.Decimal) string {
	if paid.GreaterThanOrEqual(total) {
		return "paid"
	}
	if paid.IsPositive() {
		return "partial"
	}
	return "unpaid"
}

// newPayment ตรวจข้อมูลการจ่าย 1 รายการ แล้วสร้าง models.Payment
// label ใช้นำหน้าข้อความ error เช่น "payments[0]" หรือ "payment"
func newPayment(label string, p CreatePayment) (models.Payment, error) {
	if p.Method != "cash" && p.Method != "transfer" && p.Method != "card" {
		return models.Payment{}, invalid("%s: method must be cash, transfer or card", label)
	}
	if !isMoney(p.Amount) || !p.Amount.IsPositive() {
		return models.Payment{}, invalid("%s: amount must be more than 0 with at most 2 decimals", label)
	}
	ref := strings.TrimSpace(p.Reference)
	if utf8.RuneCountInString(ref) > 100 {
		return models.Payment{}, invalid("%s: reference must be at most 100 characters", label)
	}
	return models.Payment{Method: p.Method, Amount: p.Amount, Reference: ref}, nil
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

// invalid สร้าง error 400 พร้อมรายละเอียดว่าข้อมูลส่วนไหนผิด
func invalid(format string, args ...any) error {
	return _package.NewAppError("Invalid order data", _package.CodeValidation, fmt.Sprintf(format, args...))
}
