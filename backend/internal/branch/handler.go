package branch

import (
	"backend/internal/middleware"
	_package "backend/package"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	repository Repository
}

func NewHandler(repository Repository) *Handler {
	return &Handler{repository: repository}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	h.registerRoutes(router, middleware.AuthMiddleware)
}

// registerRoutes แยก auth ออกมาให้ test ใส่ตัวตั้งผู้ใช้ปลอมแทน session ได้
func (h *Handler) registerRoutes(router fiber.Router, auth fiber.Handler) {
	branches := router.Group("/branches", auth)
	branches.Get("", h.GetActiveBranches)
}

// GetActiveBranches GET /api/branches
// คืนสาขาที่เปิดใช้งานทั้งหมดให้ทุกบัญชี (มีแค่ id / code / name ไม่มีข้อมูลสต็อก)
// บัญชีสาขาต้องรู้สาขาอื่นเพื่อเลือกสาขาปลายทางตอนโอน ส่วนการดูข้อมูลสต็อกยังถูกล็อกที่ endpoint ของ stock-log
func (h *Handler) GetActiveBranches(c fiber.Ctx) error {
	branches, err := h.repository.GetActiveBranches(c.Context())
	if err != nil {
		return _package.SendError(c, err, "branch")
	}
	return _package.SendJSON(c, fiber.StatusOK,
		_package.SuccessResponse("Branches retrieved successfully", branches))
}
