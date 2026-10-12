package database

import (
	"time"
)

// ---------- Example ----------

type Example struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Message   string    `gorm:"type:varchar(255);" json:"message,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Example) TableName() string {
	return "examples"
}

// ---------- XXXXXX ----------