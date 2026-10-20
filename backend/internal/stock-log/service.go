package stocklog

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	_package "backend/package"
)

var errBranchRequired = _package.NewAppError(
	"Branch is required",
	_package.CodeForbidden,
	"Your account is not linked to a branch",
)

const (
	activeTrue  = "true"
	activeFalse = "false"
	activeAll   = "all"
)

const (
	defaultPage     = 1
	maxPage         = 1_000_000
	defaultPageSize = 20
	maxPageSize     = 100
	maxSearchLen    = 100
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Service interface {
	GetAllStockQuantities(ctx context.Context, userRole string, sessionBranchId string, q GetAllStockQuantitiesQuery) ([]StockQuantity, _package.PageMeta, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

// GetAllStockQuantities คืนยอดคงเหลือของสินค้าแยกตามสาขา
// admin: เลือกสาขาด้วย q.BranchID ได้ แต่ถ้าไม่ส่งมา = ทุกสาขา
// บัญชีอื่น: ถูกบังคับเป็นสาขาของตัวเองเสมอ (ไม่สนค่า q.BranchID กันดูข้ามสาขา)
func (s *service) GetAllStockQuantities(ctx context.Context, userRole string, sessionBranchId string, q GetAllStockQuantitiesQuery) ([]StockQuantity, _package.PageMeta, error) {
	var noMeta _package.PageMeta

	var branchId string
	if userRole == "admin" {
		branchId = strings.TrimSpace(q.BranchID)
		if branchId != "" && !uuidPattern.MatchString(branchId) {
			return nil, noMeta, invalid("branchId must be a valid UUID")
		}
	} else {
		if sessionBranchId == "" {
			return nil, noMeta, errBranchRequired
		}
		branchId = sessionBranchId
	}

	category := strings.ToLower(strings.TrimSpace(q.Category))
	if category != "" && category != "premium" && category != "normal" {
		return nil, noMeta, invalid("category must be one of: premium, normal")
	}

	active := strings.ToLower(strings.TrimSpace(q.Active))
	switch active {
	case "":
		active = activeTrue
	case activeTrue, activeFalse, activeAll:
	default:
		return nil, noMeta, invalid("active must be one of: true, false, all")
	}

	search := strings.TrimSpace(q.Search)
	if len([]rune(search)) > maxSearchLen {
		return nil, noMeta, invalid("search must not exceed %d characters", maxSearchLen)
	}

	page, err := boundedInt("page", q.Page, defaultPage, maxPage)
	if err != nil {
		return nil, noMeta, err
	}
	pageSize, err := boundedInt("pageSize", q.PageSize, defaultPageSize, maxPageSize)
	if err != nil {
		return nil, noMeta, err
	}

	items, total, err := s.repository.GetAllStockQuantities(ctx, quantityFilter{
		BranchID: branchId,
		Search:   search,
		Category: category,
		LowStock: q.LowStock,
		Active:   active,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, noMeta, err
	}

	for i := range items {
		items[i].IsLowStock = items[i].Quantity <= items[i].MinQty
	}
	if items == nil {
		items = []StockQuantity{}
	}
	return items, _package.NewPageMeta(page, pageSize, total), nil
}

func boundedInt(name string, n, def, max int) (int, error) {
	if n == 0 {
		return def, nil
	}
	if n < 1 || n > max {
		return 0, invalid("%s must be a number between 1 and %d", name, max)
	}
	return n, nil
}

func invalid(format string, args ...any) error {
	return _package.NewAppError("Invalid request", _package.CodeValidation, fmt.Sprintf(format, args...))
}
