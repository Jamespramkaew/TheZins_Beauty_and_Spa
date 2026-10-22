package stocklog

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/models"
	_package "backend/package"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- Mock Service ----------
// handler test ไม่แตะ service จริง ใช้ mock เพื่อดูว่า handler ส่งอะไรเข้า service
// และแปลงผลลัพธ์/error เป็น HTTP status ถูกไหม

type mockService struct {
	quantities func(ctx context.Context, role, branch string, q GetAllStockQuantitiesQuery) ([]StockQuantity, _package.PageMeta, error)
	logs       func(ctx context.Context, role, branch string, q GetAllStockLogsQuery) ([]StockLogResponse, _package.PageMeta, error)
	add        func(ctx context.Context, role, branch string, data AddStockLogMovementRequest) (*models.StockLog, error)

	addCalls             int
	lastAddRequest       AddStockLogMovementRequest
	lastRole, lastBranch string
	lastQuantityQuery    GetAllStockQuantitiesQuery
	lastLogQuery         GetAllStockLogsQuery
}

func (m *mockService) GetAllStockQuantities(ctx context.Context, role, branch string, q GetAllStockQuantitiesQuery) ([]StockQuantity, _package.PageMeta, error) {
	m.lastRole, m.lastBranch, m.lastQuantityQuery = role, branch, q
	if m.quantities != nil {
		return m.quantities(ctx, role, branch, q)
	}
	return []StockQuantity{}, _package.NewPageMeta(1, 20, 0), nil
}

func (m *mockService) AddStockLogMovement(ctx context.Context, role, branch string, data AddStockLogMovementRequest) (*models.StockLog, error) {
	m.addCalls++
	m.lastRole, m.lastBranch, m.lastAddRequest = role, branch, data
	if m.add != nil {
		return m.add(ctx, role, branch, data)
	}
	return &models.StockLog{ID: "new-log", LogType: data.Type, StockID: data.StockID, BranchID: data.BranchID, EmployeeID: data.EmployeeID}, nil
}

func (m *mockService) GetStockLogsByBranchId(ctx context.Context, role, branch string, q GetAllStockLogsQuery) ([]StockLogResponse, _package.PageMeta, error) {
	m.lastRole, m.lastBranch, m.lastLogQuery = role, branch, q
	if m.logs != nil {
		return m.logs(ctx, role, branch, q)
	}
	return []StockLogResponse{}, _package.NewPageMeta(1, 20, 0), nil
}

// ---------- Helpers ----------

// newTestApp ลงทะเบียนเฉพาะ handler (ข้าม AuthMiddleware จริง)
// แล้วตั้ง role/branchId ใน Locals แทน เหมือนที่ middleware จริงทำ
func newTestApp(svc Service, role, branchID string) *fiber.App {
	h := NewHandler(svc)
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("role", role)
		c.Locals("branchId", branchID)
		return c.Next()
	})
	app.Get("/stock-logs/quantities", h.GetAllStockQuantities)
	app.Get("/stock-logs", h.GetStockLogsByBranchId)
	app.Post("/stock-logs", h.AddStockLogMovement)
	return app
}

func doPost(t *testing.T, app *fiber.App, path, body string) (*http.Response, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(raw, &parsed), "response ต้องเป็น JSON: %s", raw)
	return resp, parsed
}

func doGet(t *testing.T, app *fiber.App, path string) (*http.Response, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(raw, &parsed), "response ต้องเป็น JSON: %s", raw)
	return resp, parsed
}

// errorOf ดึง error ออกจากรูปแบบ response ของระบบ (data[0].error)
func errorOf(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	data, ok := body["data"].([]any)
	require.True(t, ok, "data ต้องเป็น array: %v", body)
	require.NotEmpty(t, data)
	return data[0].(map[string]any)["error"].(map[string]any)
}

// ---------- GET /stock-logs ----------

func TestHandler_GetStockLogs_Success(t *testing.T) {
	note := "หมายเหตุ"
	svc := &mockService{
		logs: func(ctx context.Context, role, branch string, q GetAllStockLogsQuery) ([]StockLogResponse, _package.PageMeta, error) {
			return []StockLogResponse{{
				ID: "log-1", LogType: "receive", QtyChange: 5, Note: &note,
				Stock:    StockRef{ID: validStockID, Name: "ครีม"},
				Branch:   BranchRef{ID: validBranchID, Code: "S1", Name: "สาขา 1"},
				Employee: EmployeeRef{ID: validEmployeeID, Name: "สมศรี", Role: "manager"},
			}}, _package.NewPageMeta(1, 20, 1), nil
		},
	}

	resp, body := doGet(t, newTestApp(svc, "admin", ""), "/stock-logs")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, true, body["success"])

	data := body["data"].([]any)
	require.Len(t, data, 1)
	row := data[0].(map[string]any)
	assert.Equal(t, "log-1", row["id"])
	assert.Equal(t, "receive", row["log_type"])
	assert.Equal(t, "สมศรี", row["employee"].(map[string]any)["name"])
	assert.Equal(t, "manager", row["employee"].(map[string]any)["role"])
	assert.Equal(t, "S1", row["branch"].(map[string]any)["code"])

	meta := body["meta"].(map[string]any)
	assert.EqualValues(t, 1, meta["total"])
	assert.EqualValues(t, 1, meta["total_pages"])
}

func TestHandler_GetStockLogs_EmptyDataIsArray(t *testing.T) {
	resp, body := doGet(t, newTestApp(&mockService{}, "admin", ""), "/stock-logs")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	data, ok := body["data"].([]any)
	require.True(t, ok, "data ต้องเป็น [] ไม่ใช่ null: %v", body["data"])
	assert.Empty(t, data)
}

func TestHandler_GetStockLogs_BindsQueryString(t *testing.T) {
	svc := &mockService{}
	path := "/stock-logs?stockId=" + validStockID + "&branchId=" + validBranchID +
		"&logType=open&direction=out&from=2026-10-01&to=2026-10-31" +
		"&orderItemId=" + validOrderItemID + "&usageId=" + validUsageID + "&page=2&pageSize=50"

	resp, _ := doGet(t, newTestApp(svc, "admin", ""), path)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, GetAllStockLogsQuery{
		StockID: validStockID, BranchID: validBranchID, LogType: "open", Direction: "out",
		From: "2026-10-01", To: "2026-10-31", OrderItemID: validOrderItemID, UsageID: validUsageID,
		Page: 2, PageSize: 50,
	}, svc.lastLogQuery)
}

func TestHandler_GetStockLogs_PassesRoleAndBranchFromSession(t *testing.T) {
	svc := &mockService{}
	resp, _ := doGet(t, newTestApp(svc, "branch", validBranchID), "/stock-logs")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, "branch", svc.lastRole)
	assert.Equal(t, validBranchID, svc.lastBranch)
}

func TestHandler_GetStockLogs_InvalidQueryType_Returns400(t *testing.T) {
	svc := &mockService{}
	resp, body := doGet(t, newTestApp(svc, "admin", ""), "/stock-logs?page=abc")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, false, body["success"])
	assert.Equal(t, _package.CodeValidation, errorOf(t, body)["type"])
}

func TestHandler_GetStockLogs_ServiceErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantType   string
	}{
		{"validation → 400", _package.NewAppError("Invalid request", _package.CodeValidation, "bad"), http.StatusBadRequest, _package.CodeValidation},
		{"forbidden → 403", errBranchRequired, http.StatusForbidden, _package.CodeForbidden},
		{"error อื่น → 500", errors.New("pq: connection refused"), http.StatusInternalServerError, _package.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{
				logs: func(ctx context.Context, role, branch string, q GetAllStockLogsQuery) ([]StockLogResponse, _package.PageMeta, error) {
					return nil, _package.PageMeta{}, tt.err
				},
			}
			resp, body := doGet(t, newTestApp(svc, "admin", ""), "/stock-logs")

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Equal(t, false, body["success"])
			errObj := errorOf(t, body)
			assert.Equal(t, tt.wantType, errObj["type"])
			if tt.wantStatus == http.StatusInternalServerError {
				// ห้ามหลุดรายละเอียด error ภายในออกไปให้ client
				assert.NotContains(t, body["message"], "pq:")
				assert.Empty(t, errObj["details"])
			}
		})
	}
}

// ---------- GET /stock-logs/quantities ----------

func TestHandler_GetQuantities_Success(t *testing.T) {
	svc := &mockService{
		quantities: func(ctx context.Context, role, branch string, q GetAllStockQuantitiesQuery) ([]StockQuantity, _package.PageMeta, error) {
			return []StockQuantity{{
				StockID: validStockID, StockName: "ครีมกันแดด", Category: "normal", UsesPerUnit: 1, MinQty: 5,
				IsActive: true, BranchID: validBranchID, BranchName: "สาขา 1", Quantity: 12, IsLowStock: false,
			}}, _package.NewPageMeta(1, 20, 1), nil
		},
	}

	resp, body := doGet(t, newTestApp(svc, "admin", ""), "/stock-logs/quantities")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, true, body["success"])

	row := body["data"].([]any)[0].(map[string]any)
	assert.Equal(t, "ครีมกันแดด", row["stock_name"])
	assert.EqualValues(t, 12, row["quantity"])
	assert.Equal(t, false, row["is_low_stock"])
	assert.Equal(t, "สาขา 1", row["branch_name"])
	assert.EqualValues(t, 1, body["meta"].(map[string]any)["total"])
}

func TestHandler_GetQuantities_EmptyDataIsArray(t *testing.T) {
	_, body := doGet(t, newTestApp(&mockService{}, "admin", ""), "/stock-logs/quantities")

	data, ok := body["data"].([]any)
	require.True(t, ok, "data ต้องเป็น [] ไม่ใช่ null: %v", body["data"])
	assert.Empty(t, data)
}

func TestHandler_GetQuantities_BindsQueryString(t *testing.T) {
	svc := &mockService{}
	path := "/stock-logs/quantities?branchId=" + validBranchID +
		"&search=" + "%E0%B8%84%E0%B8%A3%E0%B8%B5%E0%B8%A1" + // "ครีม" (URL-encoded)
		"&category=premium&lowStock=true&active=all&page=3&pageSize=10"

	resp, _ := doGet(t, newTestApp(svc, "admin", ""), path)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, GetAllStockQuantitiesQuery{
		BranchID: validBranchID, Search: "ครีม", Category: "premium", LowStock: true, Active: "all", Page: 3, PageSize: 10,
	}, svc.lastQuantityQuery)
}

func TestHandler_GetQuantities_PassesRoleAndBranchFromSession(t *testing.T) {
	svc := &mockService{}
	resp, _ := doGet(t, newTestApp(svc, "branch", validBranchID), "/stock-logs/quantities")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, "branch", svc.lastRole)
	assert.Equal(t, validBranchID, svc.lastBranch)
}

func TestHandler_GetQuantities_InvalidQueryType_Returns400(t *testing.T) {
	resp, body := doGet(t, newTestApp(&mockService{}, "admin", ""), "/stock-logs/quantities?lowStock=maybe")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, _package.CodeValidation, errorOf(t, body)["type"])
}

func TestHandler_GetQuantities_ServiceErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"validation → 400", _package.NewAppError("Invalid request", _package.CodeValidation, "bad"), http.StatusBadRequest},
		{"forbidden → 403", errBranchRequired, http.StatusForbidden},
		{"error อื่น → 500", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{
				quantities: func(ctx context.Context, role, branch string, q GetAllStockQuantitiesQuery) ([]StockQuantity, _package.PageMeta, error) {
					return nil, _package.PageMeta{}, tt.err
				},
			}
			resp, body := doGet(t, newTestApp(svc, "admin", ""), "/stock-logs/quantities")

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Equal(t, false, body["success"])
		})
	}
}

// ---------- POST /stock-logs ----------

func validMovementBody() string {
	return `{"type":"receive","branchId":"` + validBranchID + `","stockId":"` + validStockID +
		`","employeeId":"` + validEmployeeID + `","qty":10,"note":"รับของ"}`
}

func TestHandler_AddMovement_Success201(t *testing.T) {
	note := "รับของ"
	svc := &mockService{
		add: func(ctx context.Context, role, branch string, data AddStockLogMovementRequest) (*models.StockLog, error) {
			return &models.StockLog{
				ID: "log-1", LogType: "receive", QtyChange: 10, Note: &note,
				StockID: validStockID, BranchID: validBranchID, EmployeeID: validEmployeeID,
			}, nil
		},
	}

	resp, body := doPost(t, newTestApp(svc, "admin", ""), "/stock-logs", validMovementBody())

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, true, body["success"])
	data := body["data"].(map[string]any)
	assert.Equal(t, "log-1", data["id"])
	assert.Equal(t, "receive", data["log_type"])
	assert.EqualValues(t, 10, data["qty_change"])
	assert.Equal(t, "รับของ", data["note"])
}

func TestHandler_AddMovement_BindsJSONBody(t *testing.T) {
	svc := &mockService{}
	resp, _ := doPost(t, newTestApp(svc, "admin", ""), "/stock-logs", validMovementBody())
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	assert.Equal(t, AddStockLogMovementRequest{
		Type: "receive", BranchID: validBranchID, StockID: validStockID, EmployeeID: validEmployeeID, Qty: 10, Note: "รับของ",
	}, svc.lastAddRequest)
}

func TestHandler_AddMovement_PassesRoleAndBranchFromSession(t *testing.T) {
	svc := &mockService{}
	resp, _ := doPost(t, newTestApp(svc, "branch", validBranchID), "/stock-logs", validMovementBody())
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	assert.Equal(t, "branch", svc.lastRole)
	assert.Equal(t, validBranchID, svc.lastBranch)
}

func TestHandler_AddMovement_BadBodyReturns400(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"JSON พัง", `{"type":`},
		{"qty เป็น string", `{"type":"receive","qty":"ten"}`},
		{"qty เป็นทศนิยม", `{"type":"receive","qty":1.5}`},
		{"type เป็นตัวเลข", `{"type":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{}
			resp, body := doPost(t, newTestApp(svc, "admin", ""), "/stock-logs", tt.body)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			assert.Equal(t, false, body["success"])
			assert.Equal(t, _package.CodeValidation, errorOf(t, body)["type"])
			assert.Zero(t, svc.addCalls, "service ต้องไม่ถูกเรียก")
		})
	}
}

func TestHandler_AddMovement_WrongContentTypeReturns400(t *testing.T) {
	svc := &mockService{}
	req := httptest.NewRequest(http.MethodPost, "/stock-logs", strings.NewReader(validMovementBody()))
	req.Header.Set("Content-Type", "text/plain")
	resp, err := newTestApp(svc, "admin", "").Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Zero(t, svc.addCalls)
}

func TestHandler_AddMovement_ServiceErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantType   string
	}{
		{"validation → 400", _package.NewAppError("Invalid request", _package.CodeValidation, "qty must be positive"), http.StatusBadRequest, _package.CodeValidation},
		{"forbidden → 403", errBranchRequired, http.StatusForbidden, _package.CodeForbidden},
		{"error อื่น → 500", errors.New("pq: connection refused"), http.StatusInternalServerError, _package.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{
				add: func(ctx context.Context, role, branch string, data AddStockLogMovementRequest) (*models.StockLog, error) {
					return nil, tt.err
				},
			}
			resp, body := doPost(t, newTestApp(svc, "admin", ""), "/stock-logs", validMovementBody())

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

// ---------- ความปลอดภัยของ route ----------

func TestRegisterRoutes_PostRequiresAuth(t *testing.T) {
	svc := &mockService{}
	app := fiber.New(fiber.Config{ErrorHandler: _package.FiberErrorHandler})
	NewHandler(svc).RegisterRoutes(app.Group("/api"))

	req := httptest.NewRequest(http.MethodPost, "/api/stock-logs", strings.NewReader(validMovementBody()))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Zero(t, svc.addCalls, "ไม่มี token ต้องไม่ถึง service")
}

// เส้นทางจริงต้องผูก AuthMiddleware ทั้งสองเส้น: ไม่มี token ต้องเข้าไม่ได้ และ service ต้องไม่ถูกเรียก
func TestRegisterRoutes_RequiresAuth(t *testing.T) {
	svc := &mockService{}
	app := fiber.New(fiber.Config{ErrorHandler: _package.FiberErrorHandler})
	NewHandler(svc).RegisterRoutes(app.Group("/api"))

	for _, path := range []string{"/api/stock-logs", "/api/stock-logs/quantities"} {
		t.Run(path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			assert.Empty(t, svc.lastRole, "service ต้องไม่ถูกเรียก")
		})
	}
}

// stock_logs เป็น ledger แบบ append-only: ต้องไม่มี route แก้/ลบ (เพิ่มได้ผ่าน POST เท่านั้น)
func TestRegisterRoutes_NoUpdateOrDelete(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: _package.FiberErrorHandler})
	NewHandler(&mockService{}).RegisterRoutes(app.Group("/api"))

	for _, method := range []string{http.MethodPatch, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(method, "/api/stock-logs", nil))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Contains(t, []int{http.StatusNotFound, http.StatusMethodNotAllowed}, resp.StatusCode)
		})
	}
}
