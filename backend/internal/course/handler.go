package course

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
	router.Get("/course", middleware.AuthMiddleware, h.GetAllCourse)
	router.Post("/course", middleware.AuthMiddleware, middleware.ReqEmptyCheck, h.AddCourse)
}

func (h *Handler) GetAllCourse(c fiber.Ctx) error {

	courses, err := h.service.GetAllCourse(c.Context())
	if err != nil {
		return _package.SendError(c, err, "course")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Courses retrieved successfully", courses))
}

func (h *Handler) AddCourse(c fiber.Ctx) error {

	userRole, _ := c.Locals("role").(string)
	if userRole != "admin" {
		return _package.SendJSON(c, fiber.StatusForbidden,
			_package.ErrorResponse("Access denied", _package.CodeForbidden, "Only admin can create course"))
	}

	var req AddCourseRequest

	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	course, err := h.service.AddCourse(c.Context(), req)
	if err != nil {
		return _package.SendError(c, err, "course")
	}

	return _package.SendJSON(c, fiber.StatusCreated,
		_package.SuccessResponse("Course created successfully", course))
}
