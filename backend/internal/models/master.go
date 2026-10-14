package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Customer
type Customer struct {
	ID        string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Fname     string     `gorm:"type:varchar(100);not null" json:"fname"`
	Lname     string     `gorm:"type:varchar(100)" json:"lname"`
	Address   string     `gorm:"type:text" json:"address"`
	CitizenID *string    `gorm:"type:varchar(13)" json:"citizen_id"`
	Tier      string     `gorm:"type:varchar(10);not null;default:'general';check:tier IN ('general','vip')" json:"tier"`
	HnNo      string     `gorm:"type:varchar(20);not null;uniqueIndex" json:"hn_no"` // ร้านเป็นคนกำหนดเลข
	Tel       string     `gorm:"type:varchar(20);index" json:"tel"`
	BDay      *time.Time `gorm:"type:date" json:"b_day"`
	CreatedAt time.Time  `json:"created_at"`
}

func (Customer) TableName() string { return "customers" }

// TierChange
type TierChange struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomerID    string    `gorm:"type:uuid;not null;index" json:"customer_id"`
	Customer      *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	NewTier       string    `gorm:"type:varchar(10);not null;check:new_tier IN ('general','vip')" json:"new_tier"`
	ChangedBy     string    `gorm:"type:uuid;not null" json:"changed_by"` // user ที่เป็นคนเปลี่ยน
	ChangedByUser *User     `gorm:"foreignKey:ChangedBy" json:"changed_by_user,omitempty"`
	ChangedAt     time.Time `gorm:"not null;autoCreateTime" json:"changed_at"`
}

func (TierChange) TableName() string { return "tier_changes" }

// Course
type Course struct {
	ID          string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string          `gorm:"type:varchar(150);not null" json:"name"`
	Price       decimal.Decimal `gorm:"type:numeric(12,2);not null;check:price >= 0" json:"price"`
	MaxSession  int             `gorm:"not null;check:max_session >= 1" json:"max_session"`
	ValidDay    *int            `json:"valid_day"`                                                 // จำนวนวันก่อนหมดอายุ, null = ไม่หมดอายุ
	ServiceComm decimal.Decimal `gorm:"type:numeric(12,2);not null;default:0" json:"service_comm"` // ค่ามือหมอนวดต่อครั้ง
	Category    string          `gorm:"type:varchar(50)" json:"category"`
	IsActive    bool            `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"deleted_at"`
}

func (Course) TableName() string { return "courses" }

// Stock
// นับเฉพาะของที่ยังไม่เปิด เป็นจำนวนเต็มเสมอ
// uses_per_unit = 1 → ตัดสต็อก use -1 ทุกครั้งที่ใช้บริการ
// uses_per_unit > 1 → ตัดสต็อก open -1 ตอนเปิด (ดู StockContainer)
type Stock struct {
	ID          string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string          `gorm:"type:varchar(150);not null" json:"name"`
	Price       decimal.Decimal `gorm:"type:numeric(12,2);not null;default:0" json:"price"`
	IsSellable  bool            `gorm:"not null" json:"is_sellable"` // false = ใช้ในบริการอย่างเดียว
	Category    string          `gorm:"type:varchar(10);not null;check:category IN ('premium','normal')" json:"category"`
	UsesPerUnit int             `gorm:"not null;default:1;check:uses_per_unit >= 1" json:"uses_per_unit"` // 1 ชิ้นใช้ได้กี่ครั้ง
	MinQty      int             `gorm:"not null;default:0" json:"min_qty"`                                // จุดเตือนสั่งของ (จำนวนชิ้นที่ยังไม่เปิด)
	IsActive    bool            `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"deleted_at"`
}

func (Stock) TableName() string { return "stocks" }

// CourseItem
// use_qty: ของที่ uses_per_unit = 1 → จำนวนชิ้นที่ตัดต่อการทำ 1 ครั้ง
//
//	ของที่ uses_per_unit > 1 → จำนวนครั้งของกระปุกที่ใช้ต่อการทำ 1 ครั้ง
type CourseItem struct {
	CourseID string  `gorm:"type:uuid;primaryKey" json:"course_id"`
	Course   *Course `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	StockID  string  `gorm:"type:uuid;primaryKey" json:"stock_id"`
	Stock    *Stock  `gorm:"foreignKey:StockID" json:"stock,omitempty"`
	UseQty   int     `gorm:"not null;check:use_qty > 0" json:"use_qty"`
}

func (CourseItem) TableName() string { return "course_items" }
