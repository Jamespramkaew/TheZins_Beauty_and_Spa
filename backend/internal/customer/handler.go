package customer

import (
	"backend/internal/middleware"
	"backend/package"

	"github.com/gofiber/fiber/v3"
)

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
	customers.Post("", h.Create)
	customers.Patch("/:id", h.Update)
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
