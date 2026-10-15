package customer

import (
	"errors"
	"log"

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
	router.Get("/customers", h.Search)
	router.Get("/customers/:id", h.GetByID)
	router.Post("/customers", h.Create)
	router.Patch("/customers/:id", h.Update)
}

// Search GET /api/customers?q=
func (h *Handler) Search(c fiber.Ctx) error {
	customers, err := h.service.Search(c.Context(), c.Query("q"))
	if err != nil {
		return sendError(c, err)
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Customers retrieved successfully", customers))
}

// GetByID GET /api/customers/:id
func (h *Handler) GetByID(c fiber.Ctx) error {
	customer, err := h.service.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return sendError(c, err)
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Customer retrieved successfully", customer))
}

// Create POST /api/customers
func (h *Handler) Create(c fiber.Ctx) error {
	var req CreateCustomer
	if err := c.Bind().Body(&req); err != nil {
		return sendBadBody(c, err)
	}

	customer, err := h.service.Create(c.Context(), req)
	if err != nil {
		return sendError(c, err)
	}
	return _package.SendJSON(c, fiber.StatusCreated,
		_package.SuccessResponse("Customer created successfully", customer))
}

// Update PATCH /api/customers/:id
func (h *Handler) Update(c fiber.Ctx) error {
	var req UpdateCustomer
	if err := c.Bind().Body(&req); err != nil {
		return sendBadBody(c, err)
	}

	customer, err := h.service.Update(c.Context(), c.Params("id"), req)
	if err != nil {
		return sendError(c, err)
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Customer updated successfully", customer))
}

// sendBadBody ตอบ 400 เมื่อ body ไม่ใช่ JSON หรือชนิดข้อมูลไม่ตรง เช่น "hn_no": 123
func sendBadBody(c fiber.Ctx, err error) error {
	return _package.SendJSON(c, fiber.StatusBadRequest,
		_package.ErrorResponse("Invalid request body", "VALIDATION_ERROR", err.Error()))
}

// sendError แปลง error จาก service เป็น HTTP status
func sendError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return _package.SendJSON(c, fiber.StatusBadRequest,
			_package.ErrorResponse("Invalid customer data", "VALIDATION_ERROR", err.Error()))
	case errors.Is(err, ErrNotFound):
		return _package.SendJSON(c, fiber.StatusNotFound,
			_package.ErrorResponse("Customer not found", "NOT_FOUND", ""))
	case errors.Is(err, ErrDuplicateHN):
		return _package.SendJSON(c, fiber.StatusConflict,
			_package.ErrorResponse("HN number already exists", "CONFLICT", err.Error()))
	default:
		// error อื่นมาจากฐานข้อมูล เขียนลง log ฝั่ง server แต่ไม่ส่งรายละเอียดให้ client
		log.Printf("customer: %v", err)
		return _package.SendJSON(c, fiber.StatusInternalServerError,
			_package.ErrorResponse("Something went wrong", "DATABASE_ERROR", ""))
	}
}
