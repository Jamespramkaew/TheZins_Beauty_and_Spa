//go:build integration

package integration

import (
	"testing"
	"time"

	"backend/internal/models"
	"backend/internal/order"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// mkPaidOrder สร้างบิลราคา 100 บาท 1 ใบ (รายการสินค้า 1 บรรทัด) จ่ายแล้วตาม paid (0 = ยังไม่จ่าย)
func mkPaidOrder(t *testing.T, tx *gorm.DB, branchID, customerID, stockID string, date time.Time, paid int64) models.Order {
	t.Helper()
	o := models.Order{OrderNo: "T" + tag(), CustomerID: customerID, BranchID: branchID, OrderDate: date}
	require.NoError(t, tx.Create(&o).Error)
	item := models.OrderItem{OrderID: o.ID, LineNo: 1, ItemType: "stock", StockID: &stockID, Qty: 1, UnitPrice: decimal.NewFromInt(100)}
	require.NoError(t, tx.Create(&item).Error)
	if paid > 0 {
		p := models.Payment{OrderID: o.ID, Amount: decimal.NewFromInt(paid), Method: "cash"}
		require.NoError(t, tx.Create(&p).Error)
	}
	return o
}

// บิล 5 ใบวันเดียวกันของสาขาเดียว: เดินทีละหน้า (หน้าละ 2) ต้องได้ครบ ไม่ซ้ำ ไม่หาย
// และ total ต้องนับด้วยเงื่อนไขเดียวกับรายการ (รวมถึง status ที่คำนวณจากยอดจ่าย)
func TestOrderRepo_List_PagesAndCountsWithFilters(t *testing.T) {
	tx := newTx(t)
	repo := order.NewRepository(tx)
	branch := mkBranch(t, tx)
	customer := models.Customer{Fname: "itest", Tier: "general", HnNo: "T" + tag()}
	require.NoError(t, tx.Create(&customer).Error)
	stock := mkStock(t, tx, uniqueName("itest-order-stock"))
	sameDay := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	want := map[string]bool{}
	for _, paid := range []int64{100, 100, 40, 0, 0} { // paid 2 · partial 1 · unpaid 2
		o := mkPaidOrder(t, tx, branch.ID, customer.ID, stock.ID, sameDay, paid)
		want[o.ID] = true
	}

	seen := map[string]bool{}
	for page := 1; page <= 3; page++ {
		rows, total, err := repo.List(ctx, order.ListFilter{BranchID: &branch.ID, Page: page, PageSize: 2})
		require.NoError(t, err)
		assert.EqualValues(t, 5, total, "total ต้องนับทุกบิลของสาขานี้ ไม่ใช่แค่หน้านี้")
		for _, r := range rows {
			assert.False(t, seen[r.ID], "บิล %s โผล่ซ้ำข้ามหน้า", r.ID)
			seen[r.ID] = true
		}
	}
	assert.Equal(t, want, seen, "ต้องเห็นครบทุกบิล")

	for status, wantTotal := range map[string]int64{"paid": 2, "partial": 1, "unpaid": 2} {
		s := status
		rows, total, err := repo.List(ctx, order.ListFilter{BranchID: &branch.ID, Status: &s, Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.EqualValues(t, wantTotal, total, "total ของ status=%s", status)
		assert.Len(t, rows, int(wantTotal), "จำนวนแถวของ status=%s", status)
	}
}
