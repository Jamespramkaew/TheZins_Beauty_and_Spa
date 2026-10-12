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
