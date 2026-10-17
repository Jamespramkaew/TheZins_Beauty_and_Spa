package _package

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v3"
)

type AppError struct {
	Message string // "Invalid credentials"
	Code    string // "UNAUTHORIZED"
	Detail  string // "Username or password is incorrect"
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(message, code, detail string) *AppError {
	return &AppError{
		Message: message,
		Code:    code,
		Detail:  detail,
	}
}

// SendError จัดการ error และส่ง HTTP response (ใช้ร่วมกันทั้งระบบ)
func SendError(c fiber.Ctx, err error, module string) error {
	// ถ้าเป็น AppError (errors.As หาเจอแม้ error จะถูกห่อด้วย fmt.Errorf("%w", ...))
	var appErr *AppError
	if errors.As(err, &appErr) {
		status := GetStatusFromCode(appErr.Code)
		return SendJSON(c, status,
			ErrorResponse(appErr.Message, appErr.Code, appErr.Detail))
	}

	// Error อื่นๆ (database error, unknown error)
	log.Printf("[%s] %v", module, err)
	return SendJSON(c, fiber.StatusInternalServerError,
		ErrorResponse("Something went wrong", "INTERNAL_ERROR", ""))
}

// GetStatusFromCode แปลง error code เป็น HTTP status code
func GetStatusFromCode(code string) int {
	switch code {
	case "UNAUTHORIZED":
		return fiber.StatusUnauthorized // 401
	case "NOT_FOUND":
		return fiber.StatusNotFound // 404
	case "VALIDATION_ERROR":
		return fiber.StatusBadRequest // 400
	case "CONFLICT":
		return fiber.StatusConflict // 409
	case "FORBIDDEN":
		return fiber.StatusForbidden // 403
	default:
		return fiber.StatusInternalServerError // 500
	}
}

// SendBadBody ตอบ 400 เมื่อ body ไม่ใช่ JSON หรือชนิดข้อมูลไม่ตรง
func SendBadBody(c fiber.Ctx, err error) error {
	return SendJSON(c, fiber.StatusBadRequest,
		ErrorResponse("Invalid request body", "VALIDATION_ERROR", err.Error()))
}
