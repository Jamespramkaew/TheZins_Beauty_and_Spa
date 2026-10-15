package customer

// CreateCustomer รับ body ของ POST /api/customers
type CreateCustomer struct {
	HnNo      string  `json:"hn_no"` // ร้านกำหนดเอง ห้ามซ้ำ
	Fname     string  `json:"fname"`
	Lname     string  `json:"lname"`
	Tel       string  `json:"tel"`
	Address   string  `json:"address"`
	CitizenID *string `json:"citizen_id"`
	BDay      *string `json:"b_day"` // รูปแบบ YYYY-MM-DD เช่น 2000-01-31
}

// UpdateCustomer รับ body ของ PATCH /api/customers/:id
// ทุก field เป็น pointer: nil = ไม่ได้ส่งมา ไม่ต้องแก้
type UpdateCustomer struct {
	HnNo      *string `json:"hn_no"`
	Fname     *string `json:"fname"`
	Lname     *string `json:"lname"`
	Tel       *string `json:"tel"`
	Address   *string `json:"address"`
	CitizenID *string `json:"citizen_id"`
	BDay      *string `json:"b_day"`
}
