package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Order
type Order struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderNo    string    `gorm:"type:varchar(30);not null;uniqueIndex" json:"order_no"` // เลขบิลที่ลูกค้าเห็น
	CustomerID string    `gorm:"type:uuid;not null;index" json:"customer_id"`
	Customer   *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	BranchID   string    `gorm:"type:uuid;not null;index:idx_orders_branch_date,priority:1" json:"branch_id"`
	Branch     *Branch   `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Source     *string   `gorm:"type:varchar(10);check:source IN ('phone','walk-in','page')" json:"source"`    // null = ไม่ทราบช่องทาง
	OrderDate  time.Time `gorm:"type:date;not null;index:idx_orders_branch_date,priority:2" json:"order_date"` // วันขายจริง อาจย้อนหลังได้
	CreatedAt  time.Time `json:"created_at"`
}

func (Order) TableName() string { return "orders" }

// OrderItem = รายการในบิล (คอร์ส / สินค้า / เครดิต VIP)
type OrderItem struct {
	ID           string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID      string              `gorm:"type:uuid;not null;index" json:"order_id"`
	Order        *Order              `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	ItemType     string              `gorm:"type:varchar(10);not null;check:item_type IN ('course','stock','credit')" json:"item_type"`
	CourseID     *string             `gorm:"type:uuid" json:"course_id"`
	Course       *Course             `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	StockID      *string             `gorm:"type:uuid" json:"stock_id"`
	Stock        *Stock              `gorm:"foreignKey:StockID" json:"stock,omitempty"`
	Qty          int                 `gorm:"not null;default:1;check:qty >= 1" json:"qty"`
	UnitPrice    decimal.Decimal     `gorm:"type:numeric(12,2);not null" json:"unit_price"` // ราคาที่ขายจริง (credit = เงินที่จ่ายจริง)
	Discount     decimal.Decimal     `gorm:"type:numeric(12,2);not null;default:0" json:"discount"`
	TotalSession *int                `json:"total_session"` // คัดลอกจาก max_session ปรับได้
	CreditType   *string             `gorm:"type:varchar(10);check:credit_type IN ('course','product')" json:"credit_type"`
	CreditAmount decimal.NullDecimal `gorm:"type:numeric(12,2)" json:"credit_amount"` // วงเงินที่ได้
	ExpiresAt    *time.Time          `gorm:"type:date" json:"expires_at"`
}

func (OrderItem) TableName() string { return "order_items" }

// OrderItemSeller
type OrderItemSeller struct {
	OrderItemID  string              `gorm:"type:uuid;primaryKey" json:"order_item_id"`
	OrderItem    *OrderItem          `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
	EmployeeID   string              `gorm:"type:uuid;primaryKey;index" json:"employee_id"`
	Employee     *Employee           `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	SharePercent decimal.NullDecimal `gorm:"type:numeric(5,2)" json:"share_percent"`        // null = หารเท่ากัน
	Commission   decimal.Decimal     `gorm:"type:numeric(12,2);not null" json:"commission"` // คำนวณเก็บไว้ตอนขาย
	PayoutID     *string             `gorm:"type:uuid" json:"payout_id"`                    // null = ยังไม่จ่าย
	Payout       *CommissionPayout   `gorm:"foreignKey:PayoutID" json:"payout,omitempty"`
}

func (OrderItemSeller) TableName() string { return "order_item_sellers" }

// Payment
type Payment struct {
	ID        string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID   string          `gorm:"type:uuid;not null;index" json:"order_id"`
	Order     *Order          `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	PaidAt    time.Time       `gorm:"not null;index;autoCreateTime" json:"paid_at"`
	Amount    decimal.Decimal `gorm:"type:numeric(12,2);not null;check:amount > 0" json:"amount"`
	Method    string          `gorm:"type:varchar(10);not null;check:method IN ('cash','transfer','card','credit')" json:"method"` // credit = จ่ายด้วยกระเป๋าสินค้า VIP
	Reference string          `gorm:"type:varchar(100)" json:"reference"`                                                          // เลขสลิป หรือเลขท้ายบัตร
}

func (Payment) TableName() string { return "payments" }
