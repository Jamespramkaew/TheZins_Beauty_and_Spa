package customer

import (
	"context"
	"errors"
	"testing"

	"backend/internal/models"
	"backend/package"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRepo ทำเฉพาะ Search ที่เทสต์ใช้ (method อื่นมาจาก Repository ที่เป็น nil ถ้าเรียกจะ panic)
type fakeRepo struct {
	Repository
	called bool
	q      string
	tier   string
}

func (f *fakeRepo) Search(_ context.Context, q, tier string, _, _ int) ([]models.Customer, int64, error) {
	f.called, f.q, f.tier = true, q, tier
	return []models.Customer{}, 0, nil
}

func TestSearch_PassesTrimmedTierToRepository(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"vip", "vip"}, {" general ", "general"}, {"", ""}} {
		repo := &fakeRepo{}
		_, _, err := NewService(repo).Search(context.Background(), SearchCustomersQuery{Q: " สม ", Tier: tc.in})
		require.NoError(t, err, "tier %q", tc.in)
		assert.Equal(t, tc.want, repo.tier, "tier %q", tc.in)
		assert.Equal(t, "สม", repo.q)
	}
}

// ค่าที่ไม่ใช่ general / vip ต้องตอบ 400 ก่อนถึง DB (VIP ตัวใหญ่ก็ไม่รับ เหมือน POST /tier)
func TestSearch_RejectsUnknownTier(t *testing.T) {
	for _, tier := range []string{"member", "VIP", "all"} {
		repo := &fakeRepo{}
		_, _, err := NewService(repo).Search(context.Background(), SearchCustomersQuery{Tier: tier})

		var appErr *_package.AppError
		require.True(t, errors.As(err, &appErr), "tier %q", tier)
		assert.Equal(t, _package.CodeValidation, appErr.Code)
		assert.Equal(t, "tier must be general or vip", appErr.Detail)
		assert.False(t, repo.called, "tier %q ไม่ควรถึง repository", tier)
	}
}
