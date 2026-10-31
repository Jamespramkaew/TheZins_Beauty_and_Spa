package course

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/currentuser"
	"backend/internal/models"
	_package "backend/package"

	"github.com/gofiber/fiber/v3"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- Mock Service ----------
// handler test ไม่แตะ service จริง ใช้ mock เพื่อดูว่า handler ส่งอะไรเข้า service
// และแปลงผลลัพธ์/error เป็น HTTP status ถูกไหม

type mockService struct {
	getAll       func(ctx context.Context, q GetAllCourseQuery) ([]models.Course, _package.PageMeta, error)
	addCourse    func(ctx context.Context, data AddCourseRequest) (models.Course, error)
	updateCourse func(ctx context.Context, id string, data UpdateCourseRequest) (models.Course, error)
	deleteCourse func(ctx context.Context, id string) error

	calls      int // นับทุกครั้งที่เรียก service (ใช้เช็คว่าโดนกันก่อนถึง service)
	lastQuery  GetAllCourseQuery
	lastAdd    AddCourseRequest
	lastUpdate UpdateCourseRequest
	lastID     string
}

func (m *mockService) GetAllCourse(ctx context.Context, q GetAllCourseQuery) ([]models.Course, _package.PageMeta, error) {
	m.calls++
	m.lastQuery = q
	if m.getAll != nil {
		return m.getAll(ctx, q)
	}
	return []models.Course{{ID: idFacial, Name: "Facial"}}, _package.NewPageMeta(1, 20, 1), nil
}

func (m *mockService) AddCourse(ctx context.Context, data AddCourseRequest) (models.Course, error) {
	m.calls++
	m.lastAdd = data
	if m.addCourse != nil {
		return m.addCourse(ctx, data)
	}
	return models.Course{ID: idFacial, Name: data.Name}, nil
}

func (m *mockService) UpdateCourseById(ctx context.Context, id string, data UpdateCourseRequest) (models.Course, error) {
	m.calls++
	m.lastID, m.lastUpdate = id, data
	if m.updateCourse != nil {
		return m.updateCourse(ctx, id, data)
	}
	return models.Course{ID: id}, nil
}

func (m *mockService) DeleteCourseById(ctx context.Context, id string) error {
	m.calls++
	m.lastID = id
	if m.deleteCourse != nil {
		return m.deleteCourse(ctx, id)
	}
	return nil
}

// ---------- Helpers ----------

// newTestApp ใช้ route ชุดเดียวกับของจริง (รวม RequireAdminMiddleware / ReqEmptyCheck)
// แต่แทน AuthMiddleware (ต้องมี session) ด้วยตัวตั้ง currentuser ตาม role ที่ให้มา
func newTestApp(svc Service, role string) *fiber.App {
	app := fiber.New()
	fakeAuth := func(c fiber.Ctx) error {
		user := &currentuser.CurrentUser{ID: "u1", Role: role}
		if role != "admin" {
			branchID := "b1"
			user.BranchID = &branchID
		}
		currentuser.Set(c, user)
		return c.Next()
	}
	NewHandler(svc).registerRoutes(app, fakeAuth)
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

// ---------- สิทธิ์ ----------

// RegisterRoutes ของจริงต้องติด AuthMiddleware ที่ group: ไม่มี session = 401 ทุกเส้น
func TestRoutes_WithoutLogin_Returns401(t *testing.T) {
	tests := []struct{ method, path, body string }{
		{http.MethodGet, "/courses", ""},
		{http.MethodPost, "/courses", `{"name":"x"}`},
		{http.MethodPatch, "/courses/" + idFacial, `{"name":"x"}`},
		{http.MethodDelete, "/courses/" + idFacial, ""},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			svc := &mockService{}
			app := fiber.New()
			NewHandler(svc).RegisterRoutes(app)

			resp, _ := doRequest(t, app, tt.method, tt.path, tt.body)

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			assert.Equal(t, 0, svc.calls, "ห้ามเรียก service")
		})
	}
}

func TestRoutes_NonAdminIsForbidden(t *testing.T) {
	tests := []struct{ name, method, path, body string }{
		{"POST /courses", http.MethodPost, "/courses", `{"name":"x","price":1,"max_session":1}`},
		{"PATCH /courses/:id", http.MethodPatch, "/courses/" + idFacial, `{"name":"x"}`},
		{"DELETE /courses/:id", http.MethodDelete, "/courses/" + idFacial, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{}
			app := newTestApp(svc, "branch")

			resp, body := doRequest(t, app, tt.method, tt.path, tt.body)

			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
			assert.Equal(t, false, body["success"])
			assert.Equal(t, 0, svc.calls, "ห้ามเรียก service")
		})
	}
}

// ดูรายการคอร์สได้ทุกคนที่ login (ไม่ใช่เฉพาะ admin)
func TestGetAllCourse_BranchUserAllowed(t *testing.T) {
	app := newTestApp(&mockService{}, "branch")

	resp, _ := doRequest(t, app, http.MethodGet, "/courses", "")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ---------- GET ----------

func TestGetAllCourse_PassesSnakeCaseQueryAndReturnsMeta(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin")

	resp, body := doRequest(t, app, http.MethodGet, "/courses?search=%E0%B8%AB%E0%B8%99%E0%B9%89%E0%B8%B2&page=2&page_size=10", "")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, GetAllCourseQuery{Search: "หน้า", Page: "2", PageSize: "10"}, svc.lastQuery)

	meta, ok := body["meta"].(map[string]any)
	require.True(t, ok, "endpoint แบ่งหน้าต้องมี meta: %v", body)
	for _, key := range []string{"page", "page_size", "total", "total_pages"} {
		assert.Contains(t, meta, key)
	}
}

// ชื่อ camelCase แบบเก่าต้องไม่ถูกอ่าน
func TestGetAllCourse_IgnoresCamelCasePageSize(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin")

	doRequest(t, app, http.MethodGet, "/courses?pageSize=10", "")

	assert.Equal(t, "", svc.lastQuery.PageSize)
}

func TestGetAllCourse_ServiceErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"query ผิด 400", invalid("page must be a number between 1 and 1000000"), http.StatusBadRequest},
		{"error ไม่รู้จัก 500", assert.AnError, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{getAll: func(ctx context.Context, q GetAllCourseQuery) ([]models.Course, _package.PageMeta, error) {
				return nil, _package.PageMeta{}, tt.err
			}}
			app := newTestApp(svc, "admin")

			resp, body := doRequest(t, app, http.MethodGet, "/courses?page=abc", "")

			assert.Equal(t, tt.status, resp.StatusCode)
			assert.Equal(t, false, body["success"])
		})
	}
}

// error ภายใน (เช่น DB) ต้องไม่รั่วออกไปให้ client เห็น
func TestUnknownServiceError_DoesNotLeakInternalMessage(t *testing.T) {
	svc := &mockService{getAll: func(ctx context.Context, q GetAllCourseQuery) ([]models.Course, _package.PageMeta, error) {
		return nil, _package.PageMeta{}, assert.AnError
	}}
	app := newTestApp(svc, "admin")

	_, body := doRequest(t, app, http.MethodGet, "/courses", "")

	assert.Equal(t, "Something went wrong", body["message"])
	assert.NotContains(t, errorDetail(t, body), assert.AnError.Error())
}

// ---------- POST ----------

func TestAddCourse_Handler_Created(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin")

	resp, body := doRequest(t, app, http.MethodPost, "/courses",
		`{"name":"คอร์สหน้าใส","price":"3500.50","max_session":10,"valid_day":180,"service_comm":150,"category":"facial"}`)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, true, body["success"])
	got := svc.lastAdd
	assert.Equal(t, "คอร์สหน้าใส", got.Name)
	assert.Equal(t, "3500.5", got.Price.String())
	assert.Equal(t, 10, got.MaxSession)
	require.NotNil(t, got.ValidDay)
	assert.Equal(t, 180, *got.ValidDay)
	assert.True(t, got.ServiceComm.Equal(decimal.NewFromInt(150)))
	assert.Equal(t, "facial", got.Category)
}

// ชื่อ camelCase แบบเก่าต้องไม่ถูกอ่าน (เงียบๆ เป็นค่าว่าง แล้ว service จะตอบ 400 เอง)
func TestAddCourse_Handler_IgnoresCamelCaseFields(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin")

	doRequest(t, app, http.MethodPost, "/courses", `{"name":"x","maxSession":5,"validDay":30,"serviceComm":10}`)

	assert.Equal(t, 0, svc.lastAdd.MaxSession)
	assert.Nil(t, svc.lastAdd.ValidDay)
	assert.True(t, svc.lastAdd.ServiceComm.IsZero())
}

func TestAddCourse_Handler_BadBody_Returns400(t *testing.T) {
	tests := []struct{ name, body, detail string }{
		{"body ว่าง", "", ""},
		{"body เป็น {}", `{}`, ""},
		{"JSON พัง", `{"name":`, ""},
		{"max_session เป็นข้อความ", `{"name":"x","max_session":"3"}`, "max_session"},
		{"name ไม่ใช่ string", `{"name":123}`, "name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{}
			app := newTestApp(svc, "admin")

			resp, body := doRequest(t, app, http.MethodPost, "/courses", tt.body)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			if tt.detail != "" {
				assert.Contains(t, errorDetail(t, body), tt.detail)
			}
			assert.Equal(t, 0, svc.calls, "ห้ามเรียก service")
		})
	}
}

func TestAddCourse_Handler_ServiceErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"validation 400", invalid("course name is required"), http.StatusBadRequest},
		{"ชื่อซ้ำ 409", duplicateName("x"), http.StatusConflict},
		{"error ไม่รู้จัก 500", assert.AnError, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{addCourse: func(ctx context.Context, data AddCourseRequest) (models.Course, error) {
				return models.Course{}, tt.err
			}}
			app := newTestApp(svc, "admin")

			resp, _ := doRequest(t, app, http.MethodPost, "/courses", `{"name":"x","price":1,"max_session":1}`)

			assert.Equal(t, tt.status, resp.StatusCode)
		})
	}
}

// ---------- PATCH ----------

func TestUpdateCourse_Handler_ParsesFields(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin")

	resp, _ := doRequest(t, app, http.MethodPatch, "/courses/"+idFacial,
		`{"name":"ใหม่","price":"750.50","max_session":3,"valid_day":30,"service_comm":20,"category":"body","is_active":false}`)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, idFacial, svc.lastID)
	got := svc.lastUpdate
	require.NotNil(t, got.Name)
	assert.Equal(t, "ใหม่", *got.Name)
	require.NotNil(t, got.Price)
	assert.Equal(t, "750.5", got.Price.String())
	assert.Equal(t, 3, *got.MaxSession)
	assert.JSONEq(t, `30`, string(got.ValidDay))
	assert.True(t, got.ServiceComm.Equal(decimal.NewFromInt(20)))
	assert.Equal(t, "body", *got.Category)
	assert.Equal(t, false, *got.IsActive)
}

func TestUpdateCourse_Handler_OnlySentFieldsAreSet(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin")

	resp, _ := doRequest(t, app, http.MethodPatch, "/courses/"+idFacial, `{"max_session":1}`)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	got := svc.lastUpdate
	assert.Nil(t, got.Name)
	assert.Nil(t, got.Price)
	assert.Nil(t, got.ValidDay, "ไม่ได้ส่ง valid_day ต้องแยกออกจากส่ง null")
	assert.Nil(t, got.ServiceComm)
	assert.Nil(t, got.Category)
	assert.Nil(t, got.IsActive)
	require.NotNil(t, got.MaxSession)
	assert.Equal(t, 1, *got.MaxSession)
}

// ส่ง null มาต้องถึง service เป็น "null" (ล้างเป็นไม่หมดอายุ) ไม่ใช่หายไปเหมือนไม่ได้ส่ง
func TestUpdateCourse_Handler_ValidDayNullReachesService(t *testing.T) {
	svc := &mockService{}
	app := newTestApp(svc, "admin")

	resp, _ := doRequest(t, app, http.MethodPatch, "/courses/"+idFacial, `{"valid_day":null}`)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "null", string(svc.lastUpdate.ValidDay))
}

func TestUpdateCourse_Handler_BadBody_Returns400(t *testing.T) {
	tests := []struct{ name, body, detail string }{
		{"body ว่าง", "", ""},
		{"body เป็น {}", `{}`, ""},
		{"JSON พัง", `{"name":`, ""},
		{"name ไม่ใช่ string", `{"name":1}`, "name"},
		{"max_session เป็นทศนิยม", `{"max_session":1.5}`, "max_session"},
		{"is_active ไม่ใช่ bool", `{"is_active":"true"}`, "is_active"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{}
			app := newTestApp(svc, "admin")

			resp, body := doRequest(t, app, http.MethodPatch, "/courses/"+idFacial, tt.body)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			if tt.detail != "" {
				assert.Contains(t, errorDetail(t, body), tt.detail)
			}
			assert.Equal(t, 0, svc.calls, "ห้ามเรียก service")
		})
	}
}

func TestUpdateCourse_Handler_ServiceErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"validation 400", invalid("valid_day must be a whole number or null"), http.StatusBadRequest},
		{"ไม่พบ 404", notFound(idFacial), http.StatusNotFound},
		{"ชื่อซ้ำ 409", duplicateName("x"), http.StatusConflict},
		{"error ไม่รู้จัก 500", assert.AnError, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{updateCourse: func(ctx context.Context, id string, data UpdateCourseRequest) (models.Course, error) {
				return models.Course{}, tt.err
			}}
			app := newTestApp(svc, "admin")

			resp, _ := doRequest(t, app, http.MethodPatch, "/courses/"+idFacial, `{"name":"x"}`)

			assert.Equal(t, tt.status, resp.StatusCode)
		})
	}
}

// ---------- DELETE ----------

func TestDeleteCourse_Handler(t *testing.T) {
	t.Run("สำเร็จ 200 และส่ง id จาก path", func(t *testing.T) {
		svc := &mockService{}
		app := newTestApp(svc, "admin")

		resp, body := doRequest(t, app, http.MethodDelete, "/courses/"+idFacial, "")

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, true, body["success"])
		assert.Equal(t, idFacial, svc.lastID)
	})
	t.Run("ไม่พบ 404", func(t *testing.T) {
		svc := &mockService{deleteCourse: func(ctx context.Context, id string) error { return notFound(id) }}
		app := newTestApp(svc, "admin")

		resp, _ := doRequest(t, app, http.MethodDelete, "/courses/"+idFacial, "")

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
	t.Run("error ไม่รู้จัก 500", func(t *testing.T) {
		svc := &mockService{deleteCourse: func(ctx context.Context, id string) error { return assert.AnError }}
		app := newTestApp(svc, "admin")

		resp, _ := doRequest(t, app, http.MethodDelete, "/courses/"+idFacial, "")

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
