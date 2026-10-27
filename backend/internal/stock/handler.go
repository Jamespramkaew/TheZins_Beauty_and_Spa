package stock

import (
	"backend/internal/currentuser"
	"backend/internal/middleware"
	_package "backend/package"
	"math"
	"strings"

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
	stocks := router.Group("/stocks", middleware.AuthMiddleware)
	stocks.Get("/", h.GetAllStock)
	stocks.Get("/:id/containers", h.GetStockContainersByStockID)
	stocks.Post("/", middleware.ReqEmptyCheck, h.AddStock)
	stocks.Patch("/:id", middleware.ReqEmptyCheck, h.UpdateStockById)
	stocks.Delete("/:id", h.DeleteStockById)
	stocks.Post("/:id/restore", h.RestoreStockById)
}

// Get All Stock GET /api/stocks
func (h *Handler) GetAllStock(c fiber.Ctx) error {

	stocks, err := h.service.GetAllStock(c.Context())
	if err != nil {
		return _package.SendError(c, err, "stock")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Stocks retrieved successfully", stocks))
}

// GetStockContainersByStockID GET /api/stocks/:id/containers?status=open|closed|all&branchId=&page=1&pageSize=20
func (h *Handler) GetStockContainersByStockID(c fiber.Ctx) error {

	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c,
			_package.NewAppError("Unauthorized", _package.CodeUnauthorized, "User not found in context"),
			"stock")
	}

	q := ContainerQuery{
		StockID:  c.Params("id"),
		BranchID: c.Query("branchId"),
		Status:   c.Query("status"),
		Page:     c.Query("page"),
		PageSize: c.Query("pageSize"),
	}

	containers, meta, err := h.service.GetStockContainersByStockID(c.Context(), user, q)
	if err != nil {
		return _package.SendError(c, err, "stock")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessPagedResponse("Stock containers retrieved successfully", containers, meta))
}

func (h *Handler) AddStock(c fiber.Ctx) error {

	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c,
			_package.NewAppError("Unauthorized", _package.CodeUnauthorized, "User not found in context"),
			"stock")
	}

	if user.Role != "admin" {
		return _package.SendJSON(c, fiber.StatusForbidden,
			_package.ErrorResponse("Access denied", _package.CodeForbidden, "Only admin can create stock"))
	}

	var req AddStockRequest

	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	stock, err := h.service.AddStock(c.Context(), req)
	if err != nil {
		return _package.SendError(c, err, "stock")
	}

	return _package.SendJSON(c, fiber.StatusCreated,
		_package.SuccessResponse("Stock created successfully", stock))
}

func (h *Handler) UpdateStockById(c fiber.Ctx) error {

	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c,
			_package.NewAppError("Unauthorized", _package.CodeUnauthorized, "User not found in context"),
			"stock")
	}

	if user.Role != "admin" {
		return _package.SendJSON(c, fiber.StatusForbidden,
			_package.ErrorResponse("Access denied", _package.CodeForbidden, "Only admin can update stock"))
	}

	stockId := c.Params("id")

	// อ่าน raw JSON เพื่อเช็คว่า field ไหนถูกส่งมาบ้าง
	var rawBody map[string]interface{}
	if err := c.Bind().JSON(&rawBody); err != nil {
		return _package.SendBadBody(c, err)
	}

	var req UpdateStockRequest

	// StockName
	if val, exists := rawBody["stock_name"]; exists {
		strVal, ok := val.(string)
		if !ok {
			return badField(c, "stock_name must be a string")
		}
		req.StockName = &strVal
	}

	// Price
	// ถ้าส่ง price มา ต้องเป็นตัวเลขหรือสตริงตัวเลขเท่านั้น ไม่งั้นตอบ 400
	// (ห้ามเมินแล้วปล่อยเป็น 0 เพราะ service จะเอา 0 ไปเขียนทับราคาเดิม)
	if val, exists := rawBody["price"]; exists {
		switch v := val.(type) {
		case float64:
			req.Price = decimal.NewFromFloat(v)
		case string:
			price, err := decimal.NewFromString(strings.TrimSpace(v))
			if err != nil {
				return badField(c, "price must be a number")
			}
			req.Price = price
		default: // null, bool, object, array
			return badField(c, "price must be a number")
		}
		req.PriceSet = true
	}

	// IsSellable
	if val, exists := rawBody["is_sellable"]; exists {
		boolVal, ok := val.(bool)
		if !ok {
			return badField(c, "is_sellable must be a boolean")
		}
		req.IsSellable = &boolVal
	}

	// Category
	if val, exists := rawBody["category"]; exists {
		strVal, ok := val.(string)
		if !ok {
			return badField(c, "category must be a string")
		}
		req.Category = &strVal
	}

	// UsesPerUnit
	if val, exists := rawBody["uses_per_unit"]; exists {
		intVal, ok := wholeNumber(val)
		if !ok {
			return badField(c, "uses_per_unit must be a whole number")
		}
		req.UsesPerUnit = &intVal
	}

	// MinQty
	if val, exists := rawBody["min_qty"]; exists {
		intVal, ok := wholeNumber(val)
		if !ok {
			return badField(c, "min_qty must be a whole number")
		}
		req.MinQty = &intVal
	}

	// IsActive
	if val, exists := rawBody["is_active"]; exists {
		boolVal, ok := val.(bool)
		if !ok {
			return badField(c, "is_active must be a boolean")
		}
		req.IsActive = &boolVal
	}

	stock, err := h.service.UpdateStockById(c.Context(), stockId, req)
	if err != nil {
		return _package.SendError(c, err, "stock")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Stock updated successfully", stock))
}

// badField ตอบ 400 เมื่อ field ใน body ชนิดไม่ตรง (ห้ามเมินเงียบๆ ไม่งั้น client เข้าใจว่าบันทึกแล้ว)
func badField(c fiber.Ctx, detail string) error {
	return _package.SendError(c,
		_package.NewAppError("Invalid request", _package.CodeValidation, detail),
		"stock")
}

// wholeNumber แปลงค่า JSON เป็น int เฉพาะเมื่อเป็นจำนวนเต็มที่อยู่ในช่วง int32
// (2.7, "3", null, 1e20 ไม่ผ่าน ไม่ตัดทศนิยมทิ้งเงียบๆ)
func wholeNumber(val any) (int, bool) {
	f, ok := val.(float64)
	if !ok || f != math.Trunc(f) || f < math.MinInt32 || f > math.MaxInt32 {
		return 0, false
	}
	return int(f), true
}

// DeleteStockById DELETE /api/stocks/:id
func (h *Handler) DeleteStockById(c fiber.Ctx) error {

	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c,
			_package.NewAppError("Unauthorized", _package.CodeUnauthorized, "User not found in context"),
			"stock")
	}

	if user.Role != "admin" {
		return _package.SendJSON(c, fiber.StatusForbidden,
			_package.ErrorResponse("Access denied", _package.CodeForbidden, "Only admin can delete stock"))
	}

	stockId := c.Params("id")

	err := h.service.DeleteStockById(c.Context(), stockId)
	if err != nil {
		return _package.SendError(c, err, "stock")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Stock deleted successfully", nil))
}

// RestoreStockById POST /api/stocks/:id/restore
func (h *Handler) RestoreStockById(c fiber.Ctx) error {

	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c,
			_package.NewAppError("Unauthorized", _package.CodeUnauthorized, "User not found in context"),
			"stock")
	}

	if user.Role != "admin" {
		return _package.SendJSON(c, fiber.StatusForbidden,
			_package.ErrorResponse("Access denied", _package.CodeForbidden, "Only admin can restore stock"))
	}

	stockId := c.Params("id")

	stock, err := h.service.RestoreStockById(c.Context(), stockId)
	if err != nil {
		return _package.SendError(c, err, "stock")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Stock restored successfully", stock))
}
