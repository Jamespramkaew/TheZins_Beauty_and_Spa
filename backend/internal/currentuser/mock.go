package currentuser

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// MockMiddleware ใช้แทน login ชั่วคราว(เปิดเฉพาะ ENVIRONMENT=development)
// เลือกผู้ใช้ด้วย header X-Mock-User: s1, s2, s3 หรือ admin (ไม่ส่งมา = s1)
// ถ้าหา username ไม่เจอ จะไม่ตั้งผู้ใช้ handler ที่ต้อง login จะตอบ 401 เอง
func MockMiddleware(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		username := c.Get("X-Mock-User", "s1")

		var user models.User
		err := db.WithContext(c.Context()).Where("username = ?", username).First(&user).Error
		if err == nil {
			Set(c, &CurrentUser{ID: user.ID, Role: user.Role, BranchID: user.BranchID})
		}
		return c.Next()
	}
}
