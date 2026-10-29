package middleware

import (
	"backend/internal/currentuser"
	_package "backend/package"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func AuthMiddleware(c fiber.Ctx) error {

	currentSession := session.FromContext(c)

	if currentSession == nil {
		return _package.SendError(c,
			_package.NewAppError("Unauthorized", _package.CodeUnauthorized, "You must login first"),
			"middleware",
		)
	}

	if currentSession.Get("user_id") == nil {
		return _package.SendError(c,
			_package.NewAppError("Unauthorized", _package.CodeUnauthorized, "You must login first"),
			"middleware",
		)
	}

	c.Locals("userId", currentSession.Get("user_id"))
	c.Locals("username", currentSession.Get("username"))
	c.Locals("role", currentSession.Get("role"))
	c.Locals("branchId", currentSession.Get("branchId"))
	c.Locals("role", currentSession.Get("role"))

	// สร้าง CurrentUser จาก session ให้ handler ที่ใช้ currentuser.FromCtx (order) อ่านได้
	user := &currentuser.CurrentUser{}
	user.ID, _ = currentSession.Get("user_id").(string)
	user.Role, _ = currentSession.Get("role").(string)
	if branchID, ok := currentSession.Get("branchId").(string); ok {
		user.BranchID = &branchID // admin ไม่มี branchId → ปล่อยเป็น nil
	}
	currentuser.Set(c, user)

	return c.Next()
}

// RequireAdminMiddleware ตรวจสอบว่า user ที่ login เป็น admin หรือไม่
// ใช้หลังจาก AuthMiddleware เท่านั้น
func RequireAdminMiddleware(c fiber.Ctx) error {

	user, ok := currentuser.FromCtx(c)
	if !ok {
		return _package.SendError(c,
			_package.NewAppError("Forbidden", _package.CodeForbidden, "Unable to verify user role"),
			"middleware",
		)
	}

	if user.Role != "admin" {
		return _package.SendError(c,
			_package.NewAppError("Forbidden", _package.CodeForbidden, "Admin access required"),
			"middleware",
		)
	}

	return c.Next()
}
