package stocklog

import (
	"backend/internal/middleware"
	_package "backend/package"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Get("/stock-logs/quantities", middleware.AuthMiddleware, h.GetAllStockQuantities)
	router.Get("/stock-logs", middleware.AuthMiddleware, h.GetStockLogsByBranchId)
	router.Post("/stock-logs", middleware.AuthMiddleware, h.AddStockLogMovement)
}

// GetAllStockQuantities GET /api/stock-logs/quantities?branchId=&search=&category=&lowStock=&active=&page=&pageSize=
func (h *Handler) GetAllStockQuantities(c fiber.Ctx) error {

	userRole, _ := c.Locals("role").(string)
	sessionBranchId, _ := c.Locals("branchId").(string)

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

	userRole, _ := c.Locals("role").(string)
	sessionBranchId, _ := c.Locals("branchId").(string)

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

// AddStockLogMovement POST /api/stock-logs
func (h *Handler) AddStockLogMovement(c fiber.Ctx) error {

	userRole, _ := c.Locals("role").(string)
	sessionBranchId, _ := c.Locals("branchId").(string)

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
