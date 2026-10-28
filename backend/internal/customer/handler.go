package customer

import (
	"backend/internal/currentuser"
	"backend/internal/middleware"
	"backend/package"

	"github.com/gofiber/fiber/v3"
)

// errLoginRequired ตอบ 401 เมื่อไม่มีผู้ใช้ (ยังไม่ login)
var errLoginRequired = _package.NewAppError("Login required", _package.CodeUnauthorized, "")

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	customers := router.Group("/customers", middleware.AuthMiddleware)
	customers.Get("", h.Search)
	customers.Get("/:id", h.GetByID)
	customers.Post("", middleware.ReqEmptyCheck, h.Create)
	customers.Patch("/:id", middleware.ReqEmptyCheck, h.Update)
	customers.Post("/:id/tier", middleware.ReqEmptyCheck, h.ChangeTier)
	customers.Get("/:id/tier-changes", h.ListTierChanges)
}

// Search GET /api/customers?q=
func (h *Handler) Search(c fiber.Ctx) error {
	customers, err := h.service.Search(c.Context(), c.Query("q"))
	if err != nil {
		return _package.SendError(c, err, "customer")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Customers retrieved successfully", customers))
}

// GetByID GET /api/customers/:id
func (h *Handler) GetByID(c fiber.Ctx) error {
	customer, err := h.service.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return _package.SendError(c, err, "customer")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Customer retrieved successfully", customer))
}

// Create POST /api/customers
func (h *Handler) Create(c fiber.Ctx) error {
	var req CreateCustomer
	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	customer, err := h.service.Create(c.Context(), req)
	if err != nil {
		return _package.SendError(c, err, "customer")
	}
	return _package.SendJSON(c, fiber.StatusCreated,
		_package.SuccessResponse("Customer created successfully", customer))
}

// Update PATCH /api/customers/:id
func (h *Handler) Update(c fiber.Ctx) error {
	var req UpdateCustomer
	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	customer, err := h.service.Update(c.Context(), c.Params("id"), req)
	if err != nil {
		return _package.SendError(c, err, "customer")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Customer updated successfully", customer))
}

// ChangeTier POST /api/customers/:id/tier
func (h *Handler) ChangeTier(c fiber.Ctx) error {
	// ต้องรู้ว่าใครเปลี่ยน (AuthMiddleware ตอบ 401 ไปแล้ว ตรงนี้กันพลาดอีกชั้น)
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c, errLoginRequired, "customer")
	}

	var req ChangeTier
	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	customer, err := h.service.ChangeTier(c.Context(), c.Params("id"), user.ID, req)
	if err != nil {
		return _package.SendError(c, err, "customer")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Customer tier changed successfully", customer))
}

// ListTierChanges GET /api/customers/:id/tier-changes
func (h *Handler) ListTierChanges(c fiber.Ctx) error {
	changes, err := h.service.ListTierChanges(c.Context(), c.Params("id"))
	if err != nil {
		return _package.SendError(c, err, "customer")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Tier changes retrieved successfully", changes))
}
