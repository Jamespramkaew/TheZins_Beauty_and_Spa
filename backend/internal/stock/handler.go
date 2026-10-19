package stock

import (
	"backend/internal/middleware"
	_package "backend/package"

	"github.com/gofiber/fiber/v3"
	"github.com/shopspring/decimal"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Get("/stock", middleware.AuthMiddleware, h.GetAllStock)
	router.Get("/stock/:id/containers", middleware.AuthMiddleware, h.GetStockContainersByStockID)
	router.Post("/stock", middleware.AuthMiddleware, middleware.ReqEmptyCheck, h.AddStock)
	router.Patch("/stock/:stockId", middleware.AuthMiddleware, middleware.ReqEmptyCheck, h.UpdateStockById)
	router.Delete("/stock/:stockId", middleware.AuthMiddleware, h.DeleteStockById)
	router.Post("/stock/:stockId/restore", middleware.AuthMiddleware, h.RestoreStockById)
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

func (h *Handler) AddStock(c fiber.Ctx) error {

	userRole, _ := c.Locals("role").(string)
	if userRole != "admin" {
		return _package.SendJSON(c, fiber.StatusForbidden,
			_package.ErrorResponse("Access denied", _package.CodeForbidden, "Only admin can create stock"))
	}

	var req AddStockRequest

	if err := c.Bind().Body(&req); err != nil {
		return _package.SendError(c, err, "stock")
	}

	stock, err := h.service.AddStock(c.Context(), req)
	if err != nil {
		return _package.SendError(c, err, "stock")
	}

	return _package.SendJSON(c, fiber.StatusCreated,
		_package.SuccessResponse("Stock created successfully", stock))
}

func (h *Handler) UpdateStockById(c fiber.Ctx) error {

	userRole, _ := c.Locals("role").(string)
	if userRole != "admin" {
		return _package.SendJSON(c, fiber.StatusForbidden,
			_package.ErrorResponse("Access denied", _package.CodeForbidden, "Only admin can update stock"))
	}

	stockId := c.Params("stockId")

	// อ่าน raw JSON เพื่อเช็คว่า field ไหนถูกส่งมาบ้าง
	var rawBody map[string]interface{}
	if err := c.Bind().JSON(&rawBody); err != nil {
		return _package.SendError(c, err, "stock")
	}

	var req UpdateStockRequest

	// StockName
	if val, exists := rawBody["stockName"]; exists {
		if strVal, ok := val.(string); ok {
			req.StockName = &strVal
		}
	}

	// Price
	if val, exists := rawBody["price"]; exists {
		req.PriceSet = true
		switch v := val.(type) {
		case float64:
			req.Price = decimal.NewFromFloat(v)
		case string:
			if price, err := decimal.NewFromString(v); err == nil {
				req.Price = price
			}
		case int:
			req.Price = decimal.NewFromInt(int64(v))
		}
	}

	// IsSellable
	if val, exists := rawBody["isSellable"]; exists {
		if boolVal, ok := val.(bool); ok {
			req.IsSellable = &boolVal
		}
	}

	// Category
	if val, exists := rawBody["category"]; exists {
		if strVal, ok := val.(string); ok {
			req.Category = &strVal
		}
	}

	// UsesPerUnit
	if val, exists := rawBody["usesPerUnit"]; exists {
		if floatVal, ok := val.(float64); ok {
			intVal := int(floatVal)
			req.UsesPerUnit = &intVal
		}
	}

	// MinQty
	if val, exists := rawBody["minQty"]; exists {
		if floatVal, ok := val.(float64); ok {
			intVal := int(floatVal)
			req.MinQty = &intVal
		}
	}

	// IsActive
	if val, exists := rawBody["isActive"]; exists {
		if boolVal, ok := val.(bool); ok {
			req.IsActive = &boolVal
		}
	}

	stock, err := h.service.UpdateStockById(c.Context(), stockId, req)
	if err != nil {
		return _package.SendError(c, err, "stock")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Stock updated successfully", stock))
}

// DeleteStockById DELETE /api/stock/:stockId
func (h *Handler) DeleteStockById(c fiber.Ctx) error {

	userRole, _ := c.Locals("role").(string)
	if userRole != "admin" {
		return _package.SendJSON(c, fiber.StatusForbidden,
			_package.ErrorResponse("Access denied", _package.CodeForbidden, "Only admin can delete stock"))
	}

	stockId := c.Params("stockId")

	err := h.service.DeleteStockById(c.Context(), stockId)
	if err != nil {
		return _package.SendError(c, err, "stock")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Stock deleted successfully", nil))
}

// RestoreStockById POST /api/stock/:stockId/restore
func (h *Handler) RestoreStockById(c fiber.Ctx) error {

	userRole, _ := c.Locals("role").(string)
	if userRole != "admin" {
		return _package.SendJSON(c, fiber.StatusForbidden,
			_package.ErrorResponse("Access denied", _package.CodeForbidden, "Only admin can restore stock"))
	}

	stockId := c.Params("stockId")

	stock, err := h.service.RestoreStockById(c.Context(), stockId)
	if err != nil {
		return _package.SendError(c, err, "stock")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Stock restored successfully", stock))
}
