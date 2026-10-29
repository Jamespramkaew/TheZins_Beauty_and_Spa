package order

import "github.com/shopspring/decimal"

// CreateOrder รับ body ของ POST /api/orders
// ไม่มี branch_id, order_no, unit_price เพราะ server เป็นคนกำหนดเอง
type CreateOrder struct {
	CustomerID string          `json:"customer_id"`
	OrderDate  *string         `json:"order_date"` // YYYY-MM-DD ไม่ส่ง = วันนี้ (เวลาไทย)
	Source     *string         `json:"source"`     // phone / walk-in / page ไม่ส่ง = ไม่ทราบ
	Items      []CreateItem    `json:"items"`      // ต้องมีอย่างน้อย 1 รายการ
	Payments   []CreatePayment `json:"payments"`   // ไม่ส่งก็ได้ = ยังไม่จ่าย
}

// CreateItem รายการ 1 บรรทัดในบิล
type CreateItem struct {
	ItemType string          `json:"item_type"` // course หรือ stock
	CourseID *string         `json:"course_id"` // ต้องมีเมื่อ item_type = course
	StockID  *string         `json:"stock_id"`  // ต้องมีเมื่อ item_type = stock
	Qty      *int            `json:"qty"`       // ไม่ส่ง = 1, คอร์สต้องเป็น 1
	Discount decimal.Decimal `json:"discount"`  // ส่วนลดของทั้งบรรทัด (บาท) ไม่ส่ง = 0
}

// CreatePayment การจ่ายเงินที่จ่ายพร้อมตอนเปิดบิล
type CreatePayment struct {
	Method    string          `json:"method"`    // cash / transfer / card
	Amount    decimal.Decimal `json:"amount"`    // ต้องมากกว่า 0
	Reference string          `json:"reference"` // เลขสลิปหรือเลขท้ายบัตร ไม่ส่งก็ได้
}

// ListOrdersQuery ค่าจาก query string ของ GET /api/orders (ไม่ส่งช่องไหน = ไม่กรองช่องนั้น)
type ListOrdersQuery struct {
	CustomerID string `query:"customer_id"`
	From       string `query:"from"`      // YYYY-MM-DD
	To         string `query:"to"`        // YYYY-MM-DD
	Status     string `query:"status"`    // paid / partial / unpaid
	Page       string `query:"page"`      // ว่าง = 1
	PageSize   string `query:"page_size"` // ว่าง = 20 สูงสุด 100
}
