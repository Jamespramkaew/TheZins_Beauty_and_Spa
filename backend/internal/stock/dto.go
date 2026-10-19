package stock

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
