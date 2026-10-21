package stocklog

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"backend/internal/models"
	_package "backend/package"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validStockID     = "8f64ee52-1111-4222-8333-944455556666"
	validBranchID    = "a1b2c3d4-1111-4222-8333-944455556666"
	otherBranchID    = "b2c3d4e5-1111-4222-8333-944455556666"
	validOrderItemID = "c3d4e5f6-1111-4222-8333-944455556666"
	validUsageID     = "d4e5f6a7-1111-4222-8333-944455556666"
	validEmployeeID  = "e5f6a7b8-1111-4222-8333-944455556666"
)

// ---------- Mock Repository ----------
// เก็บ filter ที่ service ส่งเข้ามา เพื่อเช็คว่า service แปลง query เป็นเงื่อนไขถูกไหม

type mockRepo struct {
	quantities func(ctx context.Context, f quantityFilter) ([]StockQuantity, int64, error)
	logs       func(ctx context.Context, f stockLogFilter) ([]models.StockLog, int64, error)

	quantityCalls int
	logCalls      int
	lastQuantity  quantityFilter
	lastLog       stockLogFilter
}

func (m *mockRepo) GetAllStockQuantities(ctx context.Context, f quantityFilter) ([]StockQuantity, int64, error) {
	m.quantityCalls++
	m.lastQuantity = f
	if m.quantities != nil {
		return m.quantities(ctx, f)
	}
	return nil, 0, nil
}

func (m *mockRepo) GetStockLogs(ctx context.Context, f stockLogFilter) ([]models.StockLog, int64, error) {
	m.logCalls++
	m.lastLog = f
	if m.logs != nil {
		return m.logs(ctx, f)
	}
	return nil, 0, nil
}

// requireAppError เช็คว่าเป็น AppError ที่มี code ตามที่คาด
func requireAppError(t *testing.T, err error, code string) *_package.AppError {
	t.Helper()
	require.Error(t, err)
	var appErr *_package.AppError
	require.True(t, errors.As(err, &appErr), "ต้องเป็น AppError แต่ได้ %T: %v", err, err)
	assert.Equal(t, code, appErr.Code)
	return appErr
}

// ---------- GetAllStockQuantities: สิทธิ์/สาขา ----------

func TestQuantities_AdminWithoutBranch_SeesAllBranches(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	_, _, err := svc.GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{})
	require.NoError(t, err)
	assert.Equal(t, "", repo.lastQuantity.BranchID, "ว่าง = ทุกสาขา")
}

func TestQuantities_AdminCanPickBranch(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	_, _, err := svc.GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{BranchID: "  " + validBranchID + " "})
	require.NoError(t, err)
	assert.Equal(t, validBranchID, repo.lastQuantity.BranchID, "ต้อง trim ช่องว่าง")
}

func TestQuantities_AdminInvalidBranchUUID_Returns400(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	_, _, err := svc.GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{BranchID: "not-a-uuid"})
	requireAppError(t, err, _package.CodeValidation)
	assert.Zero(t, repo.quantityCalls, "ห้ามถึง repo ถ้า validate ไม่ผ่าน")
}

func TestQuantities_BranchUserForcedToOwnBranch(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	// พยายามดูข้ามสาขาด้วย ?branchId= ต้องถูกเมินและใช้สาขาจาก session
	_, _, err := svc.GetAllStockQuantities(context.Background(), "branch", validBranchID, GetAllStockQuantitiesQuery{BranchID: otherBranchID})
	require.NoError(t, err)
	assert.Equal(t, validBranchID, repo.lastQuantity.BranchID)
}

func TestQuantities_BranchUserIgnoresGarbageBranchQuery(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	// ค่า branchId ของบัญชีสาขาไม่ถูกใช้ จึงไม่ต้อง validate (ไม่ควร 400)
	_, _, err := svc.GetAllStockQuantities(context.Background(), "branch", validBranchID, GetAllStockQuantitiesQuery{BranchID: "garbage"})
	require.NoError(t, err)
	assert.Equal(t, validBranchID, repo.lastQuantity.BranchID)
}

func TestQuantities_BranchUserWithoutSessionBranch_Returns403(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	_, _, err := svc.GetAllStockQuantities(context.Background(), "branch", "", GetAllStockQuantitiesQuery{})
	requireAppError(t, err, _package.CodeForbidden)
	assert.Zero(t, repo.quantityCalls)
}

// ---------- GetAllStockQuantities: validate/normalize ----------

func TestQuantities_Category(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"ว่าง = ทั้งหมด", "", "", false},
		{"premium", "premium", "premium", false},
		{"normal", "normal", "normal", false},
		{"ตัวพิมพ์ใหญ่และช่องว่าง", "  PREMIUM ", "premium", false},
		{"ไม่รู้จัก", "luxury", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			_, _, err := NewService(repo).GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{Category: tt.in})
			if tt.wantErr {
				requireAppError(t, err, _package.CodeValidation)
				assert.Zero(t, repo.quantityCalls)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, repo.lastQuantity.Category)
		})
	}
}

func TestQuantities_Active(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"ว่าง = true", "", activeTrue, false},
		{"true", "true", activeTrue, false},
		{"false", "false", activeFalse, false},
		{"all", "all", activeAll, false},
		{"ตัวพิมพ์ใหญ่", " ALL ", activeAll, false},
		{"ไม่รู้จัก", "yes", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			_, _, err := NewService(repo).GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{Active: tt.in})
			if tt.wantErr {
				requireAppError(t, err, _package.CodeValidation)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, repo.lastQuantity.Active)
		})
	}
}

func TestQuantities_Search(t *testing.T) {
	t.Run("trim ช่องว่าง", func(t *testing.T) {
		repo := &mockRepo{}
		_, _, err := NewService(repo).GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{Search: "  ครีม  "})
		require.NoError(t, err)
		assert.Equal(t, "ครีม", repo.lastQuantity.Search)
	})

	t.Run("ภาษาไทย 100 ตัวอักษรผ่าน (นับเป็นตัวอักษร ไม่ใช่ byte)", func(t *testing.T) {
		repo := &mockRepo{}
		_, _, err := NewService(repo).GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{Search: strings.Repeat("ก", maxSearchLen)})
		require.NoError(t, err)
	})

	t.Run("101 ตัวอักษรไม่ผ่าน", func(t *testing.T) {
		repo := &mockRepo{}
		_, _, err := NewService(repo).GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{Search: strings.Repeat("ก", maxSearchLen+1)})
		requireAppError(t, err, _package.CodeValidation)
		assert.Zero(t, repo.quantityCalls)
	})
}

func TestQuantities_LowStockFlagPassedThrough(t *testing.T) {
	repo := &mockRepo{}
	_, _, err := NewService(repo).GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{LowStock: true})
	require.NoError(t, err)
	assert.True(t, repo.lastQuantity.LowStock)
}

func TestQuantities_Paging(t *testing.T) {
	tests := []struct {
		name         string
		page, size   int
		wantPage     int
		wantSize     int
		wantErr      bool
		wantErrField string
	}{
		{"ไม่ส่ง = ค่าเริ่มต้น", 0, 0, defaultPage, defaultPageSize, false, ""},
		{"ค่าปกติ", 3, 50, 3, 50, false, ""},
		{"pageSize สูงสุด", 1, maxPageSize, 1, maxPageSize, false, ""},
		{"pageSize เกินสูงสุด", 1, maxPageSize + 1, 0, 0, true, "pageSize"},
		{"pageSize ติดลบ", 1, -1, 0, 0, true, "pageSize"},
		{"page ติดลบ", -1, 10, 0, 0, true, "page"},
		{"page เกินสูงสุด", maxPage + 1, 10, 0, 0, true, "page"},
		{"page สูงสุด", maxPage, 10, maxPage, 10, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			_, meta, err := NewService(repo).GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{Page: tt.page, PageSize: tt.size})
			if tt.wantErr {
				appErr := requireAppError(t, err, _package.CodeValidation)
				assert.Contains(t, appErr.Detail, tt.wantErrField)
				assert.Zero(t, repo.quantityCalls)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPage, repo.lastQuantity.Page)
			assert.Equal(t, tt.wantSize, repo.lastQuantity.PageSize)
			assert.Equal(t, tt.wantPage, meta.Page)
			assert.Equal(t, tt.wantSize, meta.PageSize)
		})
	}
}

// ---------- GetAllStockQuantities: ผลลัพธ์ ----------

func TestQuantities_ComputesIsLowStock(t *testing.T) {
	repo := &mockRepo{
		quantities: func(ctx context.Context, f quantityFilter) ([]StockQuantity, int64, error) {
			return []StockQuantity{
				{StockName: "ต่ำกว่า", Quantity: 4, MinQty: 5},
				{StockName: "เท่ากัน", Quantity: 5, MinQty: 5},
				{StockName: "สูงกว่า", Quantity: 6, MinQty: 5},
				{StockName: "หมด", Quantity: 0, MinQty: 0},
			}, 4, nil
		},
	}

	items, meta, err := NewService(repo).GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{})
	require.NoError(t, err)
	require.Len(t, items, 4)
	assert.True(t, items[0].IsLowStock)
	assert.True(t, items[1].IsLowStock, "เท่ากับ min_qty ถือว่าต่ำ (<=)")
	assert.False(t, items[2].IsLowStock)
	assert.True(t, items[3].IsLowStock, "0 <= 0")
	assert.Equal(t, int64(4), meta.Total)
}

func TestQuantities_EmptyResultIsEmptySliceNotNil(t *testing.T) {
	items, meta, err := NewService(&mockRepo{}).GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{})
	require.NoError(t, err)
	assert.NotNil(t, items, "ต้องตอบ [] ไม่ใช่ null")
	assert.Empty(t, items)
	assert.Equal(t, 0, meta.TotalPages)
}

func TestQuantities_RepoErrorPropagates(t *testing.T) {
	boom := errors.New("db down")
	repo := &mockRepo{
		quantities: func(ctx context.Context, f quantityFilter) ([]StockQuantity, int64, error) { return nil, 0, boom },
	}
	_, _, err := NewService(repo).GetAllStockQuantities(context.Background(), "admin", "", GetAllStockQuantitiesQuery{})
	assert.ErrorIs(t, err, boom)
}

// ---------- GetStockLogsByBranchId: สิทธิ์/สาขา ----------

func TestLogs_AdminWithoutBranch_SeesAllBranches(t *testing.T) {
	repo := &mockRepo{}
	_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{})
	require.NoError(t, err)
	assert.Equal(t, "", repo.lastLog.BranchID)
}

func TestLogs_AdminCanPickBranch(t *testing.T) {
	repo := &mockRepo{}
	_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{BranchID: validBranchID})
	require.NoError(t, err)
	assert.Equal(t, validBranchID, repo.lastLog.BranchID)
}

func TestLogs_AdminInvalidBranchUUID_Returns400(t *testing.T) {
	repo := &mockRepo{}
	_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{BranchID: "xyz"})
	requireAppError(t, err, _package.CodeValidation)
	assert.Zero(t, repo.logCalls)
}

func TestLogs_BranchUserForcedToOwnBranch(t *testing.T) {
	repo := &mockRepo{}
	_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "branch", validBranchID, GetAllStockLogsQuery{BranchID: otherBranchID})
	require.NoError(t, err)
	assert.Equal(t, validBranchID, repo.lastLog.BranchID, "ห้ามดูข้ามสาขา")
}

func TestLogs_BranchUserWithoutSessionBranch_Returns403(t *testing.T) {
	repo := &mockRepo{}
	_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "branch", "", GetAllStockLogsQuery{})
	requireAppError(t, err, _package.CodeForbidden)
	assert.Zero(t, repo.logCalls)
}

// ---------- GetStockLogsByBranchId: validate/normalize ----------

func TestLogs_OptionalUUIDFilters(t *testing.T) {
	tests := []struct {
		name  string
		query GetAllStockLogsQuery
		field string
	}{
		{"stockId", GetAllStockLogsQuery{StockID: "bad"}, "stockId"},
		{"orderItemId", GetAllStockLogsQuery{OrderItemID: "bad"}, "orderItemId"},
		{"usageId", GetAllStockLogsQuery{UsageID: "bad"}, "usageId"},
	}
	for _, tt := range tests {
		t.Run(tt.name+" ไม่ใช่ UUID", func(t *testing.T) {
			repo := &mockRepo{}
			_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", tt.query)
			appErr := requireAppError(t, err, _package.CodeValidation)
			assert.Contains(t, appErr.Detail, tt.field)
			assert.Zero(t, repo.logCalls)
		})
	}

	t.Run("UUID ที่ถูกต้องถูกส่งต่อให้ repo", func(t *testing.T) {
		repo := &mockRepo{}
		_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{
			StockID: validStockID, OrderItemID: validOrderItemID, UsageID: validUsageID,
		})
		require.NoError(t, err)
		assert.Equal(t, validStockID, repo.lastLog.StockID)
		assert.Equal(t, validOrderItemID, repo.lastLog.OrderItemID)
		assert.Equal(t, validUsageID, repo.lastLog.UsageID)
	})
}

func TestLogs_LogType(t *testing.T) {
	for lt := range validLogTypes {
		t.Run("ผ่าน: "+lt, func(t *testing.T) {
			repo := &mockRepo{}
			_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{LogType: lt})
			require.NoError(t, err)
			assert.Equal(t, lt, repo.lastLog.LogType)
		})
	}

	t.Run("ตัวพิมพ์ใหญ่และช่องว่างถูก normalize", func(t *testing.T) {
		repo := &mockRepo{}
		_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{LogType: "  COUNT_DIFF "})
		require.NoError(t, err)
		assert.Equal(t, "count_diff", repo.lastLog.LogType)
	})

	t.Run("ไม่รู้จัก", func(t *testing.T) {
		repo := &mockRepo{}
		_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{LogType: "steal"})
		requireAppError(t, err, _package.CodeValidation)
		assert.Zero(t, repo.logCalls)
	})
}

func TestLogs_ValidLogTypesMatchModelCheckConstraint(t *testing.T) {
	// ต้องตรงกับ CHECK constraint ของ stock_logs.log_type ใน models.StockLog
	want := []string{"receive", "sale", "use", "open", "damaged", "expired", "internal", "transfer", "adjust", "count_diff"}
	assert.Len(t, validLogTypes, len(want))
	for _, lt := range want {
		assert.True(t, validLogTypes[lt], lt)
	}
}

func TestLogs_Direction(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"ว่าง", "", "", false},
		{"in", "in", directionIn, false},
		{"out", "out", directionOut, false},
		{"ตัวพิมพ์ใหญ่", " OUT ", directionOut, false},
		{"ไม่รู้จัก", "both", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{Direction: tt.in})
			if tt.wantErr {
				requireAppError(t, err, _package.CodeValidation)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, repo.lastLog.Direction)
		})
	}
}

func TestLogs_DateRange_UsesThaiTime(t *testing.T) {
	repo := &mockRepo{}
	_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{
		From: "2026-10-01", To: "2026-10-03",
	})
	require.NoError(t, err)

	require.NotNil(t, repo.lastLog.From)
	require.NotNil(t, repo.lastLog.Until)
	// from = 00:00 เวลาไทยของวันนั้น = 17:00 UTC ของวันก่อนหน้า
	assert.True(t, repo.lastLog.From.Equal(time.Date(2026, 9, 30, 17, 0, 0, 0, time.UTC)), "from=%v", repo.lastLog.From)
	// until = 00:00 เวลาไทยของวัน "ถัดจาก to" (นับรวมทั้งวัน to)
	assert.True(t, repo.lastLog.Until.Equal(time.Date(2026, 10, 3, 17, 0, 0, 0, time.UTC)), "until=%v", repo.lastLog.Until)
}

func TestLogs_DateRange(t *testing.T) {
	tests := []struct {
		name     string
		from, to string
		wantErr  bool
		detail   string
	}{
		{"ไม่ส่งเลย", "", "", false, ""},
		{"ส่งเฉพาะ from", "2026-10-01", "", false, ""},
		{"ส่งเฉพาะ to", "", "2026-10-01", false, ""},
		{"วันเดียวกัน (ครบทั้งวัน)", "2026-10-01", "2026-10-01", false, ""},
		{"from หลัง to", "2026-10-02", "2026-10-01", true, "from must not be after to"},
		{"from รูปแบบผิด", "01/10/2026", "", true, "from must be YYYY-MM-DD"},
		{"to รูปแบบผิด", "", "tomorrow", true, "to must be YYYY-MM-DD"},
		{"วันที่ไม่มีจริง", "2026-02-30", "", true, "from must be YYYY-MM-DD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{From: tt.from, To: tt.to})
			if tt.wantErr {
				appErr := requireAppError(t, err, _package.CodeValidation)
				assert.Equal(t, tt.detail, appErr.Detail)
				assert.Zero(t, repo.logCalls)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.from != "", repo.lastLog.From != nil)
			assert.Equal(t, tt.to != "", repo.lastLog.Until != nil)
		})
	}
}

func TestLogs_Paging(t *testing.T) {
	t.Run("ค่าเริ่มต้น", func(t *testing.T) {
		repo := &mockRepo{}
		_, meta, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{})
		require.NoError(t, err)
		assert.Equal(t, defaultPage, repo.lastLog.Page)
		assert.Equal(t, defaultPageSize, repo.lastLog.PageSize)
		assert.Equal(t, defaultPage, meta.Page)
	})

	t.Run("pageSize เกินสูงสุด", func(t *testing.T) {
		repo := &mockRepo{}
		_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{PageSize: maxPageSize + 1})
		requireAppError(t, err, _package.CodeValidation)
		assert.Zero(t, repo.logCalls)
	})

	t.Run("page ติดลบ", func(t *testing.T) {
		repo := &mockRepo{}
		_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{Page: -3})
		requireAppError(t, err, _package.CodeValidation)
	})
}

// ---------- GetStockLogsByBranchId: ผลลัพธ์ ----------

func TestLogs_AllFiltersPassedToRepo(t *testing.T) {
	repo := &mockRepo{}
	_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{
		StockID: validStockID, BranchID: validBranchID, LogType: "open", Direction: "out",
		OrderItemID: validOrderItemID, UsageID: validUsageID, Page: 2, PageSize: 10,
	})
	require.NoError(t, err)

	f := repo.lastLog
	assert.Equal(t, validStockID, f.StockID)
	assert.Equal(t, validBranchID, f.BranchID)
	assert.Equal(t, "open", f.LogType)
	assert.Equal(t, directionOut, f.Direction)
	assert.Equal(t, validOrderItemID, f.OrderItemID)
	assert.Equal(t, validUsageID, f.UsageID)
	assert.Equal(t, 2, f.Page)
	assert.Equal(t, 10, f.PageSize)
}

func TestLogs_MapsModelToResponse(t *testing.T) {
	created := time.Date(2026, 10, 21, 4, 36, 8, 0, time.UTC)
	note := "รับของเข้าตั้งต้น (seed)"
	repo := &mockRepo{
		logs: func(ctx context.Context, f stockLogFilter) ([]models.StockLog, int64, error) {
			return []models.StockLog{{
				ID: "log-1", StockID: validStockID, BranchID: validBranchID, EmployeeID: validEmployeeID,
				QtyChange: 10, LogType: "receive", Note: &note, CreatedAt: created,
				Stock:    &models.Stock{ID: validStockID, Name: "ครีมกันแดด"},
				Branch:   &models.Branch{ID: validBranchID, Code: "S1", Name: "สาขา 1"},
				Employee: &models.Employee{ID: validEmployeeID, Name: "สมศรี ใจดี", Role: "manager"},
			}}, 41, nil
		},
	}

	items, meta, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{PageSize: 20})
	require.NoError(t, err)
	require.Len(t, items, 1)

	got := items[0]
	assert.Equal(t, "log-1", got.ID)
	assert.Equal(t, "receive", got.LogType)
	assert.Equal(t, 10, got.QtyChange)
	assert.Equal(t, &note, got.Note)
	assert.Equal(t, StockRef{ID: validStockID, Name: "ครีมกันแดด"}, got.Stock)
	assert.Equal(t, BranchRef{ID: validBranchID, Code: "S1", Name: "สาขา 1"}, got.Branch)
	assert.Equal(t, EmployeeRef{ID: validEmployeeID, Name: "สมศรี ใจดี", Role: "manager"}, got.Employee)

	assert.Equal(t, int64(41), meta.Total)
	assert.Equal(t, 3, meta.TotalPages)
}

func TestLogs_EmptyResultIsEmptySliceNotNil(t *testing.T) {
	items, _, err := NewService(&mockRepo{}).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{})
	require.NoError(t, err)
	assert.NotNil(t, items, "ต้องตอบ [] ไม่ใช่ null")
	assert.Empty(t, items)
}

func TestLogs_RepoErrorPropagates(t *testing.T) {
	boom := errors.New("db down")
	repo := &mockRepo{
		logs: func(ctx context.Context, f stockLogFilter) ([]models.StockLog, int64, error) { return nil, 0, boom },
	}
	_, _, err := NewService(repo).GetStockLogsByBranchId(context.Background(), "admin", "", GetAllStockLogsQuery{})
	assert.ErrorIs(t, err, boom)
}
