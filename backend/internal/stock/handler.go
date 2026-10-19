package stock

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
	router.Get("/stock", middleware.AuthMiddleware, h.GetAllStock)
	router.Get("/stock/quantities", middleware.AuthMiddleware, h.GetStockQuantities)
	router.Get("/stock/:id/containers", middleware.AuthMiddleware, h.GetStockContainersByStockID)
}

// GetStockQuantities GET /api/stock/quantities?branchId=&search=&category=premium|normal&lowStock=true
// สินค้าแต่ละชนิดพร้อมจำนวนคงเหลือและราคา
func (h *Handler) GetStockQuantities(c fiber.Ctx) error {

	userRole, _ := c.Locals("role").(string)
	sessionBranchId, _ := c.Locals("branchId").(string)

	q := QuantityQuery{
		BranchID: c.Query("branchId"),
		Search:   c.Query("search"),
		Category: c.Query("category"),
		LowStock: c.Query("lowStock"),
	}

	quantities, err := h.service.GetStockQuantities(c.Context(), userRole, sessionBranchId, q)
	if err != nil {
		return _package.SendError(c, err, "stock")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Stock quantities retrieved successfully", quantities))
}

// Get All Stock GET /api/stock
func (h *Handler) GetAllStock(c fiber.Ctx) error {

	stocks, err := h.service.GetAllStock(c.Context())
	if err != nil {
		return _package.SendError(c, err, "stock")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Stocks retrieved successfully", stocks))
}

// GetStockContainersByStockID GET /api/stock/:id/containers?status=open|closed|all&branchId=&page=1&pageSize=20
func (h *Handler) GetStockContainersByStockID(c fiber.Ctx) error {

	userRole, _ := c.Locals("role").(string)
	sessionBranchId, _ := c.Locals("branchId").(string)

	q := ContainerQuery{
		StockID:  c.Params("id"),
		BranchID: c.Query("branchId"),
		Status:   c.Query("status"),
		Page:     c.Query("page"),
		PageSize: c.Query("pageSize"),
	}

	containers, meta, err := h.service.GetStockContainersByStockID(c.Context(), userRole, sessionBranchId, q)
	if err != nil {
		return _package.SendError(c, err, "stock")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessPagedResponse("Stock containers retrieved successfully", containers, meta))
}
