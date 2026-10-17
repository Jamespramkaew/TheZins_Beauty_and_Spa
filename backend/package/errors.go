package _package

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

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

// FiberErrorHandler ใช้กับ fiber.Config.ErrorHandler
// error ที่ไม่ได้ผ่าน SendError (เช่น body ใหญ่เกิน 413, panic) ตอบเป็น ErrorResponse เหมือนกันทั้งระบบ
func FiberErrorHandler(c fiber.Ctx, err error) error {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return SendJSON(c, fe.Code, ErrorResponse(fe.Message, codeFromStatus(fe.Code), ""))
	}
	log.Printf("[fiber] %v", err)
	return SendJSON(c, fiber.StatusInternalServerError,
		ErrorResponse("Something went wrong", CodeInternal, ""))
}

// codeFromStatus แปลง HTTP status กลับเป็น Code (ใช้กับ error ของ Fiber ที่มีแค่ status)
func codeFromStatus(status int) string {
	switch status {
	case fiber.StatusUnauthorized:
		return CodeUnauthorized
	case fiber.StatusForbidden:
		return CodeForbidden
	case fiber.StatusNotFound:
		return CodeNotFound
	case fiber.StatusConflict:
		return CodeConflict
	}
	if status >= 500 {
		return CodeInternal
	}
	return CodeValidation
}

// SendBadBody ตอบ 400 เมื่อ body ไม่ใช่ JSON หรือชนิดข้อมูลไม่ตรง
func SendBadBody(c fiber.Ctx, err error) error {
	return SendJSON(c, fiber.StatusBadRequest,
		ErrorResponse("Invalid request body", CodeValidation, bindErrorDetail(err)))
}

// bindErrorDetail แปลง error จาก c.Bind().Body เป็นข้อความที่คนอ่านเข้าใจ ไม่โผล่ชื่อ struct ของ Go
func bindErrorDetail(err error) string {
	// ชนิดข้อมูลผิด เช่น {"hn_no": 123} → "hn_no must be a string"
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return fmt.Sprintf("%s must be %s", typeErr.Field, jsonTypeName(typeErr.Type))
	}
	// JSON พัง เช่น {"hn_no":
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return "body must be valid JSON"
	}
	// ไม่ได้ส่ง Content-Type: application/json
	if errors.Is(err, fiber.ErrUnprocessableEntity) {
		return "Content-Type must be application/json"
	}
	// อื่นๆ เช่น ค่าเงินที่แปลงเป็นตัวเลขไม่ได้ ("abc")
	return "invalid value: " + strings.TrimPrefix(err.Error(), "bind from body: ")
}

// jsonTypeName ชื่อชนิดข้อมูลแบบ JSON ของ type ใน Go (ใช้ในข้อความ error)
func jsonTypeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "a string"
	case reflect.Bool:
		return "true or false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "a number"
	case reflect.Slice, reflect.Array:
		return "an array"
	case reflect.Map, reflect.Struct:
		return "an object"
	case reflect.Pointer:
		return jsonTypeName(t.Elem())
	default:
		return "a valid value"
	}
}
