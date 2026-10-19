package stock

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"backend/internal/models"
	_package "backend/package"
)

// error ที่ส่งกลับให้ client (_package.SendError แปลง Code เป็น HTTP status)
var (
	ErrBranchRequired = _package.NewAppError(
		"Branch is required",
		_package.CodeForbidden,
		"Your account is not linked to a branch",
	) // 403
)

// การแบ่งหน้า: หน้าเริ่มที่ 1, ค่าเริ่มต้น 20 แถวต่อหน้า
// maxPage กันเลขหน้าใหญ่ผิดปกติจนคำนวณ offset ล้น
const (
	defaultPage     = 1
	maxPage         = 1_000_000
	defaultPageSize = 20
	maxPageSize     = 100
	maxSearchLen    = 100 // จำนวนตัวอักษรสูงสุดของ ?search=
)

// id ที่ไม่ใช่ UUID ถ้าส่งไปถึง PostgreSQL จะ error กลายเป็น 500 เลยตรวจก่อน
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Service interface {
	GetAllStock(ctx context.Context) ([]models.Stock, error)
	GetStockContainersByStockID(ctx context.Context, userRole string, sessionBranchId string, q ContainerQuery) ([]models.StockContainer, _package.PageMeta, error)
	GetStockQuantities(ctx context.Context, userRole string, sessionBranchId string, q QuantityQuery) ([]StockQuantity, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

// ดึงรายการ Stock ทั้งหมดในระบบ
func (s *service) GetAllStock(ctx context.Context) ([]models.Stock, error) {
	return s.repository.GetAllStock(ctx)
}

// GetStockContainersByStockID คืนกระปุกของสินค้า 1 ชนิด
// admin: เลือกสาขาด้วย q.BranchID ได้ ไม่ส่งมา = ทุกสาขา
// บัญชีอื่น: ถูกบังคับเป็นสาขาของตัวเองเสมอ (ไม่สนค่า q.BranchID กันดูข้ามสาขา)

func (s *service) GetStockContainersByStockID(ctx context.Context, userRole string, sessionBranchId string, q ContainerQuery) ([]models.StockContainer, _package.PageMeta, error) {
	var noMeta _package.PageMeta

	if !uuidPattern.MatchString(q.StockID) {
		return nil, noMeta, invalid("stock id must be a valid UUID")
	}

	status := strings.ToLower(strings.TrimSpace(q.Status))
	switch status {
	case "":
		status = StatusOpen
	case StatusOpen, StatusClosed, StatusAll:
	default:
		return nil, noMeta, invalid("status must be one of: open, closed, all")
	}

	page, err := parseBoundedInt("page", q.Page, defaultPage, maxPage)
	if err != nil {
		return nil, noMeta, err
	}
	pageSize, err := parseBoundedInt("pageSize", q.PageSize, defaultPageSize, maxPageSize)
	if err != nil {
		return nil, noMeta, err
	}

	var branchId string
	if userRole == "admin" {
		branchId = strings.TrimSpace(q.BranchID)
		if branchId != "" && !uuidPattern.MatchString(branchId) {
			return nil, noMeta, invalid("branchId must be a valid UUID")
		}
	} else {
		if sessionBranchId == "" {
			return nil, noMeta, ErrBranchRequired
		}
		branchId = sessionBranchId
	}

	containers, total, err := s.repository.GetStockContainersByStockID(ctx, q.StockID, branchId, status, page, pageSize)
	if err != nil {
		return nil, noMeta, err
	}
	return containers, _package.NewPageMeta(page, pageSize, total), nil
}

// GetStockQuantities คืนสินค้าพร้อมจำนวนคงเหลือ
// admin: เลือกสาขาด้วย q.BranchID ได้ ไม่ส่งมา = รวมทุกสาขา
// บัญชีอื่น: ถูกบังคับเป็นสาขาของตัวเองเสมอ
func (s *service) GetStockQuantities(ctx context.Context, userRole string, sessionBranchId string, q QuantityQuery) ([]StockQuantity, error) {
	f := QuantityFilter{}

	search := strings.TrimSpace(q.Search)
	if len([]rune(search)) > maxSearchLen {
		return nil, invalid("search must be at most %d characters", maxSearchLen)
	}
	f.Search = search

	category := strings.ToLower(strings.TrimSpace(q.Category))
	switch category {
	case "", "premium", "normal":
		f.Category = category
	default:
		return nil, invalid("category must be one of: premium, normal")
	}

	if raw := strings.TrimSpace(q.LowStock); raw != "" {
		low, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, invalid("lowStock must be true or false")
		}
		f.LowStock = low
	}

	if userRole == "admin" {
		f.BranchID = strings.TrimSpace(q.BranchID)
		if f.BranchID != "" && !uuidPattern.MatchString(f.BranchID) {
			return nil, invalid("branchId must be a valid UUID")
		}
	} else {
		if sessionBranchId == "" {
			return nil, ErrBranchRequired
		}
		f.BranchID = sessionBranchId
	}

	return s.repository.GetStockQuantities(ctx, f)
}

func parseBoundedInt(name, raw string, def, max int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > max {
		return 0, invalid("%s must be a number between 1 and %d", name, max)
	}
	return n, nil
}

func invalid(format string, args ...any) error {
	return _package.NewAppError("Invalid request", _package.CodeValidation, fmt.Sprintf(format, args...))
}
