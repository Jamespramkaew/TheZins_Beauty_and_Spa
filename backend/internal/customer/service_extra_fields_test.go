package customer

import (
	"context"
	"errors"
	"strings"
	"testing"

	"backend/internal/models"
	"backend/package"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeRepo เก็บสิ่งที่ service ส่งให้ Create / Update ไว้ตรวจ
type writeRepo struct {
	Repository
	created *models.Customer
	updated map[string]any
}

func (r *writeRepo) Create(_ context.Context, c *models.Customer) error {
	r.created = c
	return nil
}

func (r *writeRepo) GetByID(_ context.Context, id string) (*models.Customer, error) {
	return &models.Customer{ID: id}, nil
}

func (r *writeRepo) Update(_ context.Context, _ string, fields map[string]any) error {
	r.updated = fields
	return nil
}

const testCustomerID = "11111111-1111-1111-1111-111111111111"

func requireDetail(t *testing.T, err error, detail string) {
	t.Helper()
	var appErr *_package.AppError
	require.True(t, errors.As(err, &appErr), "ต้องเป็น AppError: %v", err)
	assert.Equal(t, _package.CodeValidation, appErr.Code)
	assert.Equal(t, detail, appErr.Detail)
}

// ชื่อเล่น โรคประจำตัว และข้อมูลเพิ่มเติม ตัดช่องว่างแล้วเก็บ
func TestCreate_SavesTrimmedExtraFields(t *testing.T) {
	repo := &writeRepo{}
	_, err := NewService(repo).Create(context.Background(), CreateCustomer{
		HnNo: "HN1", Fname: "สมหญิง",
		Nickname: " หญิง ", ChronicDisease: " เบาหวาน ", Note: " แพ้ยาเพนิซิลลิน ",
	})
	require.NoError(t, err)
	assert.Equal(t, "หญิง", repo.created.Nickname)
	assert.Equal(t, "เบาหวาน", repo.created.ChronicDisease)
	assert.Equal(t, "แพ้ยาเพนิซิลลิน", repo.created.Note)
}

// ไม่บังคับกรอก: ไม่ส่งมาเลยก็บันทึกได้
func TestCreate_ExtraFieldsAreOptional(t *testing.T) {
	repo := &writeRepo{}
	_, err := NewService(repo).Create(context.Background(), CreateCustomer{HnNo: "HN1", Fname: "สมหญิง"})
	require.NoError(t, err)
	assert.Empty(t, repo.created.Nickname)
	assert.Empty(t, repo.created.ChronicDisease)
	assert.Empty(t, repo.created.Note)
}

// นับเป็นตัวอักษร (ภาษาไทย 1 ตัว = หลาย byte) ยาวพอดีผ่าน เกิน 1 ตัวไม่ผ่าน
func TestCreate_RejectsTooLongExtraFields(t *testing.T) {
	for _, tc := range []struct {
		name string
		req  CreateCustomer
		ok   bool
		want string
	}{
		{"nickname พอดี", CreateCustomer{Nickname: strings.Repeat("ก", maxNicknameLen)}, true, ""},
		{"nickname เกิน", CreateCustomer{Nickname: strings.Repeat("ก", maxNicknameLen+1)}, false, "nickname must be at most 50 characters"},
		{"chronic_disease พอดี", CreateCustomer{ChronicDisease: strings.Repeat("ก", maxNoteLen)}, true, ""},
		{"chronic_disease เกิน", CreateCustomer{ChronicDisease: strings.Repeat("ก", maxNoteLen+1)}, false, "chronic_disease must be at most 1000 characters"},
		{"note เกิน", CreateCustomer{Note: strings.Repeat("ก", maxNoteLen+1)}, false, "note must be at most 1000 characters"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.HnNo, tc.req.Fname = "HN1", "สมหญิง"
			repo := &writeRepo{}
			_, err := NewService(repo).Create(context.Background(), tc.req)
			if tc.ok {
				require.NoError(t, err)
				return
			}
			requireDetail(t, err, tc.want)
			assert.Nil(t, repo.created, "ไม่ควรถึง repository")
		})
	}
}

// PATCH: ส่งมาแค่ไหนแก้แค่นั้น และส่ง "" = ล้างค่า
func TestUpdate_ExtraFields(t *testing.T) {
	nickname, empty := " หญิง ", ""
	repo := &writeRepo{}
	_, err := NewService(repo).Update(context.Background(), testCustomerID, UpdateCustomer{Nickname: &nickname, Note: &empty})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"nickname": "หญิง", "note": ""}, repo.updated)
}

func TestUpdate_RejectsTooLongNickname(t *testing.T) {
	long := strings.Repeat("ก", maxNicknameLen+1)
	repo := &writeRepo{}
	_, err := NewService(repo).Update(context.Background(), testCustomerID, UpdateCustomer{Nickname: &long})
	requireDetail(t, err, "nickname must be at most 50 characters")
	assert.Nil(t, repo.updated)
}
