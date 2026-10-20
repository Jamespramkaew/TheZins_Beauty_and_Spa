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
