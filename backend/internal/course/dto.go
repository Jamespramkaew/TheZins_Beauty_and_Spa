package course

import "github.com/shopspring/decimal"

type AddCourseRequest struct {
	Name        string          `json:"name" validate:"required,max=150"`
	Price       decimal.Decimal `json:"price" validate:"required,gte=0"`
	MaxSession  int             `json:"maxSession" validate:"required,gte=1"`
	ValidDay    *int            `json:"validDay" validate:"omitempty,gte=1"`
	ServiceComm decimal.Decimal `json:"serviceComm" validate:"gte=0"`
	Category    string          `json:"category" validate:"max=50"`
}
