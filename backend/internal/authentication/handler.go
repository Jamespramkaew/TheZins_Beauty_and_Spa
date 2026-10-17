package authentication

import (
	"errors"

	_package "backend/package"
	"backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Post("/login", middleware.ReqEmptyCheck, h.Login)
	router.Post("/logout", h.Logout)
}

// Login POST /api/auth/login
func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest

	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	// เรียก service ทำ business logic
	user, err := h.service.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return _package.SendError(c, err, "authentication")
	}

	// ดึง session middleware จาก context (ไม่มี error, session จะถูกบันทึกอัตโนมัติท้าย request)
	sess := session.FromContext(c)
	if sess == nil {
		return _package.SendError(c, errors.New("session not found in context"), "authentication")
	}

	// เก็บข้อมูลใน session
	sess.Set("user_id", user.ID)
	sess.Set("username", user.Username)
	sess.Set("role", user.Role)

	// ส่ง response สำเร็จ
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Login successfully", fiber.Map{
			"user_id":  user.ID,
			"username": user.Username,
			"role":     user.Role,
		}))
}

// Logout POST /api/auth/logout
func (h *Handler) Logout(c fiber.Ctx) error {
	sess := session.FromContext(c)
	if sess == nil {
		return _package.SendError(c, errors.New("session not found in context"), "authentication")
	}

	// ลบ session
	if err := sess.Destroy(); err != nil {
		return _package.SendError(c, err, "authentication")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Logout successfully", nil))
}
