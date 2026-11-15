package customer

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"backend/internal/models"
	"backend/package"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRepo ทำเฉพาะ Search ที่เทสต์ใช้ (method อื่นมาจาก Repository ที่เป็น nil ถ้าเรียกจะ panic)
type fakeRepo struct {
	Repository
	called bool
	got    SearchFilter
}

func (f *fakeRepo) Search(_ context.Context, filter SearchFilter) ([]models.Customer, int64, error) {
	f.called, f.got = true, filter
	return []models.Customer{}, 0, nil
}

func TestSearch_PassesTrimmedTierToRepository(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"vip", "vip"}, {" general ", "general"}, {"credit", "credit"}, {"", ""}} {
		repo := &fakeRepo{}
		_, _, err := NewService(repo).Search(context.Background(), SearchCustomersQuery{Q: " สม ", Tier: tc.in})
		require.NoError(t, err, "tier %q", tc.in)
		assert.Equal(t, tc.want, repo.got.Tier, "tier %q", tc.in)
		assert.Equal(t, "สม", repo.got.Q)
	}
}

// ค่าที่ไม่ใช่ general / vip ต้องตอบ 400 ก่อนถึง DB (VIP ตัวใหญ่ก็ไม่รับ เหมือน POST /tier)
func TestSearch_RejectsUnknownTier(t *testing.T) {
	for _, tier := range []string{"member", "VIP", "CREDIT", "all"} {
		repo := &fakeRepo{}
		_, _, err := NewService(repo).Search(context.Background(), SearchCustomersQuery{Tier: tier})

		var appErr *_package.AppError
		require.True(t, errors.As(err, &appErr), "tier %q", tier)
		assert.Equal(t, _package.CodeValidation, appErr.Code)
		assert.Equal(t, "tier must be general, vip or credit", appErr.Detail)
		assert.False(t, repo.called, "tier %q ไม่ควรถึง repository", tier)
	}
}

// joined ส่งต่อพร้อมเส้นแบ่ง = ตอนนี้ − NewCustomerDays วัน
func TestSearch_PassesJoinedAndNewSince(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"new", "new"}, {" old ", "old"}, {"", ""}} {
		repo := &fakeRepo{}
		before := time.Now()
		_, _, err := NewService(repo).Search(context.Background(), SearchCustomersQuery{Joined: tc.in})
		after := time.Now()
		require.NoError(t, err, "joined %q", tc.in)
		assert.Equal(t, tc.want, repo.got.Joined, "joined %q", tc.in)

		window := NewCustomerDays * 24 * time.Hour
		assert.False(t, repo.got.NewSince.Before(before.Add(-window)), "NewSince ต้องไม่เก่ากว่า %d วัน", NewCustomerDays)
		assert.False(t, repo.got.NewSince.After(after.Add(-window)), "NewSince ต้องไม่ใหม่กว่า %d วัน", NewCustomerDays)
	}
}

func TestSearch_RejectsUnknownJoined(t *testing.T) {
	for _, joined := range []string{"recent", "NEW", "true"} {
		repo := &fakeRepo{}
		_, _, err := NewService(repo).Search(context.Background(), SearchCustomersQuery{Joined: joined})

		var appErr *_package.AppError
		require.True(t, errors.As(err, &appErr), "joined %q", joined)
		assert.Equal(t, _package.CodeValidation, appErr.Code)
		assert.Equal(t, "joined must be new or old", appErr.Detail)
		assert.False(t, repo.called, "joined %q ไม่ควรถึง repository", joined)
	}
}

// tierRepo จำค่าที่ service ส่งให้ ChangeTier
type tierRepo struct {
	Repository
	gotTier string
}

func (r *tierRepo) ChangeTier(_ context.Context, id, newTier, _ string) (*models.Customer, error) {
	r.gotTier = newTier
	return &models.Customer{ID: id, Tier: newTier}, nil
}

// เปลี่ยนเป็นระดับไหนก็ได้ในสามระดับ (ตัดช่องว่างให้)
func TestChangeTier_AcceptsAllTiers(t *testing.T) {
	for _, tier := range []string{"general", "vip", " credit "} {
		repo := &tierRepo{}
		got, err := NewService(repo).ChangeTier(context.Background(), testCustomerID, "user-1", ChangeTier{Tier: tier})
		require.NoError(t, err, "tier %q", tier)
		assert.Equal(t, strings.TrimSpace(tier), repo.gotTier)
		assert.Equal(t, strings.TrimSpace(tier), got.Tier)
	}
}

func TestChangeTier_RejectsUnknownTier(t *testing.T) {
	for _, tier := range []string{"Credit", "gold", ""} {
		repo := &tierRepo{}
		_, err := NewService(repo).ChangeTier(context.Background(), testCustomerID, "user-1", ChangeTier{Tier: tier})
		requireDetail(t, err, "tier must be general, vip or credit")
		assert.Empty(t, repo.gotTier, "tier %q ไม่ควรถึง repository", tier)
	}
}
