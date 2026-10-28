package stocklog

import (
	"backend/internal/currentuser"
	"backend/internal/middleware"
	_package "backend/package"

	"github.com/gofiber/fiber/v3"
)

// errLoginRequired ตอบ 401 เมื่อไม่มีผู้ใช้ใน request (ไม่ได้ผ่าน AuthMiddleware)
var errLoginRequired = _package.NewAppError("Login required", _package.CodeUnauthorized, "")

// callerRoleAndBranch อ่านผู้ใช้ด้วย currentuser แล้วแปลงเป็น role + branchId ที่ service ใช้
// admin ไม่มีสาขา (BranchID = nil) → branchId เป็น ""
func callerRoleAndBranch(c fiber.Ctx) (role, branchId string, ok bool) {
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return "", "", false
	}
	if user.BranchID != nil {
		branchId = *user.BranchID
	}
	return user.Role, branchId, true
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Get("/stock-logs/quantities", middleware.AuthMiddleware, h.GetAllStockQuantities)
	router.Get("/stock-logs", middleware.AuthMiddleware, h.GetStockLogsByBranchId)
	router.Post("/stock-logs", middleware.AuthMiddleware, middleware.ReqEmptyCheck, h.AddStockLogMovement)
	router.Post("/stock-logs/transfer", middleware.AuthMiddleware, middleware.ReqEmptyCheck, h.TransferStockLog)
}

// GetAllStockQuantities GET /api/stock-logs/quantities?branchId=&search=&category=&lowStock=&active=&page=&pageSize=
func (h *Handler) GetAllStockQuantities(c fiber.Ctx) error {

	userRole, sessionBranchId, ok := callerRoleAndBranch(c)
	if !ok {
		return _package.SendError(c, errLoginRequired, "stock-log")
	}

	var q GetAllStockQuantitiesQuery
	if err := c.Bind().Query(&q); err != nil {
		return _package.SendError(c,
			_package.NewAppError("Invalid query", _package.CodeValidation, err.Error()), "stock-log")
	}

	quantities, meta, err := h.service.GetAllStockQuantities(c.Context(), userRole, sessionBranchId, q)
	if err != nil {
		return _package.SendError(c, err, "stock-log")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessPagedResponse("Stock quantities fetched successfully", quantities, meta))
}

// GetStockLogsByBranchId GET /api/stock-logs?stockId=&branchId=&logType=&direction=&from=&to=&orderItemId=&usageId=&page=&pageSize=
func (h *Handler) GetStockLogsByBranchId(c fiber.Ctx) error {

	userRole, sessionBranchId, ok := callerRoleAndBranch(c)
	if !ok {
		return _package.SendError(c, errLoginRequired, "stock-log")
	}

	var q GetAllStockLogsQuery
	if err := c.Bind().Query(&q); err != nil {
		return _package.SendError(c,
			_package.NewAppError("Invalid query", _package.CodeValidation, err.Error()), "stock-log")
	}

	logs, meta, err := h.service.GetStockLogsByBranchId(c.Context(), userRole, sessionBranchId, q)
	if err != nil {
		return _package.SendError(c, err, "stock-log")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessPagedResponse("Stock logs fetched successfully", logs, meta))
}

// AddStockLogMovement POST /api/stock-logs (receive / expired / damaged / internal / adjust)
func (h *Handler) AddStockLogMovement(c fiber.Ctx) error {

	userRole, sessionBranchId, ok := callerRoleAndBranch(c)
	if !ok {
		return _package.SendError(c, errLoginRequired, "stock-log")
	}

	var req AddStockLogMovementRequest
	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	stockLog, err := h.service.AddStockLogMovement(c.Context(), userRole, sessionBranchId, req)
	if err != nil {
		return _package.SendError(c, err, "stock-log")
	}

	return _package.SendJSON(c, fiber.StatusCreated,
		_package.SuccessResponse("Stock log movement added successfully", stockLog))
}

func (h *Handler) TransferStockLog(c fiber.Ctx) error {

	userRole, sessionBranchId, ok := callerRoleAndBranch(c)
	if !ok {
		return _package.SendError(c, errLoginRequired, "stock-log")
	}

	var req TransferStockLogRequest
	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	if err := h.service.TransferStockLog(c.Context(), userRole, sessionBranchId, req); err != nil {
		return _package.SendError(c, err, "stock-log")
	}

	return _package.SendJSON(c, fiber.StatusCreated,
		_package.SuccessResponse("Stock transfer successfully", nil))
}
