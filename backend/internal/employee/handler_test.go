package employee

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/currentuser"
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	branchA = "11111111-1111-1111-1111-111111111111"
	branchB = "22222222-2222-2222-2222-222222222222"
)

type mockRepo struct {
	Repository
	calls      int
	lastBranch string
}

func (m *mockRepo) GetActiveEmployees(_ context.Context, branchId string) ([]models.Employee, error) {
	m.calls++
	m.lastBranch = branchId
	return []models.Employee{{ID: "e1", Name: "สมหญิง", Role: "TP", BranchID: branchA}}, nil
}

func newTestApp(repo Repository, user *currentuser.CurrentUser) *fiber.App {
	app := fiber.New()
	fakeAuth := func(c fiber.Ctx) error {
		if user != nil {
			currentuser.Set(c, user)
		}
		return c.Next()
	}
	NewHandler(repo).registerRoutes(app.Group("/api"), fakeAuth)
	return app
}

func get(t *testing.T, app *fiber.App, url string) (int, map[string]any) {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, url, nil))
	require.NoError(t, err)
	defer resp.Body.Close()
	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return resp.StatusCode, body
}

func TestGetActiveEmployees_AdminCanFilterBranch(t *testing.T) {
	repo := &mockRepo{}
	app := newTestApp(repo, &currentuser.CurrentUser{Role: "admin"})

	status, body := get(t, app, "/api/employees?branch_id="+branchB)

	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, branchB, repo.lastBranch)
	data := body["data"].([]any)
	require.Len(t, data, 1)
	assert.Equal(t, "สมหญิง", data[0].(map[string]any)["name"])
}

func TestGetActiveEmployees_AdminWithoutBranchSeesAll(t *testing.T) {
	repo := &mockRepo{}
	app := newTestApp(repo, &currentuser.CurrentUser{Role: "admin"})

	status, _ := get(t, app, "/api/employees")

	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "", repo.lastBranch)
}

func TestGetActiveEmployees_BranchAccountLockedToOwnBranch(t *testing.T) {
	repo := &mockRepo{}
	own := branchA
	app := newTestApp(repo, &currentuser.CurrentUser{Role: "branch", BranchID: &own})

	// ส่งสาขาอื่นมา ต้องถูกเมินและใช้สาขาของตัวเองแทน
	status, _ := get(t, app, "/api/employees?branch_id="+branchB)

	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, branchA, repo.lastBranch)
}

func TestGetActiveEmployees_BranchAccountWithoutBranchForbidden(t *testing.T) {
	repo := &mockRepo{}
	app := newTestApp(repo, &currentuser.CurrentUser{Role: "branch"})

	status, _ := get(t, app, "/api/employees")

	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, 0, repo.calls)
}

func TestGetActiveEmployees_InvalidBranchID(t *testing.T) {
	repo := &mockRepo{}
	app := newTestApp(repo, &currentuser.CurrentUser{Role: "admin"})

	status, _ := get(t, app, "/api/employees?branch_id=not-a-uuid")

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, 0, repo.calls)
}

func TestGetActiveEmployees_NoUser(t *testing.T) {
	app := newTestApp(&mockRepo{}, nil)

	status, _ := get(t, app, "/api/employees")

	assert.Equal(t, http.StatusUnauthorized, status)
}
