package stock

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"backend/internal/currentuser"
	"backend/internal/models"
	_package "backend/package"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
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
	GetStockContainersByStockID(ctx context.Context, user *currentuser.CurrentUser, q ContainerQuery) ([]models.StockContainer, _package.PageMeta, error)
	AddStock(ctx context.Context, data AddStockRequest) (models.Stock, error)
	UpdateStockById(ctx context.Context, stockId string, data UpdateStockRequest) (models.Stock, error)
	DeleteStockById(ctx context.Context, stockId string) error
	RestoreStockById(ctx context.Context, stockId string) (models.Stock, error)
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

func (s *service) GetStockContainersByStockID(ctx context.Context, user *currentuser.CurrentUser, q ContainerQuery) ([]models.StockContainer, _package.PageMeta, error) {
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
	if user.Role == "admin" {
		branchId = strings.TrimSpace(q.BranchID)
		if branchId != "" && !uuidPattern.MatchString(branchId) {
			return nil, noMeta, invalid("branchId must be a valid UUID")
		}
	} else {
		if user.BranchID == nil || *user.BranchID == "" {
			return nil, noMeta, ErrBranchRequired
		}
		branchId = *user.BranchID
	}

	containers, total, err := s.repository.GetStockContainersByStockID(ctx, q.StockID, branchId, status, page, pageSize)
	if err != nil {
		return nil, noMeta, err
	}
	return containers, _package.NewPageMeta(page, pageSize, total), nil
}

func (s *service) AddStock(ctx context.Context, data AddStockRequest) (models.Stock, error) {
	var emptyStock models.Stock

	// Validate ชื่อสินค้า
	stockName := strings.TrimSpace(data.StockName)
	if stockName == "" {
		return emptyStock, invalid("stock name is required")
	}
	if utf8.RuneCountInString(stockName) > 150 {
		return emptyStock, invalid("stock name must not exceed 150 characters")
	}

	// Validate Category
	category := strings.ToLower(strings.TrimSpace(data.Category))
	if category != "premium" && category != "normal" {
		return emptyStock, invalid("category must be either 'premium' or 'normal'")
	}

	// Validate Price
	if data.Price.IsNegative() {
		return emptyStock, invalid("price must be greater than or equal to 0")
	}

	// Validate UsesPerUnit
	if data.UsesPerUnit < 1 {
		return emptyStock, invalid("uses per unit must be at least 1")
	}

	// Validate MinQty
	if data.MinQty < 0 {
		return emptyStock, invalid("min quantity must be greater than or equal to 0")
	}

	// เช็คชื่อซ้ำในระบบ
	existingStock, err := s.repository.FindStockByName(ctx, stockName)
	if err != nil {
		return emptyStock, err
	}

	if existingStock != nil {
		// ถ้าไม่มี DeletedAt (ยังไม่ถูกลบ) → ห้ามเพิ่ม
		if existingStock.DeletedAt.Time.IsZero() {
			return emptyStock, duplicateName(stockName)
		}
	}

	// ถ้า IsSellable = false → ปรับ Price เป็น 0 อัตโนมัติ
	price := data.Price
	if !data.IsSellable {
		price = decimal.NewFromInt(0)
	}

	// สร้าง Stock ใหม่
	newStock := models.Stock{
		Name:        stockName,
		Price:       price,
		IsSellable:  data.IsSellable,
		Category:    category,
		UsesPerUnit: data.UsesPerUnit,
		MinQty:      data.MinQty,
		IsActive:    true,
	}

	createdStock, err := s.repository.CreateStock(ctx, newStock)
	if err != nil {
		// สองคำขอสร้างชื่อเดียวกันพร้อมกันผ่านการเช็คชื่อด้านบนได้ทั้งคู่ unique index เป็นตัวกันจริง
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return emptyStock, duplicateName(stockName)
		}
		return emptyStock, err
	}

	return createdStock, nil
}

// UpdateStockById อัปเดตข้อมูลสินค้า (เฉพาะฟิลด์ที่ส่งมา)
func (s *service) UpdateStockById(ctx context.Context, stockId string, data UpdateStockRequest) (models.Stock, error) {
	var emptyStock models.Stock

	// Validate UUID
	if !uuidPattern.MatchString(stockId) {
		return emptyStock, invalid("stock id must be a valid UUID")
	}

	// ดึงข้อมูล Stock เดิมจาก DB
	existingStock, err := s.repository.FindStockById(ctx, stockId)
	if err != nil {
		return emptyStock, err
	}
	if existingStock == nil {
		return emptyStock, _package.NewAppError(
			"Stock not found",
			_package.CodeNotFound,
			fmt.Sprintf("Stock with id '%s' not found", stockId),
		)
	}

	// ตรวจสอบและอัปเดตแต่ละฟิลด์ที่ส่งมา

	// 1. StockName
	if data.StockName != nil {
		stockName := strings.TrimSpace(*data.StockName)
		if stockName == "" {
			return emptyStock, invalid("stock name cannot be empty")
		}
		if utf8.RuneCountInString(stockName) > 150 {
			return emptyStock, invalid("stock name must not exceed 150 characters")
		}

		// เช็คชื่อซ้ำ (ยกเว้นตัวมันเอง)
		if stockName != existingStock.Name {
			duplicateStock, err := s.repository.FindStockByName(ctx, stockName)
			if err != nil {
				return emptyStock, err
			}
			if duplicateStock != nil && duplicateStock.ID != stockId {
				// ถ้าพบชื่อซ้ำกับ stock อื่น
				if duplicateStock.DeletedAt.Time.IsZero() {
					return emptyStock, duplicateName(stockName)
				}
			}
		}

		existingStock.Name = stockName
	}

	// 2. Price
	if data.PriceSet {
		if data.Price.IsNegative() {
			return emptyStock, invalid("price must be greater than or equal to 0")
		}
		existingStock.Price = data.Price
	}

	// 3. IsSellable
	if data.IsSellable != nil {
		existingStock.IsSellable = *data.IsSellable
	}

	// 4. Category
	if data.Category != nil {
		category := strings.ToLower(strings.TrimSpace(*data.Category))
		if category != "premium" && category != "normal" {
			return emptyStock, invalid("category must be either 'premium' or 'normal'")
		}
		existingStock.Category = category
	}

	// 5. UsesPerUnit
	if data.UsesPerUnit != nil {
		if *data.UsesPerUnit < 1 {
			return emptyStock, invalid("uses per unit must be at least 1")
		}
		existingStock.UsesPerUnit = *data.UsesPerUnit
	}

	// 6. MinQty
	if data.MinQty != nil {
		if *data.MinQty < 0 {
			return emptyStock, invalid("min quantity must be greater than or equal to 0")
		}
		existingStock.MinQty = *data.MinQty
	}

	// 7. IsActive
	if data.IsActive != nil {
		existingStock.IsActive = *data.IsActive
	}

	// ตรวจสอบเงื่อนไขสุดท้าย: ถ้า IsSellable = false → Price ต้องเป็น 0 เสมอ
	// (ทำหลังจากอัปเดตทุกฟิลด์แล้ว เพื่อให้แน่ใจว่าเงื่อนไขนี้มีผลเสมอ)
	if !existingStock.IsSellable {
		existingStock.Price = decimal.NewFromInt(0)
	}

	// บันทึกการเปลี่ยนแปลงลง DB
	updatedStock, err := s.repository.UpdateStock(ctx, *existingStock)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return emptyStock, duplicateName(existingStock.Name)
		}
		return emptyStock, err
	}

	return updatedStock, nil
}

// DeleteStockById ลบสินค้า:
// - ถ้าไม่มี Relation → Hard Delete (ลบจริง)
// - ถ้ามี Relation → Soft Delete (เซ็ต DeletedAt)
func (s *service) DeleteStockById(ctx context.Context, stockId string) error {

	// Validate UUID
	if !uuidPattern.MatchString(stockId) {
		return invalid("stock id must be a valid UUID")
	}

	// ดึงข้อมูล Stock
	existingStock, err := s.repository.FindStockById(ctx, stockId)
	if err != nil {
		return err
	}
	if existingStock == nil {
		return _package.NewAppError(
			"Stock not found",
			_package.CodeNotFound,
			fmt.Sprintf("Stock with id '%s' not found", stockId),
		)
	}

	// เช็คว่ามี Relation หรือไม่
	hasRelation, err := s.repository.CheckStockHasRelation(ctx, stockId)
	if err != nil {
		return err
	}

	if hasRelation {
		// มี Relation → Soft Delete
		err = s.repository.SoftDeleteStock(ctx, stockId)
		if err != nil {
			return err
		}
	} else {
		// ไม่มี Relation → Hard Delete
		err = s.repository.HardDeleteStock(ctx, stockId)
		if err != nil {
			return err
		}
	}

	return nil
}

// Helper Function

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

// nameConflict สร้าง error 409 "ชื่อสินค้าซ้ำ" ที่เดียวของทั้งไฟล์ (message และ code ต้องตรงกันทุกจุด)
func nameConflict(detail string) error {
	return _package.NewAppError("Stock name already exists", _package.CodeConflict, detail)
}

// duplicateName ชื่อซ้ำกับสินค้าที่ยังใช้งานอยู่ (ตอนสร้างและตอนเปลี่ยนชื่อ)
func duplicateName(name string) error {
	return nameConflict(fmt.Sprintf("Stock with name '%s' already exists in the system", name))
}

// duplicateNameOnRestore ชื่อซ้ำตอนกู้คืน (ข้อความ detail ต่างจากตอนสร้าง)
func duplicateNameOnRestore(name string) error {
	return nameConflict(fmt.Sprintf("Cannot restore: Stock with name '%s' already exists in the system", name))
}

func invalid(format string, args ...any) error {
	return _package.NewAppError("Invalid request", _package.CodeValidation, fmt.Sprintf(format, args...))
}

// RestoreStockById กู้คืนสินค้าที่ถูก soft delete
func (s *service) RestoreStockById(ctx context.Context, stockId string) (models.Stock, error) {
	var emptyStock models.Stock

	// Validate UUID
	if !uuidPattern.MatchString(stockId) {
		return emptyStock, invalid("stock id must be a valid UUID")
	}

	// ดึงข้อมูล Stock (รวมที่ถูก soft delete)
	existingStock, err := s.repository.FindStockByIdIncludingDeleted(ctx, stockId)
	if err != nil {
		return emptyStock, err
	}
	if existingStock == nil {
		return emptyStock, _package.NewAppError(
			"Stock not found",
			_package.CodeNotFound,
			fmt.Sprintf("Stock with id '%s' not found", stockId),
		)
	}

	// เช็คว่าถูก soft delete หรือไม่
	if existingStock.DeletedAt.Time.IsZero() {
		return emptyStock, _package.NewAppError(
			"Stock is not deleted",
			_package.CodeValidation,
			"Cannot restore stock that is not deleted",
		)
	}

	// เช็คว่าชื่อซ้ำกับสินค้าที่ยังใช้งานอยู่หรือไม่
	duplicateStock, err := s.repository.FindStockByName(ctx, existingStock.Name)
	if err != nil {
		return emptyStock, err
	}
	if duplicateStock != nil && duplicateStock.DeletedAt.Time.IsZero() {
		return emptyStock, duplicateNameOnRestore(existingStock.Name)
	}

	// Restore (ลบค่า DeletedAt)
	restoredStock, err := s.repository.RestoreStock(ctx, stockId)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return emptyStock, duplicateNameOnRestore(existingStock.Name)
		}
		return emptyStock, err
	}

	return restoredStock, nil
}
