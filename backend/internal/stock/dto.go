package stock

import "github.com/shopspring/decimal"

// สถานะของกระปุกที่ใช้กรองใน GET /api/stock/:id/containers?status=
const (
	StatusOpen   = "open"   // ยังเปิดอยู่ (closed_at IS NULL)
	StatusClosed = "closed" // ปิดแล้ว (closed_at IS NOT NULL)
	StatusAll    = "all"    // ทั้งหมด
)

// ContainerQuery = ค่าที่ client ส่งมาตอนขอรายการกระปุกของสินค้า 1 ชนิด
type ContainerQuery struct {
	StockID  string // จาก path :id
	BranchID string // จาก ?branchId= (admin เท่านั้นที่ใช้ได้ ส่วนบัญชีสาขาถูกบังคับเป็นสาขาตัวเอง)
	Status   string // จาก ?status= ว่าง = open
	Page     string // จาก ?page= ว่าง = 1
	PageSize string // จาก ?pageSize= ว่าง = 20 (สูงสุด 100)
}

// QuantityQuery = ค่าที่ client ส่งมาตอนขอยอดคงเหลือ GET /api/stock/quantities
type QuantityQuery struct {
	BranchID string // จาก ?branchId= (admin เท่านั้น ว่าง = รวมทุกสาขา)
	Search   string // จาก ?search= ค้นจากชื่อสินค้า (ไม่สนตัวพิมพ์เล็กใหญ่)
	Category string // จาก ?category= premium | normal ว่าง = ทุกหมวด
	LowStock string // จาก ?lowStock= true = เฉพาะที่ยอด <= min_qty
}

// QuantityFilter = ค่าที่ผ่านการตรวจจาก service แล้ว ส่งต่อให้ repository
type QuantityFilter struct {
	BranchID string
	Search   string
	Category string
	LowStock bool
}

// StockQuantity = สินค้า 1 ชนิดพร้อมจำนวนคงเหลือ (นับเฉพาะของที่ยังไม่เปิด)
type StockQuantity struct {
	StockID  string          `json:"stock_id"`
	Name     string          `json:"name"`
	Price    decimal.Decimal `json:"price"`
	MinQty   int             `json:"min_qty"`
	Quantity int             `json:"quantity"`
}
