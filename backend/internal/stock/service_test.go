package stock

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"backend/internal/currentuser"
	"backend/internal/models"
	_package "backend/package"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	validStockID  = "8f64ee52-1111-4222-8333-944455556666"
	validBranchID = "a1b2c3d4-1111-4222-8333-944455556666"
)

// ---------- Mock Repository ----------
// เก็บ function ที่แต่ละ test อยากให้ทำ ถ้าไม่ตั้งไว้จะคืนค่าว่าง (nil, nil)
// และนับจำนวนครั้งที่ถูกเรียก เพื่อเช็คว่า service เรียก repo ถูกตัวไหม

type mockRepo struct {
	getAllStock                   func(ctx context.Context) ([]models.Stock, error)
	getStockContainersByStockID   func(ctx context.Context, stockId, branchId, status string, page, pageSize int) ([]models.StockContainer, int64, error)
	findStockByName               func(ctx context.Context, name string) (*models.Stock, error)
	findStockById                 func(ctx context.Context, id string) (*models.Stock, error)
	createStock                   func(ctx context.Context, s models.Stock) (models.Stock, error)
	updateStock                   func(ctx context.Context, s models.Stock) (models.Stock, error)
	checkStockHasRelation         func(ctx context.Context, stockId string) (bool, error)
	softDeleteStock               func(ctx context.Context, stockId string) error
	hardDeleteStock               func(ctx context.Context, stockId string) error
	findStockByIdIncludingDeleted func(ctx context.Context, id string) (*models.Stock, error)
	restoreStock                  func(ctx context.Context, stockId string) (models.Stock, error)

	softDeleteCalls int
	hardDeleteCalls int
	createCalls     int
	updateCalls     int
	// ค่าที่ service ส่งเข้ามาล่าสุด ไว้เช็คว่าบันทึกอะไรลง DB
	lastCreated models.Stock
	lastUpdated models.Stock
	// ค่า args ล่าสุดของ GetStockContainersByStockID
	lastBranchID string
	lastStatus   string
	lastPage     int
	lastPageSize int
}

func (m *mockRepo) GetAllStock(ctx context.Context) ([]models.Stock, error) {
	if m.getAllStock != nil {
		return m.getAllStock(ctx)
	}
	return nil, nil
}

func (m *mockRepo) GetStockContainersByStockID(ctx context.Context, stockId, branchId, status string, page, pageSize int) ([]models.StockContainer, int64, error) {
	m.lastBranchID, m.lastStatus, m.lastPage, m.lastPageSize = branchId, status, page, pageSize
	if m.getStockContainersByStockID != nil {
		return m.getStockContainersByStockID(ctx, stockId, branchId, status, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockRepo) FindStockByName(ctx context.Context, name string) (*models.Stock, error) {
	if m.findStockByName != nil {
		return m.findStockByName(ctx, name)
	}
	return nil, nil
}

func (m *mockRepo) FindStockById(ctx context.Context, id string) (*models.Stock, error) {
	if m.findStockById != nil {
		return m.findStockById(ctx, id)
	}
	return nil, nil
}

func (m *mockRepo) CreateStock(ctx context.Context, s models.Stock) (models.Stock, error) {
	m.createCalls++
	m.lastCreated = s
	if m.createStock != nil {
		return m.createStock(ctx, s)
	}
	s.ID = validStockID
	return s, nil
}

func (m *mockRepo) UpdateStock(ctx context.Context, s models.Stock) (models.Stock, error) {
	m.updateCalls++
	m.lastUpdated = s
	if m.updateStock != nil {
		return m.updateStock(ctx, s)
	}
	return s, nil
}

func (m *mockRepo) CheckStockHasRelation(ctx context.Context, stockId string) (bool, error) {
	if m.checkStockHasRelation != nil {
		return m.checkStockHasRelation(ctx, stockId)
	}
	return false, nil
}

func (m *mockRepo) SoftDeleteStock(ctx context.Context, stockId string) error {
	m.softDeleteCalls++
	if m.softDeleteStock != nil {
		return m.softDeleteStock(ctx, stockId)
	}
	return nil
}

func (m *mockRepo) HardDeleteStock(ctx context.Context, stockId string) error {
	m.hardDeleteCalls++
	if m.hardDeleteStock != nil {
		return m.hardDeleteStock(ctx, stockId)
	}
	return nil
}

func (m *mockRepo) FindStockByIdIncludingDeleted(ctx context.Context, id string) (*models.Stock, error) {
	if m.findStockByIdIncludingDeleted != nil {
		return m.findStockByIdIncludingDeleted(ctx, id)
	}
	return nil, nil
}

func (m *mockRepo) RestoreStock(ctx context.Context, stockId string) (models.Stock, error) {
	if m.restoreStock != nil {
		return m.restoreStock(ctx, stockId)
	}
	return models.Stock{ID: stockId}, nil
}

// ---------- Helpers ----------

// requireAppError เช็คว่า err เป็น AppError และมี Code ตามที่คาด (Code แปลงเป็น HTTP status ที่ SendError)
func requireAppError(t *testing.T, err error, code string) *_package.AppError {
	t.Helper()
	var appErr *_package.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
	return appErr
}

func validAddRequest() AddStockRequest {
	return AddStockRequest{
		StockName:   "ครีมกันแดด",
		Price:       decimal.NewFromInt(500),
		IsSellable:  true,
		Category:    "premium",
		UsesPerUnit: 10,
		MinQty:      5,
	}
}

func activeStock() *models.Stock {
	return &models.Stock{
		ID:          validStockID,
		Name:        "ครีมกันแดด",
		Price:       decimal.NewFromInt(500),
		IsSellable:  true,
		Category:    "premium",
		UsesPerUnit: 10,
		MinQty:      5,
		IsActive:    true,
	}
}

func deletedStock() *models.Stock {
	s := activeStock()
	s.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	return s
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }

// ---------- AddStock ----------

func TestAddStock_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	req := validAddRequest()
	req.StockName = "  ครีมกันแดด  " // ต้องถูก trim

	got, err := svc.AddStock(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "ครีมกันแดด", got.Name)
	assert.True(t, got.IsActive, "สินค้าใหม่ต้อง active")
	assert.Equal(t, 1, repo.createCalls)
}

func TestAddStock_CategoryIsNormalizedToLowercase(t *testing.T) {
	repo := &mockRepo{}
	req := validAddRequest()
	req.Category = "  PREMIUM "

	_, err := NewService(repo).AddStock(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "premium", repo.lastCreated.Category)
}

func TestAddStock_NotSellableForcesPriceZero(t *testing.T) {
	repo := &mockRepo{}
	req := validAddRequest()
	req.IsSellable = false
	req.Price = decimal.NewFromInt(999)

	_, err := NewService(repo).AddStock(context.Background(), req)

	require.NoError(t, err)
	assert.True(t, repo.lastCreated.Price.IsZero(), "ไม่ขาย = ราคาต้องเป็น 0")
}

func TestAddStock_Validation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(r *AddStockRequest)
		detail string
	}{
		{"ชื่อว่าง", func(r *AddStockRequest) { r.StockName = "" }, "stock name is required"},
		{"ชื่อเป็นช่องว่างล้วน", func(r *AddStockRequest) { r.StockName = "   " }, "stock name is required"},
		{"ชื่อยาวเกิน 150 ตัวอักษร", func(r *AddStockRequest) { r.StockName = strings.Repeat("a", 151) }, "must not exceed 150"},
		{"category ผิด", func(r *AddStockRequest) { r.Category = "luxury" }, "category must be"},
		{"category ว่าง", func(r *AddStockRequest) { r.Category = "" }, "category must be"},
		{"ราคาติดลบ", func(r *AddStockRequest) { r.Price = decimal.NewFromInt(-1) }, "price must be"},
		{"usesPerUnit = 0", func(r *AddStockRequest) { r.UsesPerUnit = 0 }, "uses per unit"},
		{"minQty ติดลบ", func(r *AddStockRequest) { r.MinQty = -1 }, "min quantity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			req := validAddRequest()
			tt.mutate(&req)

			_, err := NewService(repo).AddStock(context.Background(), req)

			appErr := requireAppError(t, err, _package.CodeValidation)
			assert.Contains(t, appErr.Detail, tt.detail)
			assert.Equal(t, 0, repo.createCalls, "ผิด validation ห้ามไปถึง DB")
		})
	}
}

// ชื่อไทย 150 ตัวอักษรต้องผ่าน (len() นับ byte จะได้ 450 แล้วโดนปฏิเสธผิด)
func TestAddStock_NameLengthCountsCharactersNotBytes(t *testing.T) {
	t.Run("ไทย 150 ตัวอักษร ผ่าน", func(t *testing.T) {
		req := validAddRequest()
		req.StockName = strings.Repeat("ก", 150)

		_, err := NewService(&mockRepo{}).AddStock(context.Background(), req)

		assert.NoError(t, err)
	})
	t.Run("ไทย 151 ตัวอักษร ไม่ผ่าน", func(t *testing.T) {
		req := validAddRequest()
		req.StockName = strings.Repeat("ก", 151)

		_, err := NewService(&mockRepo{}).AddStock(context.Background(), req)

		requireAppError(t, err, _package.CodeValidation)
	})
}

func TestAddStock_DuplicateActiveName_Returns409(t *testing.T) {
	repo := &mockRepo{
		findStockByName: func(ctx context.Context, name string) (*models.Stock, error) {
			return activeStock(), nil
		},
	}

	_, err := NewService(repo).AddStock(context.Background(), validAddRequest())

	appErr := requireAppError(t, err, _package.CodeConflict)
	assert.Equal(t, "Stock name already exists", appErr.Message)
	assert.Equal(t, 0, repo.createCalls)
}

// ชื่อซ้ำกับของที่ถูก soft delete: ปัจจุบัน service ปล่อยผ่านไปสร้างต่อ
func TestAddStock_NameOfSoftDeletedStock_IsAllowedByService(t *testing.T) {
	repo := &mockRepo{
		findStockByName: func(ctx context.Context, name string) (*models.Stock, error) {
			return deletedStock(), nil
		},
	}

	_, err := NewService(repo).AddStock(context.Background(), validAddRequest())

	assert.NoError(t, err)
	assert.Equal(t, 1, repo.createCalls)
}

// สองคำขอพร้อมกันผ่านการเช็คชื่อทั้งคู่ unique index ของ DB กันไว้ ต้องตอบ 409 ไม่ใช่ 500
func TestAddStock_ConcurrentDuplicate_DBUniqueViolation_Returns409(t *testing.T) {
	repo := &mockRepo{
		createStock: func(ctx context.Context, s models.Stock) (models.Stock, error) {
			return models.Stock{}, gorm.ErrDuplicatedKey
		},
	}

	_, err := NewService(repo).AddStock(context.Background(), validAddRequest())

	requireAppError(t, err, _package.CodeConflict)
}

func TestAddStock_RepositoryErrorsArePassedThrough(t *testing.T) {
	dbErr := errors.New("connection refused")

	t.Run("FindStockByName พัง", func(t *testing.T) {
		repo := &mockRepo{findStockByName: func(ctx context.Context, name string) (*models.Stock, error) {
			return nil, dbErr
		}}
		_, err := NewService(repo).AddStock(context.Background(), validAddRequest())
		assert.ErrorIs(t, err, dbErr)
	})
	t.Run("CreateStock พัง", func(t *testing.T) {
		repo := &mockRepo{createStock: func(ctx context.Context, s models.Stock) (models.Stock, error) {
			return models.Stock{}, dbErr
		}}
		_, err := NewService(repo).AddStock(context.Background(), validAddRequest())
		assert.ErrorIs(t, err, dbErr)
	})
}

// ---------- UpdateStockById ----------

func repoWithStock(s *models.Stock) *mockRepo {
	return &mockRepo{
		findStockById: func(ctx context.Context, id string) (*models.Stock, error) {
			// คืนสำเนา กัน service แก้ค่าต้นฉบับที่ test ใช้เทียบ
			c := *s
			return &c, nil
		},
	}
}

func TestUpdateStock_InvalidUUID(t *testing.T) {
	_, err := NewService(&mockRepo{}).UpdateStockById(context.Background(), "not-a-uuid", UpdateStockRequest{})

	requireAppError(t, err, _package.CodeValidation)
}

func TestUpdateStock_NotFound(t *testing.T) {
	_, err := NewService(&mockRepo{}).UpdateStockById(context.Background(), validStockID, UpdateStockRequest{})

	requireAppError(t, err, _package.CodeNotFound)
}

func TestUpdateStock_OnlySentFieldsChange(t *testing.T) {
	repo := repoWithStock(activeStock())

	_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
		UpdateStockRequest{MinQty: intPtr(20)})

	require.NoError(t, err)
	assert.Equal(t, 20, repo.lastUpdated.MinQty)
	assert.Equal(t, "ครีมกันแดด", repo.lastUpdated.Name, "ฟิลด์ที่ไม่ได้ส่งต้องไม่เปลี่ยน")
	assert.True(t, repo.lastUpdated.Price.Equal(decimal.NewFromInt(500)), "ราคาเดิมต้องไม่ถูกเขียนทับเป็น 0")
}

func TestUpdateStock_PriceSet_UpdatesPrice(t *testing.T) {
	repo := repoWithStock(activeStock())

	_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
		UpdateStockRequest{Price: decimal.NewFromInt(750), PriceSet: true})

	require.NoError(t, err)
	assert.True(t, repo.lastUpdated.Price.Equal(decimal.NewFromInt(750)))
}

func TestUpdateStock_NegativePrice(t *testing.T) {
	repo := repoWithStock(activeStock())

	_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
		UpdateStockRequest{Price: decimal.NewFromInt(-5), PriceSet: true})

	requireAppError(t, err, _package.CodeValidation)
	assert.Equal(t, 0, repo.updateCalls)
}

func TestUpdateStock_SetNotSellable_ForcesPriceZero(t *testing.T) {
	repo := repoWithStock(activeStock())

	_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
		UpdateStockRequest{IsSellable: boolPtr(false)})

	require.NoError(t, err)
	assert.False(t, repo.lastUpdated.IsSellable)
	assert.True(t, repo.lastUpdated.Price.IsZero())
}

func TestUpdateStock_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  UpdateStockRequest
	}{
		{"ชื่อว่าง", UpdateStockRequest{StockName: strPtr("  ")}},
		{"ชื่อยาวเกิน", UpdateStockRequest{StockName: strPtr(strings.Repeat("ก", 151))}},
		{"category ผิด", UpdateStockRequest{Category: strPtr("luxury")}},
		{"usesPerUnit = 0", UpdateStockRequest{UsesPerUnit: intPtr(0)}},
		{"minQty ติดลบ", UpdateStockRequest{MinQty: intPtr(-1)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repoWithStock(activeStock())

			_, err := NewService(repo).UpdateStockById(context.Background(), validStockID, tt.req)

			requireAppError(t, err, _package.CodeValidation)
			assert.Equal(t, 0, repo.updateCalls)
		})
	}
}

func TestUpdateStock_RenameToExistingActiveName_Returns409(t *testing.T) {
	repo := repoWithStock(activeStock())
	repo.findStockByName = func(ctx context.Context, name string) (*models.Stock, error) {
		other := activeStock()
		other.ID = "ffffffff-1111-4222-8333-944455556666"
		other.Name = name
		return other, nil
	}

	_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
		UpdateStockRequest{StockName: strPtr("ชื่อที่มีคนใช้แล้ว")})

	requireAppError(t, err, _package.CodeConflict)
	assert.Equal(t, 0, repo.updateCalls)
}

func TestUpdateStock_SameNameAsItself_IsNotConflict(t *testing.T) {
	repo := repoWithStock(activeStock())
	called := false
	repo.findStockByName = func(ctx context.Context, name string) (*models.Stock, error) {
		called = true
		return activeStock(), nil
	}

	_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
		UpdateStockRequest{StockName: strPtr("ครีมกันแดด")})

	require.NoError(t, err)
	assert.False(t, called, "ชื่อไม่เปลี่ยนไม่ต้องไปเช็คชื่อซ้ำ")
}

func TestUpdateStock_ConcurrentDuplicate_DBUniqueViolation_Returns409(t *testing.T) {
	repo := repoWithStock(activeStock())
	repo.updateStock = func(ctx context.Context, s models.Stock) (models.Stock, error) {
		return models.Stock{}, gorm.ErrDuplicatedKey
	}

	_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
		UpdateStockRequest{StockName: strPtr("ชื่อใหม่")})

	requireAppError(t, err, _package.CodeConflict)
}

// ---------- DeleteStockById ----------

func TestDeleteStock_InvalidUUID(t *testing.T) {
	err := NewService(&mockRepo{}).DeleteStockById(context.Background(), "xxx")

	requireAppError(t, err, _package.CodeValidation)
}

func TestDeleteStock_NotFound(t *testing.T) {
	err := NewService(&mockRepo{}).DeleteStockById(context.Background(), validStockID)

	requireAppError(t, err, _package.CodeNotFound)
}

func TestDeleteStock_WithRelation_SoftDeletes(t *testing.T) {
	repo := repoWithStock(activeStock())
	repo.checkStockHasRelation = func(ctx context.Context, id string) (bool, error) { return true, nil }

	err := NewService(repo).DeleteStockById(context.Background(), validStockID)

	require.NoError(t, err)
	assert.Equal(t, 1, repo.softDeleteCalls)
	assert.Equal(t, 0, repo.hardDeleteCalls)
}

func TestDeleteStock_WithoutRelation_HardDeletes(t *testing.T) {
	repo := repoWithStock(activeStock())

	err := NewService(repo).DeleteStockById(context.Background(), validStockID)

	require.NoError(t, err)
	assert.Equal(t, 1, repo.hardDeleteCalls)
	assert.Equal(t, 0, repo.softDeleteCalls)
}

func TestDeleteStock_RelationCheckFails_DoesNotDelete(t *testing.T) {
	dbErr := errors.New("db down")
	repo := repoWithStock(activeStock())
	repo.checkStockHasRelation = func(ctx context.Context, id string) (bool, error) { return false, dbErr }

	err := NewService(repo).DeleteStockById(context.Background(), validStockID)

	assert.ErrorIs(t, err, dbErr)
	assert.Equal(t, 0, repo.hardDeleteCalls+repo.softDeleteCalls)
}

// ---------- RestoreStockById ----------

func repoWithDeletedStock() *mockRepo {
	return &mockRepo{
		findStockByIdIncludingDeleted: func(ctx context.Context, id string) (*models.Stock, error) {
			return deletedStock(), nil
		},
	}
}

func TestRestoreStock_InvalidUUID(t *testing.T) {
	_, err := NewService(&mockRepo{}).RestoreStockById(context.Background(), "xxx")

	requireAppError(t, err, _package.CodeValidation)
}

func TestRestoreStock_NotFound(t *testing.T) {
	_, err := NewService(&mockRepo{}).RestoreStockById(context.Background(), validStockID)

	requireAppError(t, err, _package.CodeNotFound)
}

func TestRestoreStock_NotDeleted(t *testing.T) {
	repo := &mockRepo{
		findStockByIdIncludingDeleted: func(ctx context.Context, id string) (*models.Stock, error) {
			return activeStock(), nil
		},
	}

	_, err := NewService(repo).RestoreStockById(context.Background(), validStockID)

	appErr := requireAppError(t, err, _package.CodeValidation)
	assert.Equal(t, "Stock is not deleted", appErr.Message)
}

func TestRestoreStock_NameTakenByActiveStock_Returns409(t *testing.T) {
	repo := repoWithDeletedStock()
	repo.findStockByName = func(ctx context.Context, name string) (*models.Stock, error) {
		return activeStock(), nil
	}

	_, err := NewService(repo).RestoreStockById(context.Background(), validStockID)

	appErr := requireAppError(t, err, _package.CodeConflict)
	assert.Contains(t, appErr.Detail, "Cannot restore")
}

func TestRestoreStock_Success(t *testing.T) {
	repo := repoWithDeletedStock()
	repo.restoreStock = func(ctx context.Context, id string) (models.Stock, error) {
		return *activeStock(), nil
	}

	got, err := NewService(repo).RestoreStockById(context.Background(), validStockID)

	require.NoError(t, err)
	assert.False(t, got.DeletedAt.Valid)
}

func TestRestoreStock_ConcurrentDuplicate_DBUniqueViolation_Returns409(t *testing.T) {
	repo := repoWithDeletedStock()
	repo.restoreStock = func(ctx context.Context, id string) (models.Stock, error) {
		return models.Stock{}, gorm.ErrDuplicatedKey
	}

	_, err := NewService(repo).RestoreStockById(context.Background(), validStockID)

	requireAppError(t, err, _package.CodeConflict)
}

// ---------- GetStockContainersByStockID ----------

// helper สร้าง CurrentUser สำหรับ test
func adminUser() *currentuser.CurrentUser {
	return &currentuser.CurrentUser{
		ID:       "admin-id",
		Username: "admin",
		Role:     "admin",
		BranchID: nil,
	}
}

func branchUser(branchID string) *currentuser.CurrentUser {
	return &currentuser.CurrentUser{
		ID:       "branch-user-id",
		Username: "branch",
		Role:     "branch",
		BranchID: &branchID,
	}
}

func TestGetContainers_Validation(t *testing.T) {
	tests := []struct {
		name string
		user *currentuser.CurrentUser
		q    ContainerQuery
	}{
		{"stock id ไม่ใช่ UUID", adminUser(), ContainerQuery{StockID: "abc"}},
		{"status ผิด", adminUser(), ContainerQuery{StockID: validStockID, Status: "weird"}},
		{"page = 0 ติดลบ", adminUser(), ContainerQuery{StockID: validStockID, Page: "-1"}},
		{"page ไม่ใช่ตัวเลข", adminUser(), ContainerQuery{StockID: validStockID, Page: "abc"}},
		{"page_size เกิน 100", adminUser(), ContainerQuery{StockID: validStockID, PageSize: "101"}},
		{"admin ส่ง branch_id ไม่ใช่ UUID", adminUser(), ContainerQuery{StockID: validStockID, BranchID: "zzz"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := NewService(&mockRepo{}).GetStockContainersByStockID(context.Background(), tt.user, tt.q)

			requireAppError(t, err, _package.CodeValidation)
		})
	}
}

func TestGetContainers_Defaults(t *testing.T) {
	repo := &mockRepo{}

	_, meta, err := NewService(repo).GetStockContainersByStockID(context.Background(), adminUser(),
		ContainerQuery{StockID: validStockID})

	require.NoError(t, err)
	assert.Equal(t, StatusOpen, repo.lastStatus)
	assert.Equal(t, 1, repo.lastPage)
	assert.Equal(t, 20, repo.lastPageSize)
	assert.Equal(t, "", repo.lastBranchID, "admin ไม่ส่ง branch_id = ทุกสาขา")
	assert.Equal(t, 1, meta.Page)
}

func TestGetContainers_StatusIsCaseInsensitive(t *testing.T) {
	repo := &mockRepo{}

	_, _, err := NewService(repo).GetStockContainersByStockID(context.Background(), adminUser(),
		ContainerQuery{StockID: validStockID, Status: " CLOSED "})

	require.NoError(t, err)
	assert.Equal(t, StatusClosed, repo.lastStatus)
}

func TestGetContainers_AdminCanPickBranch(t *testing.T) {
	repo := &mockRepo{}

	_, _, err := NewService(repo).GetStockContainersByStockID(context.Background(), adminUser(),
		ContainerQuery{StockID: validStockID, BranchID: validBranchID})

	require.NoError(t, err)
	assert.Equal(t, validBranchID, repo.lastBranchID)
}

// บัญชีสาขาห้ามดูข้ามสาขา: ต้องถูกบังคับเป็นสาขาของตัวเองเสมอ
func TestGetContainers_BranchUserIsForcedToOwnBranch(t *testing.T) {
	repo := &mockRepo{}
	otherBranch := "bbbbbbbb-1111-4222-8333-944455556666"

	_, _, err := NewService(repo).GetStockContainersByStockID(context.Background(), branchUser(validBranchID),
		ContainerQuery{StockID: validStockID, BranchID: otherBranch})

	require.NoError(t, err)
	assert.Equal(t, validBranchID, repo.lastBranchID)
}

func TestGetContainers_BranchUserWithoutBranch_Forbidden(t *testing.T) {
	emptyBranch := ""
	_, _, err := NewService(&mockRepo{}).GetStockContainersByStockID(context.Background(), branchUser(emptyBranch),
		ContainerQuery{StockID: validStockID})

	assert.Same(t, ErrBranchRequired, err)
}

func TestGetContainers_PageMeta(t *testing.T) {
	repo := &mockRepo{
		getStockContainersByStockID: func(ctx context.Context, stockId, branchId, status string, page, pageSize int) ([]models.StockContainer, int64, error) {
			return []models.StockContainer{}, 45, nil
		},
	}

	_, meta, err := NewService(repo).GetStockContainersByStockID(context.Background(), adminUser(),
		ContainerQuery{StockID: validStockID, Page: "2", PageSize: "20"})

	require.NoError(t, err)
	assert.Equal(t, int64(45), meta.Total)
	assert.Equal(t, 3, meta.TotalPages)
	assert.Equal(t, 2, meta.Page)
}

// ---------- parseBoundedInt ----------

func TestParseBoundedInt(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    int
		wantErr bool
	}{
		{"ว่าง = ค่าเริ่มต้น", "", 20, false},
		{"ช่องว่างล้วน = ค่าเริ่มต้น", "   ", 20, false},
		{"ค่าปกติ", "5", 5, false},
		{"ขอบล่าง", "1", 1, false},
		{"ขอบบน", "100", 100, false},
		{"เกินขอบบน", "101", 0, true},
		{"ศูนย์", "0", 0, true},
		{"ติดลบ", "-3", 0, true},
		{"ไม่ใช่ตัวเลข", "abc", 0, true},
		{"ทศนิยม", "2.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBoundedInt("page_size", tt.raw, 20, 100)

			if tt.wantErr {
				requireAppError(t, err, _package.CodeValidation)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------- เพิ่มเติม: error path ของ repository และกรณีขอบ ----------

var errDB = errors.New("db down")

func TestGetAllStock_Service(t *testing.T) {
	t.Run("ส่งผลจาก repo ต่อ", func(t *testing.T) {
		repo := &mockRepo{getAllStock: func(ctx context.Context) ([]models.Stock, error) {
			return []models.Stock{*activeStock()}, nil
		}}

		got, err := NewService(repo).GetAllStock(context.Background())

		require.NoError(t, err)
		assert.Len(t, got, 1)
	})
	t.Run("repo พัง ส่ง error ต่อ", func(t *testing.T) {
		repo := &mockRepo{getAllStock: func(ctx context.Context) ([]models.Stock, error) { return nil, errDB }}

		_, err := NewService(repo).GetAllStock(context.Background())

		assert.ErrorIs(t, err, errDB)
	})
}

func TestUpdateStock_UpdatesEveryField(t *testing.T) {
	repo := repoWithStock(activeStock())

	got, err := NewService(repo).UpdateStockById(context.Background(), validStockID, UpdateStockRequest{
		StockName:   strPtr("  ชื่อใหม่  "),
		Price:       decimal.NewFromInt(900),
		PriceSet:    true,
		IsSellable:  boolPtr(true),
		Category:    strPtr(" NORMAL "),
		UsesPerUnit: intPtr(4),
		MinQty:      intPtr(0),
		IsActive:    boolPtr(false),
	})

	require.NoError(t, err)
	assert.Equal(t, "ชื่อใหม่", got.Name, "ต้อง trim ชื่อ")
	assert.True(t, got.Price.Equal(decimal.NewFromInt(900)))
	assert.Equal(t, "normal", got.Category, "category ต้องเป็นตัวเล็ก")
	assert.Equal(t, 4, got.UsesPerUnit)
	assert.Equal(t, 0, got.MinQty, "minQty = 0 ต้องบันทึกได้")
	assert.False(t, got.IsActive)
}

// ราคาที่ส่งมาตอนสินค้าไม่ขาย ต้องถูกบังคับเป็น 0 เสมอ (ทำท้ายสุด)
func TestUpdateStock_PriceOnNotSellableStock_StaysZero(t *testing.T) {
	s := activeStock()
	s.IsSellable = false
	s.Price = decimal.Zero
	repo := repoWithStock(s)

	_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
		UpdateStockRequest{Price: decimal.NewFromInt(300), PriceSet: true})

	require.NoError(t, err)
	assert.True(t, repo.lastUpdated.Price.IsZero())
}

func TestUpdateStock_PriceZeroIsAllowed(t *testing.T) {
	repo := repoWithStock(activeStock())

	_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
		UpdateStockRequest{Price: decimal.Zero, PriceSet: true})

	require.NoError(t, err)
	assert.True(t, repo.lastUpdated.Price.IsZero())
}

// ชื่อซ้ำกับสินค้าที่ถูก soft delete ไปแล้ว เปลี่ยนชื่อมาใช้ได้
// (unique index idx_stocks_name_active ครอบคลุมเฉพาะ deleted_at IS NULL)
func TestUpdateStock_RenameToSoftDeletedName_IsAllowed(t *testing.T) {
	repo := repoWithStock(activeStock())
	repo.findStockByName = func(ctx context.Context, name string) (*models.Stock, error) {
		other := deletedStock()
		other.ID = "ffffffff-1111-4222-8333-944455556666"
		return other, nil
	}

	_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
		UpdateStockRequest{StockName: strPtr("ชื่อเก่าที่ถูกลบ")})

	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
}

func TestUpdateStock_RepositoryErrors(t *testing.T) {
	t.Run("FindStockById พัง", func(t *testing.T) {
		repo := &mockRepo{findStockById: func(ctx context.Context, id string) (*models.Stock, error) { return nil, errDB }}

		_, err := NewService(repo).UpdateStockById(context.Background(), validStockID, UpdateStockRequest{})

		assert.ErrorIs(t, err, errDB)
	})
	t.Run("FindStockByName พัง", func(t *testing.T) {
		repo := repoWithStock(activeStock())
		repo.findStockByName = func(ctx context.Context, name string) (*models.Stock, error) { return nil, errDB }

		_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
			UpdateStockRequest{StockName: strPtr("ชื่อใหม่")})

		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, 0, repo.updateCalls)
	})
	t.Run("UpdateStock พัง (ไม่ใช่ชื่อซ้ำ) ต้องไม่ถูกแปลงเป็น 409", func(t *testing.T) {
		repo := repoWithStock(activeStock())
		repo.updateStock = func(ctx context.Context, s models.Stock) (models.Stock, error) {
			return models.Stock{}, errDB
		}

		_, err := NewService(repo).UpdateStockById(context.Background(), validStockID,
			UpdateStockRequest{MinQty: intPtr(1)})

		assert.ErrorIs(t, err, errDB)
		var appErr *_package.AppError
		assert.False(t, errors.As(err, &appErr))
	})
}

func TestDeleteStock_RepositoryErrors(t *testing.T) {
	t.Run("FindStockById พัง", func(t *testing.T) {
		repo := &mockRepo{findStockById: func(ctx context.Context, id string) (*models.Stock, error) { return nil, errDB }}

		err := NewService(repo).DeleteStockById(context.Background(), validStockID)

		assert.ErrorIs(t, err, errDB)
	})
	t.Run("SoftDelete พัง", func(t *testing.T) {
		repo := repoWithStock(activeStock())
		repo.checkStockHasRelation = func(ctx context.Context, id string) (bool, error) { return true, nil }
		repo.softDeleteStock = func(ctx context.Context, id string) error { return errDB }

		err := NewService(repo).DeleteStockById(context.Background(), validStockID)

		assert.ErrorIs(t, err, errDB)
	})
	t.Run("HardDelete พัง", func(t *testing.T) {
		repo := repoWithStock(activeStock())
		repo.hardDeleteStock = func(ctx context.Context, id string) error { return errDB }

		err := NewService(repo).DeleteStockById(context.Background(), validStockID)

		assert.ErrorIs(t, err, errDB)
	})
}

func TestRestoreStock_RepositoryErrors(t *testing.T) {
	t.Run("FindStockByIdIncludingDeleted พัง", func(t *testing.T) {
		repo := &mockRepo{findStockByIdIncludingDeleted: func(ctx context.Context, id string) (*models.Stock, error) {
			return nil, errDB
		}}

		_, err := NewService(repo).RestoreStockById(context.Background(), validStockID)

		assert.ErrorIs(t, err, errDB)
	})
	t.Run("FindStockByName พัง", func(t *testing.T) {
		repo := repoWithDeletedStock()
		repo.findStockByName = func(ctx context.Context, name string) (*models.Stock, error) { return nil, errDB }

		_, err := NewService(repo).RestoreStockById(context.Background(), validStockID)

		assert.ErrorIs(t, err, errDB)
	})
	t.Run("RestoreStock พัง (ไม่ใช่ชื่อซ้ำ) ต้องไม่ถูกแปลงเป็น 409", func(t *testing.T) {
		repo := repoWithDeletedStock()
		repo.restoreStock = func(ctx context.Context, id string) (models.Stock, error) {
			return models.Stock{}, errDB
		}

		_, err := NewService(repo).RestoreStockById(context.Background(), validStockID)

		assert.ErrorIs(t, err, errDB)
		var appErr *_package.AppError
		assert.False(t, errors.As(err, &appErr))
	})
}

// ชื่อซ้ำกับสินค้าที่ลบแล้วด้วยกัน ไม่นับว่าชน (กู้คืนได้)
func TestRestoreStock_NameMatchesAnotherDeletedStock_IsAllowed(t *testing.T) {
	repo := repoWithDeletedStock()
	repo.findStockByName = func(ctx context.Context, name string) (*models.Stock, error) {
		return deletedStock(), nil
	}

	_, err := NewService(repo).RestoreStockById(context.Background(), validStockID)

	assert.NoError(t, err)
}

func TestGetContainers_RepositoryError(t *testing.T) {
	repo := &mockRepo{getStockContainersByStockID: func(ctx context.Context, stockId, branchId, status string, page, pageSize int) ([]models.StockContainer, int64, error) {
		return nil, 0, errDB
	}}

	_, _, err := NewService(repo).GetStockContainersByStockID(context.Background(), adminUser(),
		ContainerQuery{StockID: validStockID})

	assert.ErrorIs(t, err, errDB)
}

func TestGetContainers_ExplicitStatusAllAndOpen(t *testing.T) {
	for _, status := range []string{StatusOpen, StatusClosed, StatusAll} {
		t.Run(status, func(t *testing.T) {
			repo := &mockRepo{}

			_, _, err := NewService(repo).GetStockContainersByStockID(context.Background(), adminUser(),
				ContainerQuery{StockID: validStockID, Status: status})

			require.NoError(t, err)
			assert.Equal(t, status, repo.lastStatus)
		})
	}
}

// admin ส่ง branchId เป็นช่องว่างล้วน = ไม่กรองสาขา (ไม่ใช่ error)
func TestGetContainers_AdminBlankBranchIsTrimmed(t *testing.T) {
	repo := &mockRepo{}

	_, _, err := NewService(repo).GetStockContainersByStockID(context.Background(), adminUser(),
		ContainerQuery{StockID: validStockID, BranchID: "   "})

	require.NoError(t, err)
	assert.Equal(t, "", repo.lastBranchID)
}

func TestGetContainers_PageUpperBound(t *testing.T) {
	_, _, err := NewService(&mockRepo{}).GetStockContainersByStockID(context.Background(), adminUser(),
		ContainerQuery{StockID: validStockID, Page: "1000001"})

	requireAppError(t, err, _package.CodeValidation)
}
