package models

func All() []any {
	return []any{
		&Branch{}, &User{}, &Employee{}, &Room{},
		&Customer{}, &TierChange{}, &Course{}, &Stock{}, &CourseItem{},
		&CommissionRate{}, &CommissionPayout{},
		&Order{}, &OrderItem{}, &OrderItemSeller{}, &Payment{},
		&Appointment{}, &AppointmentStaff{}, &ServiceUsage{}, &UsageStaff{}, &StockLog{}, &StockContainer{},
	}
}
