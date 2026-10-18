package _package

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

// Response is the main struct for all API responses
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    interface{} `json:"meta,omitempty"` // ใช้เฉพาะ endpoint ที่แบ่งหน้า
}

// PageMeta = ข้อมูลหน้าสำหรับตารางที่แบ่งหน้า
type PageMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewPageMeta คำนวณจำนวนหน้าจากจำนวนแถวทั้งหมด (total = 0 → total_pages = 0)
func NewPageMeta(page, pageSize int, total int64) PageMeta {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return PageMeta{Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages}
}

// SuccessPagedResponse สร้าง success response พร้อม meta ของหน้า
func SuccessPagedResponse(message string, data interface{}, meta PageMeta) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	}
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Type      string `json:"type"`
	Details   string `json:"details,omitempty"`
	Timestamp string `json:"timestamp"`
}

// SuccessResponse creates a success response format
func SuccessResponse(message string, data interface{}) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse creates an error response format
func ErrorResponse(message string, errorType string, details string) Response {
	errorData := ErrorDetail{
		Type:      errorType,
		Details:   details,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return Response{
		Success: false,
		Message: message,
		Data: []interface{}{
			map[string]interface{}{
				"error": errorData,
			},
		},
	}
}

// SendJSON sends a JSON response to the client
func SendJSON(c fiber.Ctx, statusCode int, response Response) error {
	return c.Status(statusCode).JSON(response)
}
