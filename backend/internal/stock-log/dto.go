package stocklog

// GetAllStockQuantitiesQuery = ค่าจาก query string ของ GET /api/stock-logs/quantities
// bind ด้วย c.Bind().Query(&q) ส่วนการตรวจค่าและกำหนดค่าเริ่มต้นทำใน service
type GetAllStockQuantitiesQuery struct {
	BranchID string `query:"branchId"` // admin เท่านั้น ส่วนบัญชีสาขาถูกบังคับเป็นสาขาตัวเอง ว่าง (admin) = ทุกสาขา
	Search   string `query:"search"`   // ค้นชื่อสินค้า
	Category string `query:"category"` // premium | normal ว่าง = ทั้งหมด
	LowStock bool   `query:"lowStock"` // true = เฉพาะที่ยอดคงเหลือ <= min_qty
	Active   string `query:"active"`   // true | false | all ว่าง = true (all รวมสินค้าที่ถูก soft delete)
	Page     int    `query:"page"`     // 0 = 1
	PageSize int    `query:"pageSize"` // 0 = 20 สูงสุด 100
}

// StockQuantity = ยอดคงเหลือของสินค้า 1 ชนิดใน 1 สาขา (นับเฉพาะของที่ยังไม่เปิด)
// ยอด = SUM(stock_logs.qty_change) ถ้าไม่เคยมี log เลยจะเป็น 0
type StockQuantity struct {
	StockID     string `json:"stock_id"`
	StockName   string `json:"stock_name"`
	Category    string `json:"category"`
	UsesPerUnit int    `json:"uses_per_unit"`
	MinQty      int    `json:"min_qty"`
	IsActive    bool   `json:"is_active"`
	BranchID    string `json:"branch_id"`
	BranchName  string `json:"branch_name"`
	Quantity    int    `json:"quantity"`
	IsLowStock  bool   `json:"is_low_stock"` // quantity <= min_qty
}
