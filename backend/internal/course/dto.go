package course

import (
	"encoding/json"
	"time"

	"backend/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AddCourseRequest รับ body ของ POST /api/courses
type AddCourseRequest struct {
	Name        string          `json:"name" validate:"required,max=150"`
	Price       decimal.Decimal `json:"price" validate:"required,gte=0"`
	MaxSession  int             `json:"max_session" validate:"required,gte=1"`
	ValidDay    *int            `json:"valid_day" validate:"omitempty,gte=1"`
	ServiceComm decimal.Decimal `json:"service_comm" validate:"gte=0"`
	Category    string          `json:"category" validate:"max=50"`
}

// UpdateCourseRequest รับ body ของ PATCH /api/courses/:id
// pointer = nil คือไม่ได้ส่งมา ไม่ต้องแก้
type UpdateCourseRequest struct {
	Name        *string          `json:"name"`
	Price       *decimal.Decimal `json:"price"`
	MaxSession  *int             `json:"max_session"`
	// รับดิบไว้แปลงใน service เพื่อแยก 3 กรณี: ไม่ส่งมา (nil), ส่ง null (ล้างเป็นไม่หมดอายุ), ส่งตัวเลข
	ValidDay    json.RawMessage  `json:"valid_day"`
	ServiceComm *decimal.Decimal `json:"service_comm"`
	Category    *string          `json:"category"`
	IsActive    *bool            `json:"is_active"`
}

// GetAllCourseQuery = query string ของ GET /api/courses (รับเป็นข้อความ แปลงใน service)
type GetAllCourseQuery struct {
	Search   string // ค้นชื่อคอร์ส (ไม่สนตัวพิมพ์เล็กใหญ่) ว่าง = ทั้งหมด
	Page     string // ว่าง = 1
	PageSize string // ว่าง = 20 สูงสุด 100
}

// CourseResponse = คอร์ส 1 แถวใน GET /api/courses พร้อมของที่ใช้ต่อการทำ 1 ครั้ง
type CourseResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Price       decimal.Decimal      `json:"price"`
	MaxSession  int                  `json:"max_session"`
	ValidDay    *int                 `json:"valid_day"`
	ServiceComm decimal.Decimal      `json:"service_comm"`
	Category    string               `json:"category"`
	IsActive    bool                 `json:"is_active"`
	CreatedAt   time.Time            `json:"created_at"`
	DeletedAt   gorm.DeletedAt       `json:"deleted_at"`
	Items       []CourseItemResponse `json:"items"` // ไม่มีของที่ใช้ = [] (ไม่ใช่ null)
}

// CourseItemResponse = ของที่คอร์สใช้ 1 อย่าง (course_id ไม่ส่งซ้ำเพราะอยู่ในคอร์สแล้ว)
type CourseItemResponse struct {
	Stock  StockRef `json:"stock"`
	UseQty int      `json:"use_qty"`
}

type StockRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// toCourseResponse แปลง model เป็น response
// stock ที่ไม่ได้ preload (nil) จะได้ object ที่มีเฉพาะ id แทนที่จะ panic
func toCourseResponse(c models.Course) CourseResponse {
	r := CourseResponse{
		ID:          c.ID,
		Name:        c.Name,
		Price:       c.Price,
		MaxSession:  c.MaxSession,
		ValidDay:    c.ValidDay,
		ServiceComm: c.ServiceComm,
		Category:    c.Category,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt,
		DeletedAt:   c.DeletedAt,
		Items:       make([]CourseItemResponse, 0, len(c.Items)),
	}
	for _, it := range c.Items {
		item := CourseItemResponse{Stock: StockRef{ID: it.StockID}, UseQty: it.UseQty}
		if it.Stock != nil {
			item.Stock.Name = it.Stock.Name
		}
		r.Items = append(r.Items, item)
	}
	return r
}
