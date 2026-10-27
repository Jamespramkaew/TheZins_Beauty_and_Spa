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

type AddStockRequest struct {
	StockName   string          `json:"stock_name" validate:"required,max=150"`
	Price       decimal.Decimal `json:"price" validate:"required,gte=0"`
	IsSellable  bool            `json:"is_sellable"`
	Category    string          `json:"category" validate:"required,oneof=premium normal"`
	UsesPerUnit int             `json:"uses_per_unit" validate:"required,gte=1"`
	MinQty      int             `json:"min_qty" validate:"gte=0"`
}

type UpdateStockRequest struct {
	StockName   *string         `json:"stock_name"`
	Price       decimal.Decimal `json:"price"` // ไม่ใช้ pointer เพราะ decimal.Decimal unmarshal ไม่ได้
	PriceSet    bool            `json:"-"`     // flag บอกว่า price ถูกส่งมาหรือไม่
	IsSellable  *bool           `json:"is_sellable"`
	Category    *string         `json:"category"`
	UsesPerUnit *int            `json:"uses_per_unit"`
	MinQty      *int            `json:"min_qty"`
	IsActive    *bool           `json:"is_active"`
}
