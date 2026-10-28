package _package

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePaging(t *testing.T) {
	tests := []struct {
		name               string
		rawPage, rawSize   string
		wantPage, wantSize int
		wantErrDetail      string
	}{
		{"ไม่ส่ง = ค่าเริ่มต้น", "", "", DefaultPage, DefaultPageSize, ""},
		{"ค่าปกติ", "3", "50", 3, 50, ""},
		{"มีช่องว่างรอบตัวเลข", " 2 ", " 10 ", 2, 10, ""},
		{"page_size สูงสุด", "1", "100", 1, MaxPageSize, ""},
		{"page_size เกินสูงสุด", "1", "101", 0, 0, "page_size must be a number between 1 and 100"},
		{"page_size เป็น 0", "1", "0", 0, 0, "page_size must be a number between 1 and 100"},
		{"page ติดลบ", "-1", "", 0, 0, "page must be a number between 1 and 1000000"},
		{"page ไม่ใช่ตัวเลข", "abc", "", 0, 0, "page must be a number between 1 and 1000000"},
		{"page เป็นทศนิยม", "2.5", "", 0, 0, "page must be a number between 1 and 1000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, size, err := ParsePaging(tt.rawPage, tt.rawSize)
			if tt.wantErrDetail != "" {
				var appErr *AppError
				require.True(t, errors.As(err, &appErr), "ต้องเป็น AppError: %v", err)
				assert.Equal(t, CodeValidation, appErr.Code)
				assert.Equal(t, tt.wantErrDetail, appErr.Detail)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPage, page)
			assert.Equal(t, tt.wantSize, size)
		})
	}
}
