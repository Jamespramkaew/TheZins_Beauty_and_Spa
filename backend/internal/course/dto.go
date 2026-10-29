package course

import "github.com/shopspring/decimal"

// AddCourseRequest รับ body ของ POST /api/courses
type AddCourseRequest struct {
	Name        string          `json:"name" validate:"required,max=150"`
	Price       decimal.Decimal `json:"price" validate:"required,gte=0"`
	MaxSession  int             `json:"max_session" validate:"required,gte=1"`
	ValidDay    *int            `json:"valid_day" validate:"omitempty,gte=1"`
	ServiceComm decimal.Decimal `json:"service_comm" validate:"gte=0"`
	Category    string          `json:"category" validate:"max=50"`
}

// GetAllCourseQuery = query string ของ GET /api/courses (รับเป็นข้อความ แปลงใน service)
type GetAllCourseQuery struct {
	Page     string // ว่าง = 1
	PageSize string // ว่าง = 20 สูงสุด 100
}
