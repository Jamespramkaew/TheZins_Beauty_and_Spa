package middleware

import (
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

	currentBranch := currentSession.Get("branchId")
	c.Locals("branchId", currentBranch)

	return c.Next()
}
