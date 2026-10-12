package example

import (
	"strconv"

	"backend/package"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetAllExampleMessage(c fiber.Ctx) error {

	limit := parseIntOrDefault(c.Query("limit"), 10)
	offset := parseIntOrDefault(c.Query("offset"), 0)
	page := parseIntOrDefault(c.Query("page"), 0)

	if page > 0 {
		offset = (page - 1) * limit
	}

	example, totalCount, err := h.service.GetAllExampleMessage(c.Context(), limit, offset)
	if err != nil {
		return _package.SendJSON(
			c,
			fiber.StatusInternalServerError,
			_package.ErrorResponse(
				"Failed to fetch example message",
				"DATABASE_ERROR",
				err.Error(),
			),
		)
	}

	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))
	currentPage := (offset / limit) + 1

	responseData := map[string]interface{}{
		"example": example,
		"pagination": map[string]interface{}{
			"total_records": totalCount,
			"total_pages":   totalPages,
			"current_page":  currentPage,
			"limit":         limit,
			"offset":        offset,
		},
	}

	return _package.SendJSON(
		c,
		fiber.StatusOK,
		_package.SuccessResponse(
			"Upload logs retrieved successfully",
			responseData,
		),
	)
}

func parseIntOrDefault(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Get("/example", h.GetAllExampleMessage)
}