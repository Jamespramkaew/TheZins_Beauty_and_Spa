package middleware

import (
	"encoding/json"
	_package "server/package"

	"github.com/gofiber/fiber/v3"
)


func ReqEmptyCheck(c fiber.Ctx) error {
	
	body := c.Body()

	// 1. ตรวจสอบว่า body ว่างไหม
	if len(body) == 0 {
		return _package.SendJSON(
			c,
			fiber.StatusBadRequest,
			_package.ErrorResponse(
				"Request body is required",
				"VALIDATION_ERROR",
				"Request body cannot be empty",
			),
		)
	}


	var jsonData map[string]interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return _package.SendJSON(
			c,
			fiber.StatusBadRequest,
			_package.ErrorResponse(
				"Invalid JSON format",
				"VALIDATION_ERROR",
				"Request body must be valid JSON object",
			),
		)
	}

	
	if len(jsonData) == 0 {
		return _package.SendJSON(
			c,
			fiber.StatusBadRequest,
			_package.ErrorResponse(
				"Request body is empty",
				"VALIDATION_ERROR",
				"Request body must contain at least one field",
			),
		)
	}

	
	return c.Next()
}
