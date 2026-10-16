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
}

// Create POST /api/orders
func (h *Handler) Create(c fiber.Ctx) error {
	// ต้อง login ก่อน (ตอนนี้ผู้ใช้มาจาก mock middleware ภายหลังมาจาก JWT)
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendJSON(c, fiber.StatusUnauthorized,
			_package.ErrorResponse("Login required", "UNAUTHORIZED", ""))
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

// sendError แปลง error จาก service เป็น HTTP status
func sendError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return _package.SendJSON(c, fiber.StatusBadRequest,
			_package.ErrorResponse("Invalid order data", "VALIDATION_ERROR", err.Error()))
	case errors.Is(err, ErrForbidden):
		return _package.SendJSON(c, fiber.StatusForbidden,
			_package.ErrorResponse("Not allowed to create orders", "FORBIDDEN", err.Error()))
	default:
		// error อื่นมาจากฐานข้อมูล เขียนลง log ฝั่ง server แต่ไม่ส่งรายละเอียดให้ client
		log.Printf("order: %v", err)
		return _package.SendJSON(c, fiber.StatusInternalServerError,
			_package.ErrorResponse("Something went wrong", "DATABASE_ERROR", ""))
	}
}
