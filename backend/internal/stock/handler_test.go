package stock

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/currentuser"
	"backend/internal/middleware"
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
	addStock    func(ctx context.Context, data AddStockRequest) (models.Stock, error)
	updateStock func(ctx context.Context, id string, data UpdateStockRequest) (models.Stock, error)
	deleteStock func(ctx context.Context, id string) error
	restore     func(ctx context.Context, id string) (models.Stock, error)
	containers  func(ctx context.Context, user *currentuser.CurrentUser, q ContainerQuery) ([]models.StockContainer, _package.PageMeta, error)

	addCalls    int
	updateCalls int
	lastUpdate  UpdateStockRequest
	lastQuery   ContainerQuery
	lastUser    *currentuser.CurrentUser
}

func (m *mockService) GetAllStock(ctx context.Context) ([]models.Stock, error) {
	return []models.Stock{{ID: validStockID, Name: "ครีมกันแดด"}}, nil
}

func (m *mockService) GetStockContainersByStockID(ctx context.Context, user *currentuser.CurrentUser, q ContainerQuery) ([]models.StockContainer, _package.PageMeta, error) {
	m.lastUser, m.lastQuery = user, q
	if m.containers != nil {
		return m.containers(ctx, user, q)
	}
	return []models.StockContainer{}, _package.NewPageMeta(1, 20, 0), nil
}

func (m *mockService) AddStock(ctx context.Context, data AddStockRequest) (models.Stock, error) {
	m.addCalls++
	if m.addStock != nil {
		return m.addStock(ctx, data)
	}
	return models.Stock{ID: validStockID, Name: data.StockName}, nil
}

func (m *mockService) UpdateStockById(ctx context.Context, id string, data UpdateStockRequest) (models.Stock, error) {
	m.updateCalls++
	m.lastUpdate = data
	if m.updateStock != nil {
		return m.updateStock(ctx, id, data)
	}
	return models.Stock{ID: id}, nil
}

func (m *mockService) DeleteStockById(ctx context.Context, id string) error {
	if m.deleteStock != nil {
		return m.deleteStock(ctx, id)
	}
	return nil
}

func (m *mockService) RestoreStockById(ctx context.Context, id string) (models.Stock, error) {
	if m.restore != nil {
		return m.restore(ctx, id)
	}
	return models.Stock{ID: id}, nil
}

// ---------- Helpers ----------

// newTestApp สร้าง fiber app เล็กๆ ที่ลงทะเบียนเฉพาะ handler (ข้าม AuthMiddleware จริง)
// แล้วตั้ง CurrentUser ใน context แทน เหมือนที่ middleware จริงทำ
func newTestApp(svc Service, role, branchID string) *fiber.App {
	h := NewHandler(svc)
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		user := &currentuser.CurrentUser{
			ID:       "test-user-id",
			Username: "testuser",
			Role:     role,
		}
		if branchID != "" {
			user.BranchID = &branchID
		}
		currentuser.Set(c, user)
		return c.Next()
	})
	app.Get("/stocks", h.GetAllStock)
	app.Get("/stocks/:id/containers", h.GetStockContainersByStockID)
	app.Post("/stocks", middleware.RequireAdminMiddleware, h.AddStock)
	app.Patch("/stocks/:id", middleware.RequireAdminMiddleware, h.UpdateStockById)
	app.Delete("/stocks/:id", middleware.RequireAdminMiddleware, h.DeleteStockById)
	app.Post("/stocks/:id/restore", middleware.RequireAdminMiddleware, h.RestoreStockById)
	return app
}

func doRequest(t *testing.T, app *fiber.App, method, path, body string) (*http.Response, map[string]any) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
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

// errorDetail ดึง error.details ออกจากรูปแบบ response ของระบบ (data[0].error.details)
func errorDetail(t *testing.T, body map[string]any) string {
	t.Helper()
	data, ok := body["data"].([]any)
	require.True(t, ok, "data ต้องเป็น array: %v", body)
	require.NotEmpty(t, data)
	errObj := data[0].(map[string]any)["error"].(map[string]any)
	detail, _ := errObj["details"].(string)
	return detail
}

// ---------- สิทธิ์: เฉพาะ admin ----------

func TestHandlers_NonAdminIsForbidden(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"POST /stocks", http.MethodPost, "/stocks", `{"stock_name":"x"}`},
		{"PATCH /stocks/:id", http.MethodPatch, "/stocks/" + validStockID, `{"min_qty":1}`},
		{"DELETE /stocks/:id", http.MethodDelete, "/stocks/" + validStockID, ""},
		{"POST /stocks/:id/restore", http.MethodPost, "/stocks/" + validStockID + "/restore", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{}
			app := newTestApp(svc, "branch", validBranchID)

			resp, _ := doRequest(t, app, tt.method, tt.path, tt.body)

			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
			assert.Equal(t, 0, svc.addCalls+svc.updateCalls, "ห้ามเรียก service")
		})
	}
}

// ---------- GET ----------

func TestGetAllStock_OK(t *testing.T) {
	app := newTestApp(&mockService{}, "branch", validBranchID)

	resp, body := doRequest(t, app, http.MethodGet, "/stocks", "")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, true, body["success"])
}

func TestGetContainers_PassesQueryAndSessionToService(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "branch", validBranchID)

	resp, body := doRequest(t, app, http.MethodGet,
		"/stocks/"+validStockID+"/containers?status=closed&page=2&pageSize=10&branchId="+validBranchID, "")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, body["meta"], "endpoint แบ่งหน้าต้องมี meta")
	assert.Equal(t, "branch", svc.lastUser.Role)
	assert.NotNil(t, svc.lastUser.BranchID)
	assert.Equal(t, validBranchID, *svc.lastUser.BranchID)
	assert.Equal(t, ContainerQuery{
		StockID:  validStockID,
		BranchID: validBranchID,
		Status:   "closed",
		Page:     "2",
		PageSize: "10",
	}, svc.lastQuery)
}

func TestGetContainers_ServiceValidationError_Returns400(t *testing.T) {
	svc := &mockService{
		containers: func(ctx context.Context, user *currentuser.CurrentUser, q ContainerQuery) ([]models.StockContainer, _package.PageMeta, error) {
			return nil, _package.PageMeta{}, invalid("status must be one of: open, closed, all")
		},
	}
	app := newTestApp(svc, "admin", "")

	resp, body := doRequest(t, app, http.MethodGet, "/stocks/"+validStockID+"/containers?status=x", "")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, errorDetail(t, body), "status")
}

// ---------- POST ----------

func TestAddStock_Handler_Created(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin", "")

	resp, body := doRequest(t, app, http.MethodPost, "/stocks",
		`{"stock_name":"ครีมกันแดด","price":500,"is_sellable":true,"category":"premium","uses_per_unit":10,"min_qty":5}`)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, 1, svc.addCalls)
}

// ชนิดข้อมูลผิด ต้อง 400 ไม่ใช่ 500
func TestAddStock_Handler_WrongType_Returns400(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin", "")

	resp, body := doRequest(t, app, http.MethodPost, "/stocks",
		`{"stock_name":"x","price":10,"category":"normal","uses_per_unit":"3"}`)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, errorDetail(t, body), "uses_per_unit")
	assert.Equal(t, 0, svc.addCalls)
}

func TestAddStock_Handler_ServiceConflict_Returns409(t *testing.T) {
	svc := &mockService{
		addStock: func(ctx context.Context, data AddStockRequest) (models.Stock, error) {
			return models.Stock{}, duplicateName(data.StockName)
		},
	}
	app := newTestApp(svc, "admin", "")

	resp, _ := doRequest(t, app, http.MethodPost, "/stocks",
		`{"stock_name":"ซ้ำ","price":10,"category":"normal","uses_per_unit":1}`)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestAddStock_Handler_UnknownServiceError_Returns500(t *testing.T) {
	svc := &mockService{
		addStock: func(ctx context.Context, data AddStockRequest) (models.Stock, error) {
			return models.Stock{}, assert.AnError
		},
	}
	app := newTestApp(svc, "admin", "")

	resp, _ := doRequest(t, app, http.MethodPost, "/stocks",
		`{"stock_name":"x","price":10,"category":"normal","uses_per_unit":1}`)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ---------- PATCH ----------

func TestUpdateStock_Handler_ParsesFields(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin", "")

	resp, _ := doRequest(t, app, http.MethodPatch, "/stocks/"+validStockID,
		`{"stock_name":"ใหม่","price":"750.50","is_sellable":true,"category":"normal","uses_per_unit":3,"min_qty":2,"is_active":false}`)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	got := svc.lastUpdate
	require.NotNil(t, got.StockName)
	assert.Equal(t, "ใหม่", *got.StockName)
	assert.True(t, got.PriceSet)
	assert.Equal(t, "750.5", got.Price.String())
	assert.Equal(t, true, *got.IsSellable)
	assert.Equal(t, "normal", *got.Category)
	assert.Equal(t, 3, *got.UsesPerUnit)
	assert.Equal(t, 2, *got.MinQty)
	assert.Equal(t, false, *got.IsActive)
}

func TestUpdateStock_Handler_OnlySentFieldsAreSet(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin", "")

	resp, _ := doRequest(t, app, http.MethodPatch, "/stocks/"+validStockID, `{"min_qty":0}`)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	got := svc.lastUpdate
	assert.False(t, got.PriceSet, "ไม่ได้ส่ง price ห้ามถือว่าส่ง")
	assert.Nil(t, got.StockName)
	assert.Nil(t, got.IsSellable)
	require.NotNil(t, got.MinQty)
	assert.Equal(t, 0, *got.MinQty, "ส่ง 0 มาต้องแยกออกจากไม่ได้ส่ง")
}

// body ผิดชนิด/ผิดรูป ต้อง 400 และห้ามเรียก service (เดิมถูกเมินเงียบๆ หรือราคากลายเป็น 0)
func TestUpdateStock_Handler_BadFields_Returns400(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		detail string
	}{
		{"price เป็นข้อความไม่ใช่ตัวเลข", `{"price":"abc"}`, "price"},
		{"price เป็น null", `{"price":null}`, "price"},
		{"price เป็น bool", `{"price":true}`, "price"},
		{"price เป็น object", `{"price":{}}`, "price"},
		{"usesPerUnit เป็นทศนิยม", `{"uses_per_unit":2.7}`, "uses_per_unit"},
		{"usesPerUnit เป็นข้อความ", `{"uses_per_unit":"3"}`, "uses_per_unit"},
		{"usesPerUnit ใหญ่เกิน int32", `{"uses_per_unit":1e20}`, "uses_per_unit"},
		{"minQty เป็นทศนิยม", `{"min_qty":1.5}`, "min_qty"},
		{"stockName ไม่ใช่ string", `{"stock_name":123}`, "stock_name"},
		{"isSellable ไม่ใช่ bool", `{"is_sellable":"yes"}`, "is_sellable"},
		{"category ไม่ใช่ string", `{"category":1}`, "category"},
		{"isActive ไม่ใช่ bool", `{"is_active":"true"}`, "is_active"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{}
			app := newTestApp(svc, "admin", "")

			resp, body := doRequest(t, app, http.MethodPatch, "/stocks/"+validStockID, tt.body)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			assert.Contains(t, errorDetail(t, body), tt.detail)
			assert.Equal(t, 0, svc.updateCalls)
		})
	}
}

func TestUpdateStock_Handler_InvalidJSON_Returns400(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin", "")

	resp, _ := doRequest(t, app, http.MethodPatch, "/stocks/"+validStockID, `{"price":`)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, 0, svc.updateCalls)
}

func TestUpdateStock_Handler_NotFound_Returns404(t *testing.T) {
	svc := &mockService{
		updateStock: func(ctx context.Context, id string, data UpdateStockRequest) (models.Stock, error) {
			return models.Stock{}, _package.NewAppError("Stock not found", _package.CodeNotFound, "")
		},
	}
	app := newTestApp(svc, "admin", "")

	resp, _ := doRequest(t, app, http.MethodPatch, "/stocks/"+validStockID, `{"min_qty":1}`)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ---------- DELETE / RESTORE ----------

func TestDeleteStock_Handler(t *testing.T) {
	t.Run("สำเร็จ 200", func(t *testing.T) {
		app := newTestApp(&mockService{}, "admin", "")

		resp, _ := doRequest(t, app, http.MethodDelete, "/stocks/"+validStockID, "")

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("ไม่พบ 404", func(t *testing.T) {
		svc := &mockService{deleteStock: func(ctx context.Context, id string) error {
			return _package.NewAppError("Stock not found", _package.CodeNotFound, "")
		}}
		app := newTestApp(svc, "admin", "")

		resp, _ := doRequest(t, app, http.MethodDelete, "/stocks/"+validStockID, "")

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestRestoreStock_Handler(t *testing.T) {
	t.Run("สำเร็จ 200", func(t *testing.T) {
		app := newTestApp(&mockService{}, "admin", "")

		resp, _ := doRequest(t, app, http.MethodPost, "/stocks/"+validStockID+"/restore", "")

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("ชื่อซ้ำ 409", func(t *testing.T) {
		svc := &mockService{restore: func(ctx context.Context, id string) (models.Stock, error) {
			return models.Stock{}, duplicateNameOnRestore("x")
		}}
		app := newTestApp(svc, "admin", "")

		resp, _ := doRequest(t, app, http.MethodPost, "/stocks/"+validStockID+"/restore", "")

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})
}

// ---------- wholeNumber ----------

func TestWholeNumber(t *testing.T) {
	tests := []struct {
		name   string
		in     any
		want   int
		wantOK bool
	}{
		{"จำนวนเต็ม", float64(3), 3, true},
		{"ศูนย์", float64(0), 0, true},
		{"ติดลบ (ปล่อยให้ service ตรวจเอง)", float64(-4), -4, true},
		{"ค่าสูงสุด int32", float64(2147483647), 2147483647, true},
		{"ทศนิยม", 2.7, 0, false},
		{"ข้อความ", "3", 0, false},
		{"nil", nil, 0, false},
		{"bool", true, 0, false},
		{"เกิน int32", float64(2147483648), 0, false},
		{"ต่ำกว่า int32", float64(-2147483649), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := wholeNumber(tt.in)

			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// ---------- เพิ่มเติม ----------

type failingListService struct{ mockService }

func (f *failingListService) GetAllStock(ctx context.Context) ([]models.Stock, error) {
	return nil, assert.AnError
}

func TestGetAllStock_ServiceError_Returns500(t *testing.T) {
	app := newTestApp(&failingListService{}, "admin", "")

	resp, body := doRequest(t, app, http.MethodGet, "/stocks", "")

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, false, body["success"])
}

// error ภายใน (เช่น DB) ต้องไม่รั่วออกไปให้ client เห็น
func TestUnknownServiceError_DoesNotLeakInternalMessage(t *testing.T) {
	app := newTestApp(&failingListService{}, "admin", "")

	_, body := doRequest(t, app, http.MethodGet, "/stocks", "")

	assert.Equal(t, "Something went wrong", body["message"])
	assert.NotContains(t, errorDetail(t, body), assert.AnError.Error())
}

func TestUpdateStock_Handler_PriceAcceptsNumberAndNumericString(t *testing.T) {
	tests := []struct{ name, body, want string }{
		{"ตัวเลข", `{"price":120.5}`, "120.5"},
		{"สตริงตัวเลข", `{"price":"120.5"}`, "120.5"},
		{"สตริงมีช่องว่างหน้าหลัง", `{"price":" 99 "}`, "99"},
		{"ศูนย์", `{"price":0}`, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{}
			app := newTestApp(svc, "admin", "")

			resp, _ := doRequest(t, app, http.MethodPatch, "/stocks/"+validStockID, tt.body)

			require.Equal(t, http.StatusOK, resp.StatusCode)
			assert.True(t, svc.lastUpdate.PriceSet)
			assert.Equal(t, tt.want, svc.lastUpdate.Price.String())
		})
	}
}

func TestUpdateStock_Handler_PassesStockIdFromPath(t *testing.T) {
	var gotID string
	svc := &mockService{updateStock: func(ctx context.Context, id string, data UpdateStockRequest) (models.Stock, error) {
		gotID = id
		return models.Stock{ID: id}, nil
	}}
	app := newTestApp(svc, "admin", "")

	doRequest(t, app, http.MethodPatch, "/stocks/"+validStockID, `{"min_qty":1}`)

	assert.Equal(t, validStockID, gotID)
}

func TestAddStock_Handler_InvalidJSON_Returns400(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin", "")

	resp, _ := doRequest(t, app, http.MethodPost, "/stocks", `{"stock_name":`)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, 0, svc.addCalls)
}

func TestDeleteAndRestore_Handler_PassIdToService(t *testing.T) {
	var deleted, restored string
	svc := &mockService{
		deleteStock: func(ctx context.Context, id string) error { deleted = id; return nil },
		restore: func(ctx context.Context, id string) (models.Stock, error) {
			restored = id
			return models.Stock{ID: id}, nil
		},
	}
	app := newTestApp(svc, "admin", "")

	doRequest(t, app, http.MethodDelete, "/stocks/"+validStockID, "")
	doRequest(t, app, http.MethodPost, "/stocks/"+validStockID+"/restore", "")

	assert.Equal(t, validStockID, deleted)
	assert.Equal(t, validStockID, restored)
}
