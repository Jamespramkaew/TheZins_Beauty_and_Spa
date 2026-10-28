package _package

import (
	"fmt"
	"strconv"
	"strings"
)

// ค่าแบ่งหน้าที่ใช้เหมือนกันทุกรายการ (?page=&page_size=)
const (
	DefaultPage     = 1
	MaxPage         = 1_000_000
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// ParsePaging แปลง ?page= และ ?page_size= ที่รับมาเป็นข้อความให้เป็นตัวเลข ว่าง = ค่าเริ่มต้น
// ไม่ใช่ตัวเลข หรืออยู่นอกช่วง → AppError 400 พร้อมข้อความที่คนอ่านเข้าใจ (ไม่ส่งข้อความดิบของ strconv)
func ParsePaging(rawPage, rawPageSize string) (page, pageSize int, err error) {
	page, err = parseBoundedInt("page", rawPage, DefaultPage, MaxPage)
	if err != nil {
		return 0, 0, err
	}
	pageSize, err = parseBoundedInt("page_size", rawPageSize, DefaultPageSize, MaxPageSize)
	if err != nil {
		return 0, 0, err
	}
	return page, pageSize, nil
}

func parseBoundedInt(name, raw string, def, max int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > max {
		return 0, NewAppError("Invalid request", CodeValidation,
			fmt.Sprintf("%s must be a number between 1 and %d", name, max))
	}
	return n, nil
}
