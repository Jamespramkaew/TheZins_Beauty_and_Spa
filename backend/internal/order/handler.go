package order

import (
	"backend/internal/currentuser"
	"backend/internal/middleware"
	"backend/package"

	"github.com/gofiber/fiber/v3"
)

// errLoginRequired ตอบ 401 เมื่อไม่มีผู้ใช้ (ยังไม่ login)
var errLoginRequired = _package.NewAppError("Login required", _package.CodeUnauthorized, "")

// errInvalidQuery ตอบ 400 เมื่ออ่าน query ไม่ได้ (ไม่ส่งข้อความ error ดิบของ library ออกไป)
var errInvalidQuery = _package.NewAppError("Invalid query", _package.CodeValidation, "query string is invalid")

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	orders := router.Group("/orders", middleware.AuthMiddleware)
	orders.Post("", middleware.ReqEmptyCheck, h.Create)
	orders.Get("", h.List)
	orders.Get("/:id", h.GetByID)
	orders.Post("/:id/payments", middleware.ReqEmptyCheck, h.AddPayment)
}

// Create POST /api/orders
func (h *Handler) Create(c fiber.Ctx) error {
	// ต้อง login ก่อน (AuthMiddleware ตอบ 401 ไปแล้ว ตรงนี้กันพลาดอีกชั้น)
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c, errLoginRequired, "order")
	}

	var req CreateOrder
	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	order, err := h.service.Create(c.Context(), user, req)
	if err != nil {
		return _package.SendError(c, err, "order")
	}
	return _package.SendJSON(c, fiber.StatusCreated,
		_package.SuccessResponse("Order created successfully", order))
}

// List GET /api/orders?customer_id=&from=&to=&status=&page=&page_size=
func (h *Handler) List(c fiber.Ctx) error {
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c, errLoginRequired, "order")
	}

	var q ListOrdersQuery
	if err := c.Bind().Query(&q); err != nil {
		return _package.SendError(c, errInvalidQuery, "order")
	}

	orders, meta, err := h.service.List(c.Context(), user, q)
	if err != nil {
		return _package.SendError(c, err, "order")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessPagedResponse("Orders retrieved successfully", orders, meta))
}

// GetByID GET /api/orders/:id
func (h *Handler) GetByID(c fiber.Ctx) error {
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c, errLoginRequired, "order")
	}

	order, err := h.service.GetByID(c.Context(), user, c.Params("id"))
	if err != nil {
		return _package.SendError(c, err, "order")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Order retrieved successfully", order))
}

// AddPayment POST /api/orders/:id/payments
func (h *Handler) AddPayment(c fiber.Ctx) error {
	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c, errLoginRequired, "order")
	}

	var req CreatePayment
	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	order, err := h.service.AddPayment(c.Context(), user, c.Params("id"), req)
	if err != nil {
		return _package.SendError(c, err, "order")
	}
	return _package.SendJSON(c, fiber.StatusCreated,
		_package.SuccessResponse("Payment recorded successfully", order))
}
