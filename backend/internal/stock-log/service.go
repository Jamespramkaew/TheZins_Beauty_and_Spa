package stocklog

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"backend/internal/branch"
	"backend/internal/employee"
	"backend/internal/models"
	"backend/internal/stock"

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

// ค่าที่อนุญาตของ ?direction=
const (
	directionIn  = "in"
	directionOut = "out"
)

// log_type ที่ตรงกับ CHECK constraint ของ stock_logs
var validLogTypes = map[string]bool{
	"receive": true, "sale": true, "use": true, "open": true, "damaged": true,
	"expired": true, "internal": true, "transfer": true, "adjust": true, "count_diff": true,
}

// log type สำหรับ AddStockLogMovement พร้อมเครื่องหมายของ qty_change (+1 = เพิ่มยอด, -1 = ลดยอด)
var movementSign = map[string]int{
	"receive":  1,
	"expired":  -1,
	"damaged":  -1,
	"internal": -1,
}

// วันที่ใน ?from= ?to= ถือเป็นเวลาไทย (UTC+7 ไม่มี DST) ใช้ FixedZone เพื่อไม่ต้องพึ่ง tzdata ของเครื่อง
var thaiTime = time.FixedZone("ICT", 7*60*60)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Service interface {
	GetAllStockQuantities(ctx context.Context, userRole string, sessionBranchId string, q GetAllStockQuantitiesQuery) ([]StockQuantity, _package.PageMeta, error)
	GetStockLogsByBranchId(ctx context.Context, userRole string, sessionBranchId string, q GetAllStockLogsQuery) ([]StockLogResponse, _package.PageMeta, error)
	AddStockLogMovement(ctx context.Context, userRole string, branchId string, data AddStockLogMovementRequest) (*models.StockLog, error)
}

type service struct {
	repository         Repository
	stockRepository    stock.Repository
	branchRepository   branch.Repository
	employeeRepository employee.Repository
}

func NewService(repository Repository, stockRepository stock.Repository, branchRepository branch.Repository, employeeRepository employee.Repository) Service {
	return &service{
		repository:         repository,
		stockRepository:    stockRepository,
		branchRepository:   branchRepository,
		employeeRepository: employeeRepository,
	}
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

// GetStockLogsByBranchId คืนประวัติ stock_logs ทีละหน้า เรียงใหม่สุดก่อน
// admin: เลือกสาขาด้วย q.BranchID ได้ ไม่ส่งมา = ทุกสาขา
// บัญชีอื่น: ถูกบังคับเป็นสาขาของตัวเองเสมอ (ไม่สนค่า q.BranchID กันดูข้ามสาขา)
func (s *service) GetStockLogsByBranchId(ctx context.Context, userRole string, sessionBranchId string, q GetAllStockLogsQuery) ([]StockLogResponse, _package.PageMeta, error) {
	var noMeta _package.PageMeta

	var branchId string
	if userRole == "admin" {
		branchId = strings.TrimSpace(q.BranchID)
		if err := checkOptionalUUID("branchId", branchId); err != nil {
			return nil, noMeta, err
		}
	} else {
		if sessionBranchId == "" {
			return nil, noMeta, errBranchRequired
		}
		branchId = sessionBranchId
	}

	stockId := strings.TrimSpace(q.StockID)
	if err := checkOptionalUUID("stockId", stockId); err != nil {
		return nil, noMeta, err
	}
	orderItemId := strings.TrimSpace(q.OrderItemID)
	if err := checkOptionalUUID("orderItemId", orderItemId); err != nil {
		return nil, noMeta, err
	}
	usageId := strings.TrimSpace(q.UsageID)
	if err := checkOptionalUUID("usageId", usageId); err != nil {
		return nil, noMeta, err
	}

	logType := strings.ToLower(strings.TrimSpace(q.LogType))
	if logType != "" && !validLogTypes[logType] {
		return nil, noMeta, invalid("logType must be one of: receive, sale, use, open, damaged, expired, internal, transfer, adjust, count_diff")
	}

	direction := strings.ToLower(strings.TrimSpace(q.Direction))
	switch direction {
	case "", directionIn, directionOut:
	default:
		return nil, noMeta, invalid("direction must be one of: in, out")
	}

	var from, until *time.Time
	if raw := strings.TrimSpace(q.From); raw != "" {
		d, err := time.ParseInLocation(time.DateOnly, raw, thaiTime)
		if err != nil {
			return nil, noMeta, invalid("from must be YYYY-MM-DD")
		}
		from = &d
	}
	if raw := strings.TrimSpace(q.To); raw != "" {
		d, err := time.ParseInLocation(time.DateOnly, raw, thaiTime)
		if err != nil {
			return nil, noMeta, invalid("to must be YYYY-MM-DD")
		}
		end := d.AddDate(0, 0, 1) // นับรวมทั้งวันของ to
		until = &end
	}
	if from != nil && until != nil && !from.Before(*until) {
		return nil, noMeta, invalid("from must not be after to")
	}

	page, err := boundedInt("page", q.Page, defaultPage, maxPage)
	if err != nil {
		return nil, noMeta, err
	}
	pageSize, err := boundedInt("pageSize", q.PageSize, defaultPageSize, maxPageSize)
	if err != nil {
		return nil, noMeta, err
	}

	logs, total, err := s.repository.GetStockLogs(ctx, stockLogFilter{
		StockID:     stockId,
		BranchID:    branchId,
		LogType:     logType,
		Direction:   direction,
		From:        from,
		Until:       until,
		OrderItemID: orderItemId,
		UsageID:     usageId,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		return nil, noMeta, err
	}
	items := make([]StockLogResponse, 0, len(logs)) // ตอบ [] ไม่ใช่ null
	for _, l := range logs {
		items = append(items, toStockLogResponse(l))
	}
	return items, _package.NewPageMeta(page, pageSize, total), nil
}

// AddStockLogMovement เพิ่ม Stock-log สำหรับ receive, expired, damaged, internal
func (s *service) AddStockLogMovement(ctx context.Context, userRole string, branchId string, data AddStockLogMovementRequest) (*models.StockLog, error) {

	logType := strings.ToLower(strings.TrimSpace(data.Type))
	if logType == "" {
		return nil, invalid("type is required")
	}
	sign, ok := movementSign[logType]
	if !ok {
		return nil, invalid("type must be one of: receive, expired, damaged, internal")
	}

	if data.Qty <= 0 {
		return nil, invalid("qty must be positive")
	}

	stockID := strings.TrimSpace(data.StockID)
	if err := checkRequiredUUID("stockId", stockID); err != nil {
		return nil, err
	}

	employeeID := strings.TrimSpace(data.EmployeeID)
	if err := checkRequiredUUID("employeeId", employeeID); err != nil {
		return nil, err
	}

	var selectBranchId string
	if userRole == "admin" {
		branchIdCheck := strings.TrimSpace(data.BranchID)
		if err := checkRequiredUUID("branchId", branchIdCheck); err != nil {
			return nil, err
		}
		selectBranchId = branchIdCheck
	} else {
		if branchId == "" {
			return nil, errBranchRequired
		}
		selectBranchId = branchId
	}

	foundStock, err := s.stockRepository.FindStockById(ctx, stockID) // ไม่รวมสินค้าที่ถูก soft delete
	if err != nil {
		return nil, err
	}
	if foundStock == nil {
		return nil, invalid("stockId not found")
	}

	foundBranch, err := s.branchRepository.FindBranchById(ctx, selectBranchId)
	if err != nil {
		return nil, err
	}
	if foundBranch == nil {
		return nil, invalid("branchId not found")
	}
	if !foundBranch.IsActive {
		return nil, invalid("branch is not active")
	}

	foundEmployee, err := s.employeeRepository.FindEmployeeById(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	if foundEmployee == nil {
		return nil, invalid("employeeId not found")
	}
	if !foundEmployee.IsActive {
		return nil, invalid("employee is not active")
	}

	qtyChange := data.Qty * sign

	var note *string
	if n := strings.TrimSpace(data.Note); n != "" {
		note = &n
	}

	stockLog := &models.StockLog{
		LogType:    logType,
		StockID:    stockID,
		BranchID:   selectBranchId,
		EmployeeID: employeeID,
		QtyChange:  qtyChange,
		Note:       note,
	}

	if err := s.repository.CreateStockLog(ctx, stockLog); err != nil {
		return nil, err
	}
	return stockLog, nil
}

// checkOptionalUUID ว่างผ่าน ถ้ามีค่าต้องเป็น UUID
func checkOptionalUUID(name, v string) error {
	if v != "" && !uuidPattern.MatchString(v) {
		return invalid("%s must be a valid UUID", name)
	}
	return nil
}

func invalid(format string, args ...any) error {
	return _package.NewAppError("Invalid request", _package.CodeValidation, fmt.Sprintf(format, args...))
}

func checkRequiredUUID(name, v string) error {
	if v == "" {
		return invalid("%s is required", name)
	}
	if !uuidPattern.MatchString(v) {
		return invalid("%s must be a valid UUID", name)
	}
	return nil
}
