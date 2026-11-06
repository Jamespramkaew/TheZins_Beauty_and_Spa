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
		rows, total, err := repo.Search(ctx, customer.SearchFilter{Q: prefix, Page: page, PageSize: 2})
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

	rows, total, err := repo.Search(ctx, customer.SearchFilter{Q: prefix, Page: 4, PageSize: 2})
	require.NoError(t, err)
	assert.EqualValues(t, 5, total)
	assert.Empty(t, rows, "เกินหน้าสุดท้าย = ว่าง ไม่ error")
}

// tier ต้องกรองทั้งตอนมี q และไม่มี q และ total ต้องนับเฉพาะ tier นั้น
// คู่กับ q: เงื่อนไขค้นหามี OR อยู่ ถ้าไม่ถูกครอบวงเล็บ คนที่ชื่อหรือเบอร์ตรงจะหลุด tier มาด้วย
func TestCustomerRepo_Search_FiltersByTier(t *testing.T) {
	tx := newTx(t)
	repo := customer.NewRepository(tx)
	prefix := uniqueName("itest-tier")

	tiers := []string{"vip", "vip", "general", "general", "general"}
	for _, tier := range tiers {
		c := models.Customer{Fname: prefix, Tel: prefix, Tier: tier, HnNo: "T" + tag()}
		require.NoError(t, tx.Create(&c).Error)
	}

	for _, tc := range []struct {
		tier string
		want int64
	}{{"vip", 2}, {"general", 3}, {"", 5}} {
		rows, total, err := repo.Search(ctx, customer.SearchFilter{Q: prefix, Tier: tc.tier, Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.EqualValues(t, tc.want, total, "tier %q", tc.tier)
		assert.Len(t, rows, int(tc.want), "tier %q", tc.tier)
		for _, r := range rows {
			if tc.tier != "" {
				assert.Equal(t, tc.tier, r.Tier)
			}
		}
	}

	// ไม่มี q: ลูกค้าคนอื่นใน DB ก็นับด้วย จึงเช็คแค่ว่าทุกแถวเป็น vip
	rows, _, err := repo.Search(ctx, customer.SearchFilter{Tier: "vip", Page: 1, PageSize: 100})
	require.NoError(t, err)
	assert.NotEmpty(t, rows)
	for _, r := range rows {
		assert.Equal(t, "vip", r.Tier)
	}
}

// joined แบ่งด้วย created_at เทียบกับ NewSince: new = ตั้งแต่ NewSince ขึ้นไป · old = ก่อนหน้านั้น
// ใช้คู่กับ q และ tier ได้ และ total นับตามเงื่อนไขทั้งหมด
func TestCustomerRepo_Search_FiltersByJoined(t *testing.T) {
	tx := newTx(t)
	repo := customer.NewRepository(tx)
	prefix := uniqueName("itest-join")
	now := time.Now()
	newSince := now.Add(-customer.NewCustomerDays * 24 * time.Hour)

	for _, c := range []struct {
		age  time.Duration
		tier string
	}{
		{time.Hour, "vip"},                  // ใหม่
		{6*24*time.Hour + 23*time.Hour, ""}, // ใหม่ (เกือบครบ 7 วัน)
		{7*24*time.Hour + time.Hour, "vip"}, // เก่า (เกิน 7 วันมา 1 ชม.)
		{30 * 24 * time.Hour, ""},           // เก่า
	} {
		tier := c.tier
		if tier == "" {
			tier = "general"
		}
		row := models.Customer{Fname: prefix, Tier: tier, HnNo: "J" + tag(), CreatedAt: now.Add(-c.age)}
		require.NoError(t, tx.Create(&row).Error)
	}

	for _, tc := range []struct {
		joined, tier string
		want         int64
	}{{"new", "", 2}, {"old", "", 2}, {"", "", 4}, {"new", "vip", 1}, {"old", "vip", 1}, {"old", "general", 1}} {
		rows, total, err := repo.Search(ctx, customer.SearchFilter{Q: prefix, Tier: tc.tier, Joined: tc.joined, NewSince: newSince, Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.EqualValues(t, tc.want, total, "joined %q tier %q", tc.joined, tc.tier)
		assert.Len(t, rows, int(tc.want), "joined %q tier %q", tc.joined, tc.tier)
		for _, r := range rows {
			switch tc.joined {
			case "new":
				assert.False(t, r.CreatedAt.Before(newSince), "ลูกค้าใหม่ต้องเพิ่มหลัง NewSince")
			case "old":
				assert.True(t, r.CreatedAt.Before(newSince), "ลูกค้าเก่าต้องเพิ่มก่อน NewSince")
			}
		}
	}
}
