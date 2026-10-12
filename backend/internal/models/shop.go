package models

import "time"

// Branch
type Branch struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"type:varchar(100);not null" json:"name"`
	IsActive bool   `gorm:"not null;default:true" json:"is_active"`
}

func (Branch) TableName() string { return "branches" }

// User
type User struct {
	ID       uint    `gorm:"primaryKey" json:"id"`
	Username string  `gorm:"type:varchar(50);not null;uniqueIndex" json:"username"`
	Password string  `gorm:"type:varchar(255);not null" json:"-"`                                                                                                // เก็บ bcrypt hash เท่านั้น
	BranchID *uint   `gorm:"check:chk_users_role_branch,(role = 'admin' AND branch_id IS NULL) OR (role = 'branch' AND branch_id IS NOT NULL)" json:"branch_id"` // null = admin
	Branch   *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Role     string  `gorm:"type:varchar(10);not null;check:role IN ('branch','admin')" json:"role"`
}

func (User) TableName() string { return "users" }

// Employee
type Employee struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Role      string    `gorm:"type:varchar(10);not null;check:role IN ('consult','manager','TP')" json:"role"`
	BranchID  uint      `gorm:"not null;index" json:"branch_id"`
	Branch    *Branch   `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func (Employee) TableName() string { return "employees" }

// Room
type Room struct {
	ID       uint    `gorm:"primaryKey" json:"id"`
	BranchID uint    `gorm:"not null" json:"branch_id"`
	Branch   *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	RoomName string  `gorm:"type:varchar(50);not null" json:"room_name"`
	IsActive bool    `gorm:"not null;default:true" json:"is_active"`
}

func (Room) TableName() string { return "rooms" }
