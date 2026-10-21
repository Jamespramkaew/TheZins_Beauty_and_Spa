package stocklog

import (
	"encoding/json"
	"testing"
	"time"

	"backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToStockLogResponse_FullyPreloaded(t *testing.T) {
	created := time.Date(2026, 10, 21, 4, 36, 8, 0, time.UTC)
	orderItemID, usageID, note := validOrderItemID, validUsageID, "แก้ที่คีย์ผิด"

	got := toStockLogResponse(models.StockLog{
		ID: "log-1", StockID: validStockID, BranchID: validBranchID, EmployeeID: validEmployeeID,
		QtyChange: -2, LogType: "adjust", OrderItemID: &orderItemID, UsageID: &usageID, Note: &note, CreatedAt: created,
		Stock:    &models.Stock{ID: validStockID, Name: "เซรั่ม"},
		Branch:   &models.Branch{ID: validBranchID, Code: "S2", Name: "สาขา 2"},
		Employee: &models.Employee{ID: validEmployeeID, Name: "จินตนา แสนสุข", Role: "TP"},
	})

	assert.Equal(t, "log-1", got.ID)
	assert.Equal(t, "adjust", got.LogType)
	assert.Equal(t, -2, got.QtyChange)
	assert.Equal(t, &note, got.Note)
	assert.Equal(t, &orderItemID, got.OrderItemID)
	assert.Equal(t, &usageID, got.UsageID)
	assert.True(t, got.CreatedAt.Equal(created))
	assert.Equal(t, StockRef{ID: validStockID, Name: "เซรั่ม"}, got.Stock)
	assert.Equal(t, BranchRef{ID: validBranchID, Code: "S2", Name: "สาขา 2"}, got.Branch)
	assert.Equal(t, EmployeeRef{ID: validEmployeeID, Name: "จินตนา แสนสุข", Role: "TP"}, got.Employee)
}

// ไม่ preload ความสัมพันธ์ (nil) ต้องไม่ panic และยังมี id จาก foreign key
func TestToStockLogResponse_NilRelationsKeepIDs(t *testing.T) {
	got := toStockLogResponse(models.StockLog{
		ID: "log-1", StockID: validStockID, BranchID: validBranchID, EmployeeID: validEmployeeID,
	})

	assert.Equal(t, StockRef{ID: validStockID}, got.Stock)
	assert.Equal(t, BranchRef{ID: validBranchID}, got.Branch)
	assert.Equal(t, EmployeeRef{ID: validEmployeeID}, got.Employee)
}

// JSON ต้องมีเฉพาะฟิลด์ที่ตั้งใจให้เห็น และไม่มี id ซ้ำที่ระดับบน
func TestStockLogResponse_JSONShape(t *testing.T) {
	got := toStockLogResponse(models.StockLog{
		ID: "log-1", StockID: validStockID, BranchID: validBranchID, EmployeeID: validEmployeeID,
		QtyChange: 1, LogType: "receive",
		Stock:    &models.Stock{ID: validStockID, Name: "x"},
		Branch:   &models.Branch{ID: validBranchID, Code: "S1", Name: "y"},
		Employee: &models.Employee{ID: validEmployeeID, Name: "z", Role: "manager"},
	})

	raw, err := json.Marshal(got)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))

	assert.ElementsMatch(t,
		[]string{"id", "log_type", "qty_change", "note", "order_item_id", "usage_id", "created_at", "stock", "branch", "employee"},
		keys(m))
	assert.ElementsMatch(t, []string{"id", "name"}, keys(m["stock"].(map[string]any)))
	assert.ElementsMatch(t, []string{"id", "code", "name"}, keys(m["branch"].(map[string]any)))
	assert.ElementsMatch(t, []string{"id", "name", "role"}, keys(m["employee"].(map[string]any)))

	// note / order_item_id / usage_id ไม่มีค่าต้องเป็น null ไม่ใช่หายไป
	assert.Contains(t, m, "note")
	assert.Nil(t, m["note"])
	assert.Nil(t, m["order_item_id"])
	assert.Nil(t, m["usage_id"])
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
