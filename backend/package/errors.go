package _package

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v3"
)

// Code = ค่า type ใน error response ทั้งระบบใช้ชุดนี้ชุดเดียว (หน้าเว็บใช้เลือกข้อความที่จะแสดง)
const (
	CodeValidation   = "VALIDATION_ERROR" // 400 ข้อมูลที่ส่งมาผิด
	CodeUnauthorized = "UNAUTHORIZED"     // 401 ยังไม่ login หรือ login ไม่ผ่าน
	CodeForbidden    = "FORBIDDEN"        // 403 login แล้วแต่ไม่มีสิทธิ์
	CodeNotFound     = "NOT_FOUND"        // 404 ไม่พบข้อมูล หรือไม่มี route นี้
	CodeConflict     = "CONFLICT"         // 409 ข้อมูลซ้ำ
	CodeInternal     = "INTERNAL_ERROR"   // 500 server หรือ DB มีปัญหา
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
		ErrorResponse("Something went wrong", CodeInternal, ""))
}

// GetStatusFromCode แปลง error code เป็น HTTP status code
func GetStatusFromCode(code string) int {
	switch code {
	case CodeUnauthorized:
		return fiber.StatusUnauthorized // 401
	case CodeNotFound:
		return fiber.StatusNotFound // 404
	case CodeValidation:
		return fiber.StatusBadRequest // 400
	case CodeConflict:
		return fiber.StatusConflict // 409
	case CodeForbidden:
		return fiber.StatusForbidden // 403
	default:
		return fiber.StatusInternalServerError // 500
	}
}

// SendBadBody ตอบ 400 เมื่อ body ไม่ใช่ JSON หรือชนิดข้อมูลไม่ตรง
func SendBadBody(c fiber.Ctx, err error) error {
	return SendJSON(c, fiber.StatusBadRequest,
		ErrorResponse("Invalid request body", CodeValidation, err.Error()))
}
