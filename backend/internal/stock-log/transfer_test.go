package stocklog

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"backend/internal/models"
	_package "backend/package"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- helpers ----------

// ต้นทาง = validBranchID (ตรงกับสาขาของพนักงานใน mock) ปลายทาง = otherBranchID
func validTransfer() TransferStockLogRequest {
	return TransferStockLogRequest{
		StockID:      validStockID,
		FromBranchId: validBranchID,
		ToBranchId:   otherBranchID,
		EmployeeID:   validEmployeeID,
		Qty:          5,
	}
}

func newTransferService(repo *mockRepo) Service {
	return NewService(repo, &mockStockRepo{}, &mockBranchRepo{}, &mockEmployeeRepo{})
}

func intPtr(n int) *int { return &n }

// ---------- Service: TransferStockLog ----------

func TestTransfer_Success_AdminSendsNormalizedRequestToRepo(t *testing.T) {
	repo := &mockRepo{}
	req := validTransfer()
	req.StockID = "  " + validStockID + " "
	req.FromBranchId = " " + validBranchID + " "
	req.ToBranchId = " " + otherBranchID + "  "
	req.EmployeeID = " " + validEmployeeID + " "

	err := newTransferService(repo).TransferStockLog(context.Background(), "admin", "", req)
	require.NoError(t, err)

	require.Equal(t, 1, repo.transferCalls)
	assert.Equal(t, "transfer", repo.lastTransferType)
	got := repo.lastTransfer
	assert.Equal(t, validStockID, got.StockID, "ต้อง trim")
	assert.Equal(t, validBranchID, got.FromBranchId)
	assert.Equal(t, otherBranchID, got.ToBranchId)
	assert.Equal(t, validEmployeeID, got.EmployeeID)
	assert.Equal(t, 5, got.Qty, "qty ต้องเป็นบวก repo เป็นคนใส่เครื่องหมายให้แต่ละสาขา")
}

func TestTransfer_NoteIsGeneratedByServer(t *testing.T) {
	repo := &mockRepo{}
	branches := &mockBranchRepo{
		find: func(ctx context.Context, id string) (*models.Branch, error) {
			names := map[string]string{validBranchID: "สาขา 1", otherBranchID: "สาขา 2"}
			return &models.Branch{ID: id, Name: names[id], IsActive: true}, nil
		},
	}
	req := validTransfer()
	req.Note = "client พยายามเขียน note เอง"

	err := NewService(repo, &mockStockRepo{}, branches, &mockEmployeeRepo{}).TransferStockLog(context.Background(), "admin", "", req)
	require.NoError(t, err)
	assert.Equal(t, "Transfer from สาขา 1 to สาขา 2", repo.lastTransfer.Note)
}

func TestTransfer_FromBranch(t *testing.T) {
	t.Run("admin ต้องส่ง fromBranchId", func(t *testing.T) {
		repo := &mockRepo{}
		req := validTransfer()
		req.FromBranchId = ""

		err := newTransferService(repo).TransferStockLog(context.Background(), "admin", "", req)
		appErr := requireAppError(t, err, _package.CodeValidation)
		assert.Contains(t, appErr.Detail, "fromBranchId")
		assert.Zero(t, repo.transferCalls)
	})

	t.Run("admin fromBranchId ไม่ใช่ UUID", func(t *testing.T) {
		repo := &mockRepo{}
		req := validTransfer()
		req.FromBranchId = "xyz"

		err := newTransferService(repo).TransferStockLog(context.Background(), "admin", "", req)
		requireAppError(t, err, _package.CodeValidation)
		assert.Zero(t, repo.transferCalls)
	})

	t.Run("บัญชีสาขาถูกบังคับเป็นสาขาใน session (ไม่สนค่าที่ส่งมา)", func(t *testing.T) {
		repo := &mockRepo{}
		req := validTransfer()
		req.FromBranchId = "garbage" // ไม่ถูกใช้ จึงไม่ต้อง 400

		err := newTransferService(repo).TransferStockLog(context.Background(), "branch", validBranchID, req)
		require.NoError(t, err)
		assert.Equal(t, validBranchID, repo.lastTransfer.FromBranchId)
	})

	t.Run("บัญชีสาขาไม่ต้องส่ง fromBranchId", func(t *testing.T) {
		repo := &mockRepo{}
		req := validTransfer()
		req.FromBranchId = ""

		err := newTransferService(repo).TransferStockLog(context.Background(), "branch", validBranchID, req)
		require.NoError(t, err)
		assert.Equal(t, validBranchID, repo.lastTransfer.FromBranchId)
	})

	t.Run("บัญชีสาขาที่ session ไม่มีสาขา → 403", func(t *testing.T) {
		repo := &mockRepo{}
		err := newTransferService(repo).TransferStockLog(context.Background(), "branch", "", validTransfer())
		requireAppError(t, err, _package.CodeForbidden)
		assert.Zero(t, repo.transferCalls)
	})

	t.Run("บัญชีสาขาส่งสาขาอื่นเป็นต้นทาง → ถูกเขียนทับ ไม่โอนจากสาขาอื่น", func(t *testing.T) {
		repo := &mockRepo{}
		req := validTransfer()
		req.FromBranchId = otherBranchID
		req.ToBranchId = "f6a7b8c9-1111-4222-8333-944455556666"

		err := newTransferService(repo).TransferStockLog(context.Background(), "branch", validBranchID, req)
		require.NoError(t, err)
		assert.Equal(t, validBranchID, repo.lastTransfer.FromBranchId)
	})
}

func TestTransfer_ValidationReturns400AndNeverReachesRepo(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(r *TransferStockLogRequest)
		field  string
	}{
		{"qty เป็น 0", func(r *TransferStockLogRequest) { r.Qty = 0 }, "qty"},
		{"qty ติดลบ", func(r *TransferStockLogRequest) { r.Qty = -3 }, "qty"},
		{"stockId ว่าง", func(r *TransferStockLogRequest) { r.StockID = "" }, "stockId"},
		{"stockId ไม่ใช่ UUID", func(r *TransferStockLogRequest) { r.StockID = "abc" }, "stockId"},
		{"toBranchId ว่าง", func(r *TransferStockLogRequest) { r.ToBranchId = "  " }, "toBranchId"},
		{"toBranchId ไม่ใช่ UUID", func(r *TransferStockLogRequest) { r.ToBranchId = "abc" }, "toBranchId"},
		{"ต้นทางกับปลายทางเป็นสาขาเดียวกัน", func(r *TransferStockLogRequest) { r.ToBranchId = r.FromBranchId }, "different"},
		{"employeeId ว่าง", func(r *TransferStockLogRequest) { r.EmployeeID = "" }, "employeeId"},
		{"employeeId ไม่ใช่ UUID", func(r *TransferStockLogRequest) { r.EmployeeID = "abc" }, "employeeId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			stocks, branches, employees := &mockStockRepo{}, &mockBranchRepo{}, &mockEmployeeRepo{}
			req := validTransfer()
			tt.mutate(&req)

			err := NewService(repo, stocks, branches, employees).TransferStockLog(context.Background(), "admin", "", req)
			appErr := requireAppError(t, err, _package.CodeValidation)
			assert.Contains(t, appErr.Detail, tt.field)
			assert.Zero(t, repo.transferCalls)
			assert.Zero(t, stocks.calls+branches.calls+employees.calls+repo.stockQtyCalls, "format ผิดต้องไม่ถึง DB")
		})
	}
}

func TestTransfer_ExistenceChecks(t *testing.T) {
	t.Run("ไม่พบสินค้า", func(t *testing.T) {
		repo := &mockRepo{}
		stocks := &mockStockRepo{find: func(ctx context.Context, id string) (*models.Stock, error) { return nil, nil }}

		err := NewService(repo, stocks, &mockBranchRepo{}, &mockEmployeeRepo{}).TransferStockLog(context.Background(), "admin", "", validTransfer())
		appErr := requireAppError(t, err, _package.CodeValidation)
		assert.Contains(t, appErr.Detail, "stockId")
		assert.Zero(t, repo.transferCalls)
	})

	notFoundFor := func(target string) *mockBranchRepo {
		return &mockBranchRepo{find: func(ctx context.Context, id string) (*models.Branch, error) {
			if id == target {
				return nil, nil
			}
			return &models.Branch{ID: id, IsActive: true}, nil
		}}
	}
	inactiveFor := func(target string) *mockBranchRepo {
		return &mockBranchRepo{find: func(ctx context.Context, id string) (*models.Branch, error) {
			return &models.Branch{ID: id, IsActive: id != target}, nil
		}}
	}

	tests := []struct {
		name     string
		branches *mockBranchRepo
		want     string
	}{
		{"ไม่พบสาขาต้นทาง", notFoundFor(validBranchID), "origin branch not found"},
		{"ไม่พบสาขาปลายทาง", notFoundFor(otherBranchID), "destination branch not found"},
		{"สาขาต้นทางถูกปิด", inactiveFor(validBranchID), "origin branch is not active"},
		{"สาขาปลายทางถูกปิด", inactiveFor(otherBranchID), "destination branch is not active"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			err := NewService(repo, &mockStockRepo{}, tt.branches, &mockEmployeeRepo{}).TransferStockLog(context.Background(), "admin", "", validTransfer())
			appErr := requireAppError(t, err, _package.CodeValidation)
			assert.Equal(t, tt.want, appErr.Detail)
			assert.Zero(t, repo.transferCalls)
		})
	}
}

func TestTransfer_EmployeeChecks(t *testing.T) {
	tests := []struct {
		name string
		emp  *models.Employee
		want string
	}{
		{"ไม่พบพนักงาน", nil, "employeeId not found"},
		{"พนักงานไม่ active", &models.Employee{ID: validEmployeeID, IsActive: false, BranchID: validBranchID}, "employee is not active"},
		{"พนักงานอยู่สาขาปลายทาง ไม่ใช่ต้นทาง", &models.Employee{ID: validEmployeeID, IsActive: true, BranchID: otherBranchID}, "employee does not belong to the origin branch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			employees := &mockEmployeeRepo{find: func(ctx context.Context, id string) (*models.Employee, error) { return tt.emp, nil }}

			err := NewService(repo, &mockStockRepo{}, &mockBranchRepo{}, employees).TransferStockLog(context.Background(), "admin", "", validTransfer())
			appErr := requireAppError(t, err, _package.CodeValidation)
			assert.Equal(t, tt.want, appErr.Detail)
			assert.Zero(t, repo.transferCalls)
		})
	}
}

func TestTransfer_StockAvailability(t *testing.T) {
	tests := []struct {
		name      string
		available int
		qty       int
		wantErr   bool
	}{
		{"ยอดพอดี", 5, 5, false},
		{"ยอดเหลือเฟือ", 100, 5, false},
		{"ยอดขาด 1", 4, 5, true},
		{"ไม่มีของเลย", 0, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{stockQtyIs: intPtr(tt.available)}
			req := validTransfer()
			req.Qty = tt.qty

			err := newTransferService(repo).TransferStockLog(context.Background(), "admin", "", req)

			assert.Equal(t, validStockID, repo.lastStockQtyStock)
			assert.Equal(t, validBranchID, repo.lastStockQtyBr, "ต้องเช็คยอดที่สาขาต้นทาง ไม่ใช่ปลายทาง")
			if tt.wantErr {
				appErr := requireAppError(t, err, _package.CodeValidation)
				assert.Contains(t, appErr.Detail, "insufficient stock")
				assert.Zero(t, repo.transferCalls, "ยอดไม่พอต้องไม่เขียน")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, repo.transferCalls)
		})
	}
}

func TestTransfer_RepoErrorsPropagate(t *testing.T) {
	boom := errors.New("db down")

	t.Run("GetStockQuantity พัง", func(t *testing.T) {
		repo := &mockRepo{stockQty: func(ctx context.Context, s, b string) (int, error) { return 0, boom }}
		err := newTransferService(repo).TransferStockLog(context.Background(), "admin", "", validTransfer())
		assert.ErrorIs(t, err, boom)
		assert.Zero(t, repo.transferCalls)
	})

	t.Run("TransferStockLog พัง", func(t *testing.T) {
		repo := &mockRepo{transfer: func(ctx context.Context, d TransferStockLogRequest, lt string) error { return boom }}
		err := newTransferService(repo).TransferStockLog(context.Background(), "admin", "", validTransfer())
		assert.ErrorIs(t, err, boom)
	})
}

// ---------- Service: AddStockLogMovement เช็คยอดคงเหลือ ----------

func TestAdd_StockAvailability(t *testing.T) {
	tests := []struct {
		name      string
		logType   string
		qty       int
		available int
		wantErr   bool
		wantCheck bool // ควรเรียก GetStockQuantity ไหม
	}{
		{"expired ยอดพอดี", "expired", 10, 10, false, true},
		{"expired ยอดไม่พอ", "expired", 10, 9, true, true},
		{"damaged ยอดไม่พอ", "damaged", 1, 0, true, true},
		{"internal ยอดไม่พอ", "internal", 3, 2, true, true},
		{"adjust ลบ ยอดไม่พอ", "adjust", -4, 3, true, true},
		{"adjust ลบ ยอดพอดี", "adjust", -3, 3, false, true},
		{"receive ไม่เช็คยอด (ยอดเป็น 0 ก็รับได้)", "receive", 10, 0, false, false},
		{"adjust บวก ไม่เช็คยอด", "adjust", 4, 0, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{stockQtyIs: intPtr(tt.available)}
			req := validMovement()
			req.Type = tt.logType
			req.Qty = tt.qty
			req.Note = "แก้ข้อมูล"

			got, err := newTransferService(repo).AddStockLogMovement(context.Background(), "admin", "", req)

			assert.Equal(t, tt.wantCheck, repo.stockQtyCalls > 0)
			if tt.wantErr {
				appErr := requireAppError(t, err, _package.CodeValidation)
				assert.Contains(t, appErr.Detail, "insufficient stock")
				assert.Nil(t, got)
				assert.Zero(t, repo.createCalls, "ยอดไม่พอต้องไม่เขียน")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, repo.createCalls)
		})
	}
}

func TestAdd_StockAvailability_ChecksSelectedBranch(t *testing.T) {
	repo := &mockRepo{}
	req := validMovement()
	req.Type = "expired"
	req.BranchID = otherBranchID // บัญชีสาขาถูกบังคับเป็นสาขาใน session

	employees := &mockEmployeeRepo{}
	_, err := NewService(repo, &mockStockRepo{}, &mockBranchRepo{}, employees).AddStockLogMovement(context.Background(), "branch", validBranchID, req)
	require.NoError(t, err)
	assert.Equal(t, validBranchID, repo.lastStockQtyBr)
	assert.Equal(t, validStockID, repo.lastStockQtyStock)
}

func TestAdd_StockQuantityError_Propagates(t *testing.T) {
	boom := errors.New("db down")
	repo := &mockRepo{stockQty: func(ctx context.Context, s, b string) (int, error) { return 0, boom }}
	req := validMovement()
	req.Type = "expired"

	_, err := newTransferService(repo).AddStockLogMovement(context.Background(), "admin", "", req)
	assert.ErrorIs(t, err, boom)
	assert.Zero(t, repo.createCalls)
}

// ---------- Handler: POST /stock-logs/transfer ----------

func validTransferBody() string {
	return `{"stock_id":"` + validStockID + `","from_branch_id":"` + validBranchID +
		`","to_branch_id":"` + otherBranchID + `","employee_id":"` + validEmployeeID + `","qty":5}`
}

func TestHandler_Transfer_Success201(t *testing.T) {
	svc := &mockService{}
	resp, body := doPost(t, newTestApp(svc, "admin", ""), "/stock-logs/transfer", validTransferBody())

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, 1, svc.transferCalls)
}

func TestHandler_Transfer_BindsJSONBody(t *testing.T) {
	svc := &mockService{}
	resp, _ := doPost(t, newTestApp(svc, "admin", ""), "/stock-logs/transfer", validTransferBody())
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	assert.Equal(t, TransferStockLogRequest{
		StockID: validStockID, FromBranchId: validBranchID, ToBranchId: otherBranchID, EmployeeID: validEmployeeID, Qty: 5,
	}, svc.lastTransferRequest)
}

func TestHandler_Transfer_ClientCannotSetNote(t *testing.T) {
	svc := &mockService{}
	body := `{"stock_id":"` + validStockID + `","to_branch_id":"` + otherBranchID + `","employee_id":"` + validEmployeeID +
		`","qty":5,"note":"แอบเขียนเอง"}`
	resp, _ := doPost(t, newTestApp(svc, "admin", ""), "/stock-logs/transfer", body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	assert.Empty(t, svc.lastTransferRequest.Note, "note ต้องถูกสร้างโดย service เท่านั้น")
}

func TestHandler_Transfer_PassesRoleAndBranchFromSession(t *testing.T) {
	svc := &mockService{}
	resp, _ := doPost(t, newTestApp(svc, "branch", validBranchID), "/stock-logs/transfer", validTransferBody())
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	assert.Equal(t, "branch", svc.lastRole)
	assert.Equal(t, validBranchID, svc.lastBranch)
}

func TestHandler_Transfer_BadBodyReturns400(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"JSON พัง", `{"stock_id":`},
		{"qty เป็น string", `{"qty":"five"}`},
		{"qty เป็นทศนิยม", `{"qty":1.5}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{}
			resp, body := doPost(t, newTestApp(svc, "admin", ""), "/stock-logs/transfer", tt.body)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			assert.Equal(t, _package.CodeValidation, errorOf(t, body)["type"])
			assert.Zero(t, svc.transferCalls, "service ต้องไม่ถูกเรียก")
		})
	}
}

func TestHandler_Transfer_ServiceErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantType   string
	}{
		{"validation → 400", _package.NewAppError("Invalid request", _package.CodeValidation, "insufficient stock"), http.StatusBadRequest, _package.CodeValidation},
		{"forbidden → 403", errBranchRequired, http.StatusForbidden, _package.CodeForbidden},
		{"error อื่น → 500", errors.New("pq: connection refused"), http.StatusInternalServerError, _package.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{
				transfer: func(ctx context.Context, role, branch string, data TransferStockLogRequest) error { return tt.err },
			}
			resp, body := doPost(t, newTestApp(svc, "admin", ""), "/stock-logs/transfer", validTransferBody())

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Equal(t, false, body["success"])
			errObj := errorOf(t, body)
			assert.Equal(t, tt.wantType, errObj["type"])
			if tt.wantStatus == http.StatusInternalServerError {
				assert.NotContains(t, body["message"], "pq:")
				assert.Empty(t, errObj["details"])
			}
		})
	}
}
