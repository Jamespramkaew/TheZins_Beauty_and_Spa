package stocklog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// repository ส่วนที่ต้องใช้ DB จริง (JOIN, SUM, preload, แบ่งหน้า) ทดสอบด้วย integration test
// ไฟล์นี้ทดสอบเฉพาะฟังก์ชันล้วนที่ไม่แตะ DB

// escapeLike กัน % และ _ ที่ user พิมพ์มาถูกตีความเป็น wildcard ของ ILIKE
func TestEscapeLike(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"ข้อความปกติ", "ครีม", "ครีม"},
		{"ว่าง", "", ""},
		{"เปอร์เซ็นต์", "100%", `100\%`},
		{"underscore", "a_b", `a\_b`},
		{"backslash ต้อง escape ก่อน", `a\b`, `a\\b`},
		{"ผสมทุกตัว", `%_\`, `\%\_\\`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, escapeLike(tt.in))
		})
	}
}
