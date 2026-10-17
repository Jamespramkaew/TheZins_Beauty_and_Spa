package order

import (
	"errors"
	"log"

	"backend/internal/currentuser"
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
	router.Post("/orders", h.Create)
	router.Get("/orders", h.List)
	router.Get("/orders/:id", h.GetByID)
	router.Post("/orders/:id/payments", h.AddPayment)
}

// Create POST /api/orders
func (h *Handler) Create(c fiber.Ctx) error {
	// ต้อง login ก่อน (ตอนนี้ผู้ใช้มาจาก mock middleware ภายหลังมาจาก JWT)
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return sendUnauthorized(c)
	}

	var req CreateOrder
	if err := c.Bind().Body(&req); err != nil {
		return _package.SendJSON(c, fiber.StatusBadRequest,
			_package.ErrorResponse("Invalid request body", "VALIDATION_ERROR", err.Error()))
	}

	order, err := h.service.Create(c.Context(), user, req)
	if err != nil {
		return sendError(c, err)
	}
	return _package.SendJSON(c, fiber.StatusCreated,
		_package.SuccessResponse("Order created successfully", order))
}

// List GET /api/orders?customer_id=&from=&to=&status=
func (h *Handler) List(c fiber.Ctx) error {
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return sendUnauthorized(c)
	}

	var q ListOrdersQuery
	if err := c.Bind().Query(&q); err != nil {
		return _package.SendJSON(c, fiber.StatusBadRequest,
			_package.ErrorResponse("Invalid query", "VALIDATION_ERROR", err.Error()))
	}

	orders, err := h.service.List(c.Context(), user, q)
	if err != nil {
		return sendError(c, err)
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Orders retrieved successfully", orders))
}

// GetByID GET /api/orders/:id
func (h *Handler) GetByID(c fiber.Ctx) error {
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return sendUnauthorized(c)
	}

	order, err := h.service.GetByID(c.Context(), user, c.Params("id"))
	if err != nil {
		return sendError(c, err)
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Order retrieved successfully", order))
}

// AddPayment POST /api/orders/:id/payments
func (h *Handler) AddPayment(c fiber.Ctx) error {
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return sendUnauthorized(c)
	}

	var req CreatePayment
	if err := c.Bind().Body(&req); err != nil {
		return _package.SendJSON(c, fiber.StatusBadRequest,
			_package.ErrorResponse("Invalid request body", "VALIDATION_ERROR", err.Error()))
	}

	order, err := h.service.AddPayment(c.Context(), user, c.Params("id"), req)
	if err != nil {
		return sendError(c, err)
	}
	return _package.SendJSON(c, fiber.StatusCreated,
		_package.SuccessResponse("Payment recorded successfully", order))
}

// sendUnauthorized ตอบ 401 เมื่อไม่มีผู้ใช้ (ยังไม่ login)
func sendUnauthorized(c fiber.Ctx) error {
	return _package.SendJSON(c, fiber.StatusUnauthorized,
		_package.ErrorResponse("Login required", "UNAUTHORIZED", ""))
}

// sendError แปลง error จาก service เป็น HTTP status
func sendError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return _package.SendJSON(c, fiber.StatusBadRequest,
			_package.ErrorResponse("Invalid order data", "VALIDATION_ERROR", err.Error()))
	case errors.Is(err, ErrForbidden):
		return _package.SendJSON(c, fiber.StatusForbidden,
			_package.ErrorResponse("Not allowed", "FORBIDDEN", err.Error()))
	case errors.Is(err, ErrNotFound):
		return _package.SendJSON(c, fiber.StatusNotFound,
			_package.ErrorResponse("Order not found", "NOT_FOUND", ""))
	default:
		// error อื่นมาจากฐานข้อมูล เขียนลง log ฝั่ง server แต่ไม่ส่งรายละเอียดให้ client
		log.Printf("order: %v", err)
		return _package.SendJSON(c, fiber.StatusInternalServerError,
			_package.ErrorResponse("Something went wrong", "DATABASE_ERROR", ""))
	}
}
