//go:build integration

package integration

import (
	"testing"
	"time"

	"backend/internal/customer"
	"backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ลูกค้า 5 คนที่ created_at เท่ากันทุกคน: เดินทีละหน้า (หน้าละ 2) ต้องได้ครบ 5 คน ไม่ซ้ำ ไม่หาย
// (ถ้าเรียงแค่ created_at โดยไม่มี id ตัดสิน Postgres อาจคืนลำดับต่างกันในแต่ละหน้า)
func TestCustomerRepo_Search_PagesWithoutDuplicatesOrGaps(t *testing.T) {
	tx := newTx(t)
	repo := customer.NewRepository(tx)
	prefix := uniqueName("itest-page")
	sameTime := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)

	want := map[string]bool{}
	for i := 0; i < 5; i++ {
		c := models.Customer{Fname: prefix, Tier: "general", HnNo: "P" + tag(), CreatedAt: sameTime}
		require.NoError(t, tx.Create(&c).Error)
		want[c.ID] = true
	}

	seen := map[string]bool{}
	for page := 1; page <= 3; page++ {
		rows, total, err := repo.Search(ctx, prefix, page, 2)
		require.NoError(t, err)
		assert.EqualValues(t, 5, total, "total ต้องนับทุกคนที่ตรง q ไม่ใช่แค่หน้านี้")
		if page < 3 {
			assert.Len(t, rows, 2)
		} else {
			assert.Len(t, rows, 1, "หน้าสุดท้ายเหลือ 1 คน")
		}
		for _, r := range rows {
			assert.False(t, seen[r.ID], "ลูกค้า %s โผล่ซ้ำข้ามหน้า", r.ID)
			seen[r.ID] = true
		}
	}
	assert.Equal(t, want, seen, "ต้องเห็นครบทุกคน")

	rows, total, err := repo.Search(ctx, prefix, 4, 2)
	require.NoError(t, err)
	assert.EqualValues(t, 5, total)
	assert.Empty(t, rows, "เกินหน้าสุดท้าย = ว่าง ไม่ error")
}
