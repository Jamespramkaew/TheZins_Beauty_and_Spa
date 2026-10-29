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
	courses := router.Group("/courses", middleware.AuthMiddleware)
	courses.Get("", h.GetAllCourse)
	courses.Post("", middleware.RequireAdminMiddleware, middleware.ReqEmptyCheck, h.AddCourse)
	courses.Patch("/:id", middleware.RequireAdminMiddleware, middleware.ReqEmptyCheck, h.UpdateCourseById)
}

// GetAllCourse GET /api/courses?page=&page_size=
func (h *Handler) GetAllCourse(c fiber.Ctx) error {
	q := GetAllCourseQuery{
		Page:     c.Query("page"),
		PageSize: c.Query("page_size"),
	}

	courses, meta, err := h.service.GetAllCourse(c.Context(), q)
	if err != nil {
		return _package.SendError(c, err, "course")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessPagedResponse("Courses retrieved successfully", courses, meta))
}

// AddCourse POST /api/courses (admin เท่านั้น เช็คที่ route)
func (h *Handler) AddCourse(c fiber.Ctx) error {
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

// UpdateCourseById PATCH /api/courses/:id (admin เท่านั้น เช็คที่ route)
func (h *Handler) UpdateCourseById(c fiber.Ctx) error {
	var req UpdateCourseRequest
	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	course, err := h.service.UpdateCourseById(c.Context(), c.Params("id"), req)
	if err != nil {
		return _package.SendError(c, err, "course")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Course updated successfully", course))
}
