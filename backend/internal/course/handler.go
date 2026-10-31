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
	h.registerRoutes(router, middleware.AuthMiddleware)
}

// registerRoutes แยก auth ออกมาให้ test ใส่ตัวตั้งผู้ใช้ปลอมแทน session ได้ โดยยังใช้ route ชุดเดียวกับของจริง
func (h *Handler) registerRoutes(router fiber.Router, auth fiber.Handler) {
	courses := router.Group("/courses", auth)
	courses.Get("", h.GetAllCourse)
	courses.Post("", middleware.RequireAdminMiddleware, middleware.ReqEmptyCheck, h.AddCourse)
	courses.Patch("/:id", middleware.RequireAdminMiddleware, middleware.ReqEmptyCheck, h.UpdateCourseById)
	courses.Delete("/:id", middleware.RequireAdminMiddleware, h.DeleteCourseById)
}

// GetAllCourse GET /api/courses?search=&page=&page_size=
func (h *Handler) GetAllCourse(c fiber.Ctx) error {
	q := GetAllCourseQuery{
		Search:   c.Query("search"),
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

// DeleteCourseById DELETE /api/courses/:id (admin เท่านั้น เช็คที่ route)
func (h *Handler) DeleteCourseById(c fiber.Ctx) error {
	if err := h.service.DeleteCourseById(c.Context(), c.Params("id")); err != nil {
		return _package.SendError(c, err, "course")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Course deleted successfully", nil))
}
