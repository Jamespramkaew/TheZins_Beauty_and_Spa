package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Appointment
type Appointment struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoomID     string    `gorm:"type:uuid;not null" json:"room_id"`
	Room       *Room     `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	CustomerID string    `gorm:"type:uuid;not null;index" json:"customer_id"`
	Customer   *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	StartAt    time.Time `gorm:"not null;index" json:"start_at"`
	EndAt      time.Time `gorm:"not null;check:end_at > start_at" json:"end_at"`
	Status     string    `gorm:"type:varchar(10);not null;default:'booked';check:status IN ('booked','done','cancelled','no-show')" json:"status"`
	Note       string    `gorm:"type:text" json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Appointment) TableName() string { return "appointments" }

// AppointmentStaff
type AppointmentStaff struct {
	AppointmentID string       `gorm:"type:uuid;primaryKey" json:"appointment_id"`
	Appointment   *Appointment `gorm:"foreignKey:AppointmentID" json:"appointment,omitempty"`
	EmployeeID    string       `gorm:"type:uuid;primaryKey;index" json:"employee_id"`
	Employee      *Employee    `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (AppointmentStaff) TableName() string { return "appointment_staff" }

// ServiceUsage
type ServiceUsage struct {
	ID           string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderItemID  string              `gorm:"type:uuid;not null;index" json:"order_item_id"` // คอร์สที่ซื้อ หรือรายการเครดิต VIP
	OrderItem    *OrderItem          `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
	CourseID     *string             `gorm:"type:uuid" json:"course_id"` // ใช้เฉพาะ VIP: ใช้บริการอะไร
	Course       *Course             `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	PriceCharged decimal.NullDecimal `gorm:"type:numeric(12,2)" json:"price_charged"` // ใช้เฉพาะ VIP: ราคาที่ตัดจากวงเงิน
	BranchID     string              `gorm:"type:uuid;not null" json:"branch_id"`     // สาขาที่ใช้บริการ
	Branch       *Branch             `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	UsedAt       time.Time           `gorm:"not null;index;autoCreateTime" json:"used_at"`
	Note         string              `gorm:"type:text" json:"note"`
}

func (ServiceUsage) TableName() string { return "service_usages" }

// UsageStaff
type UsageStaff struct {
	UsageID    string            `gorm:"type:uuid;primaryKey" json:"usage_id"`
	Usage      *ServiceUsage     `gorm:"foreignKey:UsageID" json:"usage,omitempty"`
	EmployeeID string            `gorm:"type:uuid;primaryKey;index" json:"employee_id"`
	Employee   *Employee         `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Commission decimal.Decimal   `gorm:"type:numeric(12,2);not null" json:"commission"`
	PayoutID   *string           `gorm:"type:uuid" json:"payout_id"` // null = ยังไม่จ่าย
	Payout     *CommissionPayout `gorm:"foreignKey:PayoutID" json:"payout,omitempty"`
}

func (UsageStaff) TableName() string { return "usage_staff" }

// StockLog = สมุดบัญชีสต็อก นับเฉพาะของที่ยังไม่เปิด
// ห้ามแก้หรือลบแถว (บังคับในโค้ด API) ถ้าผิดให้ทำรายการ adjust
// สต็อกคงเหลือ = SUM(qty_change) แยกตาม branch_id, stock_id
// index idx_stock_logs_balance (branch_id, stock_id) INCLUDE (qty_change) อยู่ใน extra.sql
// log_type:
//
//	receive    รับของเข้า (+)
//	sale       ขาย (-) มี order_item_id
//	use        ใช้ในบริการ ของที่ uses_per_unit = 1 (-) มี usage_id
//	open       เปิดของที่ uses_per_unit > 1 (-1) คู่กับ stock_containers
//	damaged    ของเสีย เฉพาะของที่ยังไม่เปิด (-)
//	expired    หมดอายุ เฉพาะของที่ยังไม่เปิด (-)
//	internal   ใช้ภายในร้าน ไม่ใช่กับลูกค้า (-)
//	transfer   โอนระหว่างสาขา (+/-) บันทึก 2 แถว
//	adjust     แก้ข้อมูลที่คีย์ผิด รู้สาเหตุ (+/-)
//	count_diff นับปลายวันไม่ตรง ไม่รู้สาเหตุ (+/-)
type StockLog struct {
	ID          string        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StockID     string        `gorm:"type:uuid;not null" json:"stock_id"`
	Stock       *Stock        `gorm:"foreignKey:StockID" json:"stock,omitempty"`
	BranchID    string        `gorm:"type:uuid;not null" json:"branch_id"`
	Branch      *Branch       `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	QtyChange   int           `gorm:"not null" json:"qty_change"` // + เข้า / - ออก (จำนวนชิ้น)
	LogType     string        `gorm:"type:varchar(20);not null;check:log_type IN ('receive','sale','use','open','damaged','expired','internal','transfer','adjust','count_diff')" json:"log_type"`
	OrderItemID *string       `gorm:"type:uuid" json:"order_item_id"` // ออกเพราะขาย
	OrderItem   *OrderItem    `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
	UsageID     *string       `gorm:"type:uuid" json:"usage_id"` // ออกเพราะใช้ในบริการ
	Usage       *ServiceUsage `gorm:"foreignKey:UsageID" json:"usage,omitempty"`
	Note        *string       `gorm:"type:text" json:"note"` // เหตุผล/หมายเหตุ กรณี adjust หรือ count_diff (adjust และ count_diff ควรกรอกเสมอ) null = ไม่มี
	// พนักงานที่ทำรายการ (บัญชีสาขาใช้ร่วมกัน จึงระบุตัวคนด้วย employee ไม่ใช่ user)
	// ค่ามาจาก client จึงต้องตรวจว่าเป็นพนักงานที่ active ของสาขาเดียวกับแถวนี้
	EmployeeID string    `gorm:"type:uuid;not null" json:"employee_id"`
	Employee   *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func (StockLog) TableName() string { return "stock_logs" }

// StockContainer = ของใช้หลายครั้ง (uses_per_unit > 1) ที่ถูกเปิดใช้ 1 ชิ้น เช่น ครีม 1 กระปุก
// - เปิด = สร้างแถวนี้ + stock_logs (open, -1)
// - ไม่บันทึกอะไรตอนใช้แต่ละครั้ง นับจาก service_usages ในช่วง opened_at ถึง closed_at
// - ถ้าเปิดซ้อนกัน การใช้บริการนับให้กระปุกที่เปิดล่าสุดและยังไม่ปิด
// - ปิดเพราะ damaged/expired ใส่แค่ close_reason ห้ามสร้าง stock_logs ซ้ำ (ตัดไปแล้วตอนเปิด)
type StockContainer struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StockID     string     `gorm:"type:uuid;not null" json:"stock_id"`
	Stock       *Stock     `gorm:"foreignKey:StockID" json:"stock,omitempty"`
	BranchID    string     `gorm:"type:uuid;not null" json:"branch_id"`
	Branch      *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	OpenedAt    time.Time  `gorm:"not null;autoCreateTime" json:"opened_at"`
	ClosedAt    *time.Time `gorm:"index;check:(closed_at IS NULL) = (close_reason IS NULL) AND (closed_at IS NULL) = (uses_count IS NULL)" json:"closed_at"` // null = ยังเปิดอยู่ (ปิดแล้วต้องมี close_reason และ uses_count เสมอ)
	CloseReason *string    `gorm:"type:varchar(10);check:close_reason IN ('finished','damaged','expired')" json:"close_reason"`
	UsesCount   *int       `json:"uses_count"` // คำนวณเก็บไว้ครั้งเดียวตอนปิด
}

func (StockContainer) TableName() string { return "stock_containers" }
