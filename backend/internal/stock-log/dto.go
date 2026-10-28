package stocklog

import (
	"time"

	"backend/internal/models"
)

// GetAllStockQuantitiesQuery = ค่าจาก query string ของ GET /api/stock-logs/quantities
type GetAllStockQuantitiesQuery struct {
	BranchID string `query:"branch_id"` // admin เท่านั้น ส่วนบัญชีสาขาถูกบังคับเป็นสาขาตัวเอง ว่าง (admin) = ทุกสาขา
	Search   string `query:"search"`    // ค้นชื่อสินค้า
	Category string `query:"category"`  // premium | normal ว่าง = ทั้งหมด
	LowStock bool   `query:"low_stock"` // true = เฉพาะที่ยอดคงเหลือ <= min_qty
	Active   string `query:"active"`    // true | false | all ว่าง = true (all รวมสินค้าที่ถูก soft delete)
	Page     int    `query:"page"`      // 0 = 1
	PageSize int    `query:"page_size"` // 0 = 20 สูงสุด 100
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

// GetAllStockLogsQuery = ค่าจาก query string ของ GET /api/stock-logs
type GetAllStockLogsQuery struct {
	StockID     string `query:"stock_id"`      // สินค้าชนิดเดียว
	BranchID    string `query:"branch_id"`     // admin เท่านั้น ส่วนบัญชีสาขาถูกบังคับเป็นสาขาตัวเอง ว่าง (admin) = ทุกสาขา
	LogType     string `query:"log_type"`      // receive | sale | use | open | damaged | expired | internal | transfer | adjust | count_diff
	Direction   string `query:"direction"`     // in (qty_change > 0) | out (qty_change < 0) ว่าง = ทั้งหมด
	From        string `query:"from"`          // YYYY-MM-DD (เวลาไทย) นับตั้งแต่ 00:00 ของวันนั้น
	To          string `query:"to"`            // YYYY-MM-DD (เวลาไทย) นับถึงสิ้นวันนั้น
	OrderItemID string `query:"order_item_id"` // log ที่เกิดจากรายการขายนี้
	UsageID     string `query:"usage_id"`      // log ที่เกิดจากการใช้บริการนี้
	Page        int    `query:"page"`          // 0 = 1
	PageSize    int    `query:"page_size"`     // 0 = 20 สูงสุด 100
}

// StockLogResponse = แถวประวัติสต็อก 1 แถวใน GET /api/stock-logs (เฉพาะฟิลด์ที่หน้าจอต้องใช้)
// id ของสินค้า/สาขา/พนักงานอยู่ใน object ข้างในแล้ว จึงไม่ซ้ำที่ระดับบน
type StockLogResponse struct {
	ID          string      `json:"id"`
	LogType     string      `json:"log_type"`
	QtyChange   int         `json:"qty_change"`
	Note        *string     `json:"note"`
	OrderItemID *string     `json:"order_item_id"` // ไม่ null = เกิดจากการขาย
	UsageID     *string     `json:"usage_id"`      // ไม่ null = เกิดจากการใช้บริการ
	CreatedAt   time.Time   `json:"created_at"`
	Stock       StockRef    `json:"stock"`
	Branch      BranchRef   `json:"branch"`
	Employee    EmployeeRef `json:"employee"`
}

type StockRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type BranchRef struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type EmployeeRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// toStockLogResponse แปลง model เป็น response
// ความสัมพันธ์ที่ไม่ได้ preload (nil) จะได้ object ที่มีเฉพาะ id แทนที่จะ panic
func toStockLogResponse(l models.StockLog) StockLogResponse {
	r := StockLogResponse{
		ID:          l.ID,
		LogType:     l.LogType,
		QtyChange:   l.QtyChange,
		Note:        l.Note,
		OrderItemID: l.OrderItemID,
		UsageID:     l.UsageID,
		CreatedAt:   l.CreatedAt,
		Stock:       StockRef{ID: l.StockID},
		Branch:      BranchRef{ID: l.BranchID},
		Employee:    EmployeeRef{ID: l.EmployeeID},
	}
	if l.Stock != nil {
		r.Stock.Name = l.Stock.Name
	}
	if l.Branch != nil {
		r.Branch.Code = l.Branch.Code
		r.Branch.Name = l.Branch.Name
	}
	if l.Employee != nil {
		r.Employee.Name = l.Employee.Name
		r.Employee.Role = l.Employee.Role
	}
	return r
}

// AddStockLogMovementRequest = request body ของ POST /api/stock-logs
type AddStockLogMovementRequest struct {
	Type       string `json:"type"`        // receive | expired | damaged | internal | adjust
	BranchID   string `json:"branch_id"`   // admin ต้องส่ง ส่วนบัญชีสาขาถูกบังคับเป็นสาขาใน session
	StockID    string `json:"stock_id"`    // ต้องมีอยู่จริงและไม่ถูกลบ
	EmployeeID string `json:"employee_id"` // พนักงาน active ของสาขาเดียวกับรายการ
	// receive/expired/damaged/internal: ต้องเป็นบวก server ใส่เครื่องหมายให้
	// adjust: ส่งพร้อมเครื่องหมาย (+ เพิ่ม / - ลด) ห้ามเป็น 0
	Qty  int    `json:"qty"`
	Note string `json:"note"` // adjust บังคับ ชนิดอื่นไม่บังคับ (ว่าง = NULL)
}

// TransferStockLog = request body ของ POST /api/stock-logs/transfer
type TransferStockLogRequest struct {
	StockID      string `json:"stock_id"`
	Type         string `json:"type"`
	FromBranchId string `json:"from_branch_id"`
	ToBranchId   string `json:"to_branch_id"`
	EmployeeID   string `json:"employee_id"`
	Qty          int    `json:"qty"`
	Note         string `json:"-"` // client ไม่ต้องส่ง service สร้างให้เองว่าย้ายจากไหนไปไหน
}
