package course

import (
	"encoding/json"
	"testing"

	"backend/internal/models"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToCourseResponse_ItemsKeepOnlyStockRef(t *testing.T) {
	validDay := 180
	got := toCourseResponse(models.Course{
		ID: idFacial, Name: "หน้าใส", Price: decimal.NewFromInt(3500), MaxSession: 10, ValidDay: &validDay,
		ServiceComm: decimal.NewFromInt(250), Category: "facial", IsActive: true,
		Items: []models.CourseItem{
			{CourseID: idFacial, StockID: "stock-1", UseQty: 2, Stock: &models.Stock{ID: "stock-1", Name: "เจล", Price: decimal.NewFromInt(99)}},
		},
	})

	assert.Equal(t, idFacial, got.ID)
	assert.Equal(t, &validDay, got.ValidDay)
	assert.True(t, got.ServiceComm.Equal(decimal.NewFromInt(250)))
	assert.Equal(t, []CourseItemResponse{{Stock: StockRef{ID: "stock-1", Name: "เจล"}, UseQty: 2}}, got.Items)

	raw, err := json.Marshal(got.Items[0])
	require.NoError(t, err)
	assert.JSONEq(t, `{"stock":{"id":"stock-1","name":"เจล"},"use_qty":2}`, string(raw),
		"item ต้องมีแค่ stock (id, name) กับ use_qty")
}

// ไม่ preload stock (nil) ต้องไม่ panic และยังมี id จาก foreign key
func TestToCourseResponse_NilStockKeepsID(t *testing.T) {
	got := toCourseResponse(models.Course{Items: []models.CourseItem{{StockID: "stock-1", UseQty: 1}}})

	assert.Equal(t, []CourseItemResponse{{Stock: StockRef{ID: "stock-1"}, UseQty: 1}}, got.Items)
}

// ไม่มีของที่ใช้ต้องได้ [] ไม่ใช่ null หน้าบ้านจะได้ map ได้เลย
func TestToCourseResponse_NoItemsIsEmptyArray(t *testing.T) {
	raw, err := json.Marshal(toCourseResponse(models.Course{}))
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(raw, &body))
	assert.Equal(t, []any{}, body["items"])
}
