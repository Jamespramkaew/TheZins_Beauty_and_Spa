// Package currentuser เก็บและอ่านผู้ใช้ที่ login อยู่ของแต่ละ request ผ่าน c.Locals
// middleware.AuthMiddleware เป็นคนเรียก Set (สร้างจาก session หลัง login)
// handler ที่ต้องรู้ว่าใครเรียก (เช่น orders) เรียก FromCtx
package currentuser

import "github.com/gofiber/fiber/v3"

type CurrentUser struct {
	ID       string
	Role     string  // "admin" หรือ "branch"
	BranchID *string // nil = admin (ไม่ผูกสาขา)
}

// localsKey ชื่อช่องใน c.Locals ที่เก็บผู้ใช้ปัจจุบัน
const localsKey = "current_user"

// Set เก็บผู้ใช้ปัจจุบันไว้กับ request นี้
func Set(c fiber.Ctx, user *CurrentUser) {
	c.Locals(localsKey, user)
}

// FromCtx อ่านผู้ใช้ปัจจุบัน ok = false ถ้ายังไม่มีใคร login
func FromCtx(c fiber.Ctx) (*CurrentUser, bool) {
	user, ok := c.Locals(localsKey).(*CurrentUser)
	return user, ok && user != nil
}
