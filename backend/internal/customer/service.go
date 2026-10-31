package customer

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"backend/internal/models"
	"backend/package"

	"gorm.io/gorm"
)

// error ที่ส่งกลับให้ client (_package.SendError แปลง Code เป็น HTTP status)
var (
	ErrNotFound    = _package.NewAppError("Customer not found", _package.CodeNotFound, "")                           // 404
	ErrDuplicateHN = _package.NewAppError("HN number already exists", _package.CodeConflict, "hn_no already exists") // 409
)

var (
	uuidPattern      = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	citizenIDPattern = regexp.MustCompile(`^[0-9]{13}$`)
)

type Service interface {
	Search(ctx context.Context, q SearchCustomersQuery) ([]models.Customer, _package.PageMeta, error)
	GetByID(ctx context.Context, id string) (*models.Customer, error)
	Create(ctx context.Context, req CreateCustomer) (*models.Customer, error)
	Update(ctx context.Context, id string, req UpdateCustomer) (*models.Customer, error)
	ChangeTier(ctx context.Context, id, changedBy string, req ChangeTier) (*models.Customer, error)
	ListTierChanges(ctx context.Context, id string) ([]TierChangeRow, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) Search(ctx context.Context, q SearchCustomersQuery) ([]models.Customer, _package.PageMeta, error) {
	page, pageSize, err := _package.ParsePaging(q.Page, q.PageSize)
	if err != nil {
		return nil, _package.PageMeta{}, err
	}
	customers, total, err := s.repository.Search(ctx, strings.TrimSpace(q.Q), page, pageSize)
	if err != nil {
		return nil, _package.PageMeta{}, err
	}
	return customers, _package.NewPageMeta(page, pageSize, total), nil
}

func (s *service) GetByID(ctx context.Context, id string) (*models.Customer, error) {
	// id ที่ไม่ใช่ UUID ไม่มีในตารางแน่นอน และถ้าส่งไปถึง PostgreSQL จะ error กลายเป็น 500
	if !uuidPattern.MatchString(id) {
		return nil, ErrNotFound
	}

	customer, err := s.repository.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return customer, err
}

func (s *service) Create(ctx context.Context, req CreateCustomer) (*models.Customer, error) {
	customer := models.Customer{
		HnNo:    strings.TrimSpace(req.HnNo),
		Fname:   strings.TrimSpace(req.Fname),
		Lname:   strings.TrimSpace(req.Lname),
		Tel:     strings.TrimSpace(req.Tel),
		Address: strings.TrimSpace(req.Address),
	}

	if err := checkText("hn_no", customer.HnNo, true, 20); err != nil {
		return nil, err
	}
	if err := checkText("fname", customer.Fname, true, 100); err != nil {
		return nil, err
	}
	if err := checkText("lname", customer.Lname, false, 100); err != nil {
		return nil, err
	}
	if err := checkText("tel", customer.Tel, false, 20); err != nil {
		return nil, err
	}

	var err error
	if customer.CitizenID, err = parseCitizenID(req.CitizenID); err != nil {
		return nil, err
	}
	if customer.BDay, err = parseBDay(req.BDay); err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, &customer); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateHN
		}
		return nil, err
	}
	return &customer, nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateCustomer) (*models.Customer, error) {
	// เช็คก่อนว่ามีลูกค้าคนนี้ ไม่มีก็ตอบ 404
	if _, err := s.GetByID(ctx, id); err != nil {
		return nil, err
	}

	// เก็บเฉพาะ field ที่ส่งมา (pointer ไม่เป็น nil)
	fields := map[string]any{}

	if err := setText(fields, "hn_no", req.HnNo, true, 20); err != nil {
		return nil, err
	}
	if err := setText(fields, "fname", req.Fname, true, 100); err != nil {
		return nil, err
	}
	if err := setText(fields, "lname", req.Lname, false, 100); err != nil {
		return nil, err
	}
	if err := setText(fields, "tel", req.Tel, false, 20); err != nil {
		return nil, err
	}
	if req.Address != nil {
		fields["address"] = strings.TrimSpace(*req.Address)
	}
	if req.CitizenID != nil {
		citizenID, err := parseCitizenID(req.CitizenID)
		if err != nil {
			return nil, err
		}
		fields["citizen_id"] = citizenID // nil = ล้างค่าเป็น NULL
	}
	if req.BDay != nil {
		bDay, err := parseBDay(req.BDay)
		if err != nil {
			return nil, err
		}
		fields["b_day"] = bDay // nil = ล้างค่าเป็น NULL
	}

	if len(fields) == 0 {
		return nil, invalid("no fields to update")
	}

	if err := s.repository.Update(ctx, id, fields); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateHN
		}
		return nil, err
	}

	// ดึงใหม่เพื่อส่งข้อมูลหลังแก้กลับไป
	return s.repository.GetByID(ctx, id)
}

// ChangeTier เปลี่ยนระดับลูกค้า changedBy = id ของบัญชีที่ login อยู่ (มาจาก handler ไม่ได้มาจาก body)
func (s *service) ChangeTier(ctx context.Context, id, changedBy string, req ChangeTier) (*models.Customer, error) {
	if !uuidPattern.MatchString(id) {
		return nil, ErrNotFound
	}

	tier := strings.TrimSpace(req.Tier)
	if tier != "general" && tier != "vip" {
		return nil, invalid("tier must be general or vip")
	}

	customer, err := s.repository.ChangeTier(ctx, id, tier, changedBy)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return customer, err
}

// ListTierChanges ประวัติการเปลี่ยนระดับ ใหม่ → เก่า
func (s *service) ListTierChanges(ctx context.Context, id string) ([]TierChangeRow, error) {
	// เช็คก่อนว่ามีลูกค้าคนนี้ ไม่งั้นลูกค้าที่ไม่มีอยู่จริงจะได้ [] เหมือนลูกค้าที่ยังไม่เคยเปลี่ยนระดับ
	if _, err := s.GetByID(ctx, id); err != nil {
		return nil, err
	}
	return s.repository.ListTierChanges(ctx, id)
}

// checkText ตรวจว่าต้องมีค่าไหม และยาวไม่เกิน maxLen ตัวอักษร
func checkText(field, value string, required bool, maxLen int) error {
	if required && value == "" {
		return invalid("%s is required", field)
	}
	if utf8.RuneCountInString(value) > maxLen {
		return invalid("%s must be at most %d characters", field, maxLen)
	}
	return nil
}

// setText ใช้ตอน PATCH: ถ้าส่งมา (ไม่เป็น nil) ให้ตัดช่องว่าง ตรวจ แล้วใส่ลง fields
func setText(fields map[string]any, column string, value *string, required bool, maxLen int) error {
	if value == nil {
		return nil
	}
	v := strings.TrimSpace(*value)
	if err := checkText(column, v, required, maxLen); err != nil {
		return err
	}
	fields[column] = v
	return nil
}

// parseCitizenID: nil หรือ "" = ไม่มีค่า (NULL), นอกนั้นต้องเป็นตัวเลข 13 หลัก
func parseCitizenID(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil, nil
	}
	if !citizenIDPattern.MatchString(v) {
		return nil, invalid("citizen_id must be 13 digits")
	}
	return &v, nil
}

// parseBDay: nil หรือ "" = ไม่มีค่า (NULL), นอกนั้นต้องเป็นรูปแบบ YYYY-MM-DD
func parseBDay(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse(time.DateOnly, v)
	if err != nil {
		return nil, invalid("b_day must be YYYY-MM-DD")
	}
	return &t, nil
}

// invalid สร้าง error 400 พร้อมรายละเอียดว่าข้อมูลช่องไหนผิด
func invalid(format string, args ...any) error {
	return _package.NewAppError("Invalid customer data", _package.CodeValidation, fmt.Sprintf(format, args...))
}
