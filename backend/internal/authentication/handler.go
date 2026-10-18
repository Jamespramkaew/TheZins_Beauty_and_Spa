package authentication

import (
	"errors"

	"backend/internal/middleware"
	_package "backend/package"

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
	auth := router.Group("/auth")
	auth.Post("/login", middleware.ReqEmptyCheck, h.Login)
	auth.Post("/logout", h.Logout)
	auth.Patch("/change-password", middleware.AuthMiddleware, middleware.ReqEmptyCheck, h.ChangePassword)
}

// Login POST /api/auth/login
func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest

	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	if req.Username == "" || req.Password == "" {
		return _package.SendError(c,
			_package.NewAppError("Invalid request", _package.CodeValidation, "username and password are required"),
			"authentication",
		)
	}

	user, err := h.service.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		return _package.SendError(c, err, "authentication")
	}

	sess := session.FromContext(c)
	if sess == nil {
		return _package.SendError(c, errors.New("session not found in context"), "authentication")
	}

	// เก็บข้อมูลใน session
	sess.Set("user_id", user.ID)
	sess.Set("username", user.Username)
	sess.Set("role", user.Role)

	if user.BranchID != nil {
		sess.Set("branchId", *user.BranchID)
	}

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

	if err := sess.Destroy(); err != nil {
		return _package.SendError(c, err, "authentication")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "the_zins_session",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Lax",
	})

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Logout successfully", nil))
}

// Change Password PATCH /api/auth/change-password
func (h *Handler) ChangePassword(c fiber.Ctx) error {

	var req ChangePasswordRequest

	if err := c.Bind().Body(&req); err != nil {
		return _package.SendBadBody(c, err)
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		return _package.SendError(c,
			_package.NewAppError("Invalid request", _package.CodeValidation, "currentPassword and newPassword are required"),
			"authentication",
		)
	}

	username := c.Locals("username").(string)

	if err := h.service.ChangePassword(c.Context(), username, req.CurrentPassword, req.NewPassword); err != nil {
		return _package.SendError(c, err, "authentication")
	}

	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Password changed successfully", nil))

}
