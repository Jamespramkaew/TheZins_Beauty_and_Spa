package employee

import (
	"regexp"
	"strings"

	"backend/internal/currentuser"
	"backend/internal/middleware"
	_package "backend/package"

	"github.com/gofiber/fiber/v3"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// EmployeeResponse = ข้อมูลพนักงานที่หน้าจอใช้เลือกผู้ทำรายการ
type EmployeeResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	BranchID string `json:"branch_id"`
}

type Handler struct {
	repository Repository
}

func NewHandler(repository Repository) *Handler {
	return &Handler{repository: repository}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	h.registerRoutes(router, middleware.AuthMiddleware)
}

// registerRoutes แยก auth ออกมาให้ test ใส่ตัวตั้งผู้ใช้ปลอมแทน session ได้
func (h *Handler) registerRoutes(router fiber.Router, auth fiber.Handler) {
	employees := router.Group("/employees", auth)
	employees.Get("", h.GetActiveEmployees)
}

// GetActiveEmployees GET /api/employees?branch_id=
// admin: เลือกสาขาด้วย branch_id ได้ ไม่ส่งมา = ทุกสาขา
// บัญชีสาขา: ถูกบังคับเป็นสาขาของตัวเองเสมอ (ไม่สนค่า branch_id กันดูข้ามสาขา)
func (h *Handler) GetActiveEmployees(c fiber.Ctx) error {
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c,
			_package.NewAppError("Login required", _package.CodeUnauthorized, ""), "employee")
	}

	var branchId string
	if user.Role == "admin" {
		branchId = strings.TrimSpace(c.Query("branch_id"))
		if branchId != "" && !uuidPattern.MatchString(branchId) {
			return _package.SendError(c,
				_package.NewAppError("Invalid request", _package.CodeValidation, "branch_id must be a valid UUID"), "employee")
		}
	} else {
		if user.BranchID == nil || *user.BranchID == "" {
			return _package.SendError(c,
				_package.NewAppError("Branch is required", _package.CodeForbidden, "Your account is not linked to a branch"), "employee")
		}
		branchId = *user.BranchID
	}

	employees, err := h.repository.GetActiveEmployees(c.Context(), branchId)
	if err != nil {
		return _package.SendError(c, err, "employee")
	}

	items := make([]EmployeeResponse, 0, len(employees))
	for _, e := range employees {
		items = append(items, EmployeeResponse{ID: e.ID, Name: e.Name, Role: e.Role, BranchID: e.BranchID})
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Employees retrieved successfully", items))
}
