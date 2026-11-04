package branch

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/currentuser"
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	Repository
	err error
}

func (m *mockRepo) GetActiveBranches(context.Context) ([]models.Branch, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []models.Branch{{ID: "b1", Code: "S1", Name: "สาขา 1", IsActive: true}, {ID: "b2", Code: "S2", Name: "สาขา 2", IsActive: true}}, nil
}

func newTestApp(repo Repository, user *currentuser.CurrentUser) *fiber.App {
	app := fiber.New()
	fakeAuth := func(c fiber.Ctx) error {
		currentuser.Set(c, user)
		return c.Next()
	}
	NewHandler(repo).registerRoutes(app.Group("/api"), fakeAuth)
	return app
}

// บัญชีสาขาต้องเห็นสาขาอื่นด้วย (ใช้เลือกสาขาปลายทางตอนโอน) endpoint นี้ไม่มีข้อมูลสต็อก
func TestGetActiveBranches_BranchAccountSeesAllBranches(t *testing.T) {
	own := "b1"
	app := newTestApp(&mockRepo{}, &currentuser.CurrentUser{Role: "branch", BranchID: &own})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/branches", nil))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var body struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Len(t, body.Data, 2)
	assert.Equal(t, "S1", body.Data[0]["code"])
}

func TestGetActiveBranches_RepositoryError(t *testing.T) {
	app := newTestApp(&mockRepo{err: errors.New("db down")}, &currentuser.CurrentUser{Role: "admin"})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/branches", nil))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
