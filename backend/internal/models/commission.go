package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// CommissionRate
type CommissionRate struct {
	ID        string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Type      string              `gorm:"type:varchar(20);not null;check:type IN ('course_sale','product_sale')" json:"type"`
	MinAmount decimal.Decimal     `gorm:"type:numeric(12,2);not null" json:"min_amount"`
	MaxAmount decimal.NullDecimal `gorm:"type:numeric(12,2)" json:"max_amount"`   // null = ไม่มีเพดาน
	Rate      decimal.Decimal     `gorm:"type:numeric(5,2);not null" json:"rate"` // เป็น % เช่น 5.00
}

func (CommissionRate) TableName() string { return "commission_rates" }

// CommissionPayout
// เมื่อ paid_at มีค่าแล้ว ถือว่าล็อก ห้ามแก้
type CommissionPayout struct {
	ID              string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EmployeeID      string          `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee        *Employee       `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	PeriodStart     time.Time       `gorm:"type:date;not null" json:"period_start"`
	PeriodEnd       time.Time       `gorm:"type:date;not null" json:"period_end"`
	CourseSaleComm  decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"course_sale_comm"`  // รวมค่าคอมขายคอร์ส
	ProductSaleComm decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"product_sale_comm"` // รวมค่าคอมขายสินค้า
	ServiceComm     decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"service_comm"`      // รวมค่ามือบริการ
	Total           decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"total"`
	PaidAt          *time.Time      `json:"paid_at"` // null = ยังไม่จ่าย
	Note            string          `gorm:"type:text" json:"note"`
}

func (CommissionPayout) TableName() string { return "commission_payouts" }
