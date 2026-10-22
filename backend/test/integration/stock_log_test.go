//go:build integration

package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"backend/internal/branch"
	"backend/internal/employee"
	"backend/internal/models"
	"backend/internal/stock"
	stocklog "backend/internal/stock-log"
	_package "backend/package"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// repository ของ stock-log ใช้ filter ชนิด unexported จึงเรียกตรงจากนอก package ไม่ได้
// ไฟล์นี้จึงทดสอบผ่าน service + repository จริง (และ handler สำหรับรูปแบบ JSON) บน DB จริง
// ทุก test ทำงานใน transaction แล้ว rollback ข้อมูลใน DB dev (seed) ที่มีอยู่จึงไม่ถูกแตะ
// และทุก query กรองด้วยสินค้า/สาขาที่สร้างขึ้นเองเสมอ เพื่อไม่ปนกับข้อมูล seed

// ---------- helpers ----------

func slSvc(tx *gorm.DB) stocklog.Service {
	return stocklog.NewService(stocklog.NewRepository(tx), stock.NewRepository(tx), branch.NewRepository(tx), employee.NewRepository(tx))
}

func slQuantities(t *testing.T, svc stocklog.Service, q stocklog.GetAllStockQuantitiesQuery) ([]stocklog.StockQuantity, _package.PageMeta) {
	t.Helper()
	items, meta, err := svc.GetAllStockQuantities(ctx, "admin", "", q)
	require.NoError(t, err)
	return items, meta
}

func slLogs(t *testing.T, svc stocklog.Service, q stocklog.GetAllStockLogsQuery) ([]stocklog.StockLogResponse, _package.PageMeta) {
	t.Helper()
	items, meta, err := svc.GetStockLogsByBranchId(ctx, "admin", "", q)
	require.NoError(t, err)
	return items, meta
}

// slStock สร้างสินค้าพร้อม min_qty
func slStock(t *testing.T, tx *gorm.DB, name string, minQty int) models.Stock {
	t.Helper()
	s := mkStock(t, tx, name)
	if minQty != 0 {
		require.NoError(t, tx.Model(&models.Stock{}).Where("id = ?", s.ID).Update("min_qty", minQty).Error)
		s.MinQty = minQty
	}
	return s
}

// slSetInactive ตั้ง is_active = false (สร้างด้วย false ตรงๆ ไม่ได้ เพราะ GORM ข้ามค่า zero ที่มี default:true)
func slSetInactive(t *testing.T, tx *gorm.DB, model any, id string) {
	t.Helper()
	require.NoError(t, tx.Model(model).Where("id = ?", id).Update("is_active", false).Error)
}

func slLogAt(t *testing.T, tx *gorm.DB, stockID, branchID, employeeID, logType string, qty int, at time.Time) models.StockLog {
	t.Helper()
	l := models.StockLog{StockID: stockID, BranchID: branchID, EmployeeID: employeeID, QtyChange: qty, LogType: logType, CreatedAt: at}
	require.NoError(t, tx.Create(&l).Error)
	return l
}

func slQtyNames(items []stocklog.StockQuantity) []string {
	out := make([]string, 0, len(items))
	for _, i := range items {
		out = append(out, i.StockName)
	}
	return out
}

func slLogIDs(items []stocklog.StockLogResponse) []string {
	out := make([]string, 0, len(items))
	for _, i := range items {
		out = append(out, i.ID)
	}
	return out
}

// ---------- GET /stock-logs/quantities ----------

func TestStockLog_Quantities_SumsAllLogTypesPerBranch(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	p := uniqueName("itest-q-sum")
	s := slStock(t, tx, p, 5)
	a, b := mkBranch(t, tx), mkBranch(t, tx)
	empA, empB := mkEmployee(t, tx, a.ID, "manager"), mkEmployee(t, tx, b.ID, "manager")
	now := time.Now()

	// สาขา A: รับ 10, เปิดกระปุก -1, adjust +3, count_diff -2 = 10
	slLogAt(t, tx, s.ID, a.ID, empA.ID, "receive", 10, now)
	slLogAt(t, tx, s.ID, a.ID, empA.ID, "open", -1, now)
	slLogAt(t, tx, s.ID, a.ID, empA.ID, "adjust", 3, now)
	slLogAt(t, tx, s.ID, a.ID, empA.ID, "count_diff", -2, now)
	// สาขา B: รับ 4 (ต้องไม่ปนกับ A)
	slLogAt(t, tx, s.ID, b.ID, empB.ID, "receive", 4, now)

	itemsA, metaA := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: a.ID, Search: p})
	require.Len(t, itemsA, 1)
	assert.Equal(t, 10, itemsA[0].Quantity)
	assert.Equal(t, s.ID, itemsA[0].StockID)
	assert.Equal(t, p, itemsA[0].StockName)
	assert.Equal(t, a.ID, itemsA[0].BranchID)
	assert.Equal(t, a.Name, itemsA[0].BranchName)
	assert.Equal(t, 5, itemsA[0].MinQty)
	assert.Equal(t, int64(1), metaA.Total)

	itemsB, _ := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: b.ID, Search: p})
	require.Len(t, itemsB, 1)
	assert.Equal(t, 4, itemsB[0].Quantity)
}

func TestStockLog_Quantities_StockWithoutLogsShowsZero(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	p := uniqueName("itest-q-zero")
	slStock(t, tx, p, 0)
	branch := mkBranch(t, tx)

	items, meta := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: branch.ID, Search: p})

	require.Len(t, items, 1, "LEFT JOIN ต้องไม่ทำให้สินค้าที่ไม่เคยมี log หาย")
	assert.Equal(t, 0, items[0].Quantity)
	assert.Equal(t, int64(1), meta.Total)
}

func TestStockLog_Quantities_LowStock(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	p := uniqueName("itest-q-low")
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")

	atMin := slStock(t, tx, p+"-atmin", 5) // 5 <= 5 ต่ำ
	above := slStock(t, tx, p+"-above", 5) // 6 > 5 ไม่ต่ำ
	slStock(t, tx, p+"-zero", 0)           // 0 <= 0 ต่ำ
	slLogAt(t, tx, atMin.ID, branch.ID, emp.ID, "receive", 5, time.Now())
	slLogAt(t, tx, above.ID, branch.ID, emp.ID, "receive", 6, time.Now())

	t.Run("ไม่กรอง: ได้ครบและ is_low_stock ถูก", func(t *testing.T) {
		items, _ := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: branch.ID, Search: p})
		require.Len(t, items, 3)
		low := map[string]bool{}
		for _, i := range items {
			low[i.StockName] = i.IsLowStock
		}
		assert.True(t, low[p+"-atmin"])
		assert.False(t, low[p+"-above"])
		assert.True(t, low[p+"-zero"])
	})

	t.Run("lowStock=true: เฉพาะที่ต่ำ และ total ตรงกับที่กรอง", func(t *testing.T) {
		items, meta := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: branch.ID, Search: p, LowStock: true})
		assert.ElementsMatch(t, []string{p + "-atmin", p + "-zero"}, slQtyNames(items))
		assert.Equal(t, int64(2), meta.Total)
	})
}

func TestStockLog_Quantities_ActiveFilter(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	p := uniqueName("itest-q-act")
	branch := mkBranch(t, tx)

	slStock(t, tx, p+"-active", 0)
	inactive := slStock(t, tx, p+"-inactive", 0)
	slSetInactive(t, tx, &models.Stock{}, inactive.ID)
	deleted := slStock(t, tx, p+"-deleted", 0)
	require.NoError(t, tx.Delete(&models.Stock{}, "id = ?", deleted.ID).Error)

	query := func(active string) []string {
		items, _ := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: branch.ID, Search: p, Active: active})
		return slQtyNames(items)
	}

	assert.ElementsMatch(t, []string{p + "-active"}, query(""), "ค่าเริ่มต้น = เฉพาะที่ใช้งานอยู่")
	assert.ElementsMatch(t, []string{p + "-active"}, query("true"))
	assert.ElementsMatch(t, []string{p + "-inactive"}, query("false"), "false = ปิดใช้งานแต่ยังไม่ถูกลบ")
	assert.ElementsMatch(t, []string{p + "-active", p + "-inactive", p + "-deleted"}, query("all"), "all รวมที่ถูก soft delete")
}

func TestStockLog_Quantities_CategoryFilter(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	p := uniqueName("itest-q-cat")
	branch := mkBranch(t, tx)

	slStock(t, tx, p+"-normal", 0) // mkStock = normal
	premium := models.Stock{Name: p + "-premium", Category: "premium", UsesPerUnit: 1, IsSellable: true, IsActive: true}
	require.NoError(t, tx.Create(&premium).Error)

	premiumItems, _ := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: branch.ID, Search: p, Category: "premium"})
	assert.Equal(t, []string{p + "-premium"}, slQtyNames(premiumItems))

	normalItems, _ := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: branch.ID, Search: p, Category: "normal"})
	assert.Equal(t, []string{p + "-normal"}, slQtyNames(normalItems))
}

func TestStockLog_Quantities_SearchEscapesWildcards(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	p := uniqueName("itest-q-srch")
	branch := mkBranch(t, tx)

	slStock(t, tx, p+"-100%", 0)
	slStock(t, tx, p+"-100x", 0)
	slStock(t, tx, p+"-a_c", 0)
	slStock(t, tx, p+"-abc", 0)

	search := func(s string) []string {
		items, _ := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: branch.ID, Search: s})
		return slQtyNames(items)
	}

	assert.Equal(t, []string{p + "-100%"}, search(p+"-100%"), "% ต้องเป็นตัวอักษร ไม่ใช่ wildcard")
	assert.Equal(t, []string{p + "-a_c"}, search(p+"-a_c"), "_ ต้องเป็นตัวอักษร ไม่ใช่ wildcard")
	assert.Len(t, search(p), 4, "ค้นด้วย prefix ได้ครบ")
	assert.Equal(t, []string{p + "-abc"}, search(p[len("itest-"):]+"-abc"), "ค้นแบบ contains (กลางข้อความ) ได้")
}

func TestStockLog_Quantities_SearchIsCaseInsensitive(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	p := uniqueName("itest-q-case")
	branch := mkBranch(t, tx)
	slStock(t, tx, p, 0)

	upper := []rune(p)
	for i, r := range upper {
		if r >= 'a' && r <= 'z' {
			upper[i] = r - 32
		}
	}
	items, _ := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: branch.ID, Search: string(upper)})
	assert.Equal(t, []string{p}, slQtyNames(items))
}

func TestStockLog_Quantities_Pagination(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	p := uniqueName("itest-q-page")
	branch := mkBranch(t, tx)
	for _, n := range []string{"5", "3", "1", "4", "2"} {
		slStock(t, tx, p+"-"+n, 0)
	}

	page := func(n int) ([]string, _package.PageMeta) {
		items, meta := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: branch.ID, Search: p, Page: n, PageSize: 2})
		return slQtyNames(items), meta
	}

	p1, m1 := page(1)
	p2, _ := page(2)
	p3, _ := page(3)
	p4, m4 := page(4)

	assert.Equal(t, []string{p + "-1", p + "-2"}, p1, "เรียงชื่อน้อยไปมาก")
	assert.Equal(t, []string{p + "-3", p + "-4"}, p2)
	assert.Equal(t, []string{p + "-5"}, p3)
	assert.Empty(t, p4, "เกินหน้าสุดท้ายต้องว่าง")
	assert.Equal(t, int64(5), m1.Total)
	assert.Equal(t, 3, m1.TotalPages)
	assert.Equal(t, int64(5), m4.Total, "total ต้องถูกแม้เกินหน้า")
}

func TestStockLog_Quantities_BranchUserSeesOnlyOwnBranch(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	p := uniqueName("itest-q-own")
	s := slStock(t, tx, p, 0)
	a, b := mkBranch(t, tx), mkBranch(t, tx)
	slLogAt(t, tx, s.ID, a.ID, mkEmployee(t, tx, a.ID, "manager").ID, "receive", 7, time.Now())
	slLogAt(t, tx, s.ID, b.ID, mkEmployee(t, tx, b.ID, "manager").ID, "receive", 9, time.Now())

	// บัญชีสาขา A พยายามขอดูสาขา B ต้องได้ข้อมูลสาขา A เท่านั้น
	items, _, err := svc.GetAllStockQuantities(ctx, "branch", a.ID, stocklog.GetAllStockQuantitiesQuery{BranchID: b.ID, Search: p})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, a.ID, items[0].BranchID)
	assert.Equal(t, 7, items[0].Quantity)
}

func TestStockLog_Quantities_AdminWithoutBranchSeesAllActiveBranches(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	p := uniqueName("itest-q-all")
	slStock(t, tx, p, 0)
	a, b := mkBranch(t, tx), mkBranch(t, tx)
	closed := mkBranch(t, tx)
	slSetInactive(t, tx, &models.Branch{}, closed.ID)

	var activeBranches int64
	require.NoError(t, tx.Model(&models.Branch{}).Where("is_active = ?", true).Count(&activeBranches).Error)

	items, _ := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{Search: p, PageSize: 100})

	got := map[string]bool{}
	for _, i := range items {
		got[i.BranchID] = true
	}
	assert.Len(t, items, int(activeBranches), "ต้องได้ 1 แถวต่อสาขาที่ยังเปิดอยู่")
	assert.True(t, got[a.ID])
	assert.True(t, got[b.ID])
	assert.False(t, got[closed.ID], "สาขาที่ปิดแล้วต้องไม่แสดง")
}

// ---------- GET /stock-logs ----------

func TestStockLog_Logs_PreloadsRelationsAndOrdersNewestFirst(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-l-order"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "TP")
	base := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)

	old := slLogAt(t, tx, s.ID, branch.ID, emp.ID, "receive", 10, base)
	mid := slLogAt(t, tx, s.ID, branch.ID, emp.ID, "open", -1, base.Add(time.Hour))
	newest := slLogAt(t, tx, s.ID, branch.ID, emp.ID, "adjust", 2, base.Add(2*time.Hour))

	items, meta := slLogs(t, svc, stocklog.GetAllStockLogsQuery{StockID: s.ID})

	assert.Equal(t, []string{newest.ID, mid.ID, old.ID}, slLogIDs(items), "ใหม่สุดก่อน")
	assert.Equal(t, int64(3), meta.Total)

	first := items[0]
	assert.Equal(t, "adjust", first.LogType)
	assert.Equal(t, 2, first.QtyChange)
	assert.Equal(t, stocklog.StockRef{ID: s.ID, Name: s.Name}, first.Stock)
	assert.Equal(t, stocklog.BranchRef{ID: branch.ID, Code: branch.Code, Name: branch.Name}, first.Branch)
	assert.Equal(t, stocklog.EmployeeRef{ID: emp.ID, Name: emp.Name, Role: "TP"}, first.Employee)
	assert.Nil(t, first.Note)
}

func TestStockLog_Logs_SoftDeletedStockStillShowsName(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-l-deleted"), 0)
	branch := mkBranch(t, tx)
	slLogAt(t, tx, s.ID, branch.ID, mkEmployee(t, tx, branch.ID, "manager").ID, "receive", 5, time.Now())
	require.NoError(t, tx.Delete(&models.Stock{}, "id = ?", s.ID).Error)

	items, _ := slLogs(t, svc, stocklog.GetAllStockLogsQuery{StockID: s.ID})

	require.Len(t, items, 1, "log ห้ามหายเมื่อสินค้าถูกลบ")
	assert.Equal(t, s.Name, items[0].Stock.Name, "ประวัติต้องยังแสดงชื่อสินค้าที่ถูก soft delete")
}

func TestStockLog_Logs_NoteIsReturned(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-l-note"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")
	note := "คีย์ผิดเป็น 8 ที่ถูกคือ 10"
	require.NoError(t, tx.Create(&models.StockLog{
		StockID: s.ID, BranchID: branch.ID, EmployeeID: emp.ID, QtyChange: 2, LogType: "adjust", Note: &note,
	}).Error)

	items, _ := slLogs(t, svc, stocklog.GetAllStockLogsQuery{StockID: s.ID})

	require.Len(t, items, 1)
	require.NotNil(t, items[0].Note)
	assert.Equal(t, note, *items[0].Note)
}

func TestStockLog_Logs_Filters(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-l-filter"), 0)
	other := slStock(t, tx, uniqueName("itest-l-other"), 0)
	a, b := mkBranch(t, tx), mkBranch(t, tx)
	empA, empB := mkEmployee(t, tx, a.ID, "manager"), mkEmployee(t, tx, b.ID, "manager")
	now := time.Now()

	recvA := slLogAt(t, tx, s.ID, a.ID, empA.ID, "receive", 10, now)
	openA := slLogAt(t, tx, s.ID, a.ID, empA.ID, "open", -1, now)
	recvB := slLogAt(t, tx, s.ID, b.ID, empB.ID, "receive", 3, now)
	zeroB := slLogAt(t, tx, s.ID, b.ID, empB.ID, "count_diff", 0, now) // นับแล้วตรงพอดี: ไม่ใช่ทั้งเข้าและออก
	otherA := slLogAt(t, tx, other.ID, a.ID, empA.ID, "receive", 1, now)

	item := mkOrderItem(t, tx, s.ID)
	sale := models.StockLog{StockID: s.ID, BranchID: a.ID, EmployeeID: empA.ID, QtyChange: -1, LogType: "sale", OrderItemID: &item.ID, CreatedAt: now}
	require.NoError(t, tx.Create(&sale).Error)

	ids := func(q stocklog.GetAllStockLogsQuery) []string {
		items, _ := slLogs(t, svc, q)
		got := slLogIDs(items)
		sort.Strings(got)
		return got
	}
	sorted := func(l ...models.StockLog) []string {
		out := make([]string, 0, len(l))
		for _, x := range l {
			out = append(out, x.ID)
		}
		sort.Strings(out)
		return out
	}

	assert.Equal(t, sorted(recvA, openA, recvB, zeroB, sale), ids(stocklog.GetAllStockLogsQuery{StockID: s.ID}), "กรองตามสินค้า")
	assert.Equal(t, sorted(otherA), ids(stocklog.GetAllStockLogsQuery{StockID: other.ID}))
	assert.Equal(t, sorted(recvA, openA, sale), ids(stocklog.GetAllStockLogsQuery{StockID: s.ID, BranchID: a.ID}), "กรองตามสาขา")
	assert.Equal(t, sorted(recvB, zeroB), ids(stocklog.GetAllStockLogsQuery{StockID: s.ID, BranchID: b.ID}))
	assert.Equal(t, sorted(openA), ids(stocklog.GetAllStockLogsQuery{StockID: s.ID, LogType: "open"}), "กรองตามชนิด")
	assert.Equal(t, sorted(recvA, recvB), ids(stocklog.GetAllStockLogsQuery{StockID: s.ID, LogType: "receive"}))
	assert.Equal(t, sorted(recvA, recvB), ids(stocklog.GetAllStockLogsQuery{StockID: s.ID, Direction: "in"}), "in = qty_change > 0")
	assert.Equal(t, sorted(openA, sale), ids(stocklog.GetAllStockLogsQuery{StockID: s.ID, Direction: "out"}), "out = qty_change < 0")
	assert.Equal(t, sorted(sale), ids(stocklog.GetAllStockLogsQuery{StockID: s.ID, OrderItemID: item.ID}), "กรองตามรายการขาย")
	assert.Equal(t, sorted(zeroB), ids(stocklog.GetAllStockLogsQuery{StockID: s.ID, LogType: "count_diff"}))
	assert.Empty(t, ids(stocklog.GetAllStockLogsQuery{StockID: s.ID, LogType: "transfer"}))
}

func TestStockLog_Logs_DateRangeUsesThaiTime(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-l-date"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")

	// 23:59:59 ของ 1 ต.ค. เวลาไทย = 16:59:59 UTC / 00:00:00 ของ 2 ต.ค. เวลาไทย = 17:00:00 UTC
	endOfOct1 := slLogAt(t, tx, s.ID, branch.ID, emp.ID, "receive", 1, time.Date(2026, 10, 1, 16, 59, 59, 0, time.UTC))
	startOfOct2 := slLogAt(t, tx, s.ID, branch.ID, emp.ID, "receive", 1, time.Date(2026, 10, 1, 17, 0, 0, 0, time.UTC))

	ids := func(from, to string) []string {
		items, _ := slLogs(t, svc, stocklog.GetAllStockLogsQuery{StockID: s.ID, From: from, To: to})
		got := slLogIDs(items)
		sort.Strings(got)
		return got
	}
	both := []string{endOfOct1.ID, startOfOct2.ID}
	sort.Strings(both)

	assert.Equal(t, []string{endOfOct1.ID}, ids("2026-10-01", "2026-10-01"), "วันที่ 1 ต.ค. ครบทั้งวัน")
	assert.Equal(t, []string{startOfOct2.ID}, ids("2026-10-02", "2026-10-02"), "วันที่ 2 ต.ค. เริ่มที่เที่ยงคืนเวลาไทย")
	assert.Equal(t, both, ids("2026-10-01", "2026-10-02"))
	assert.Equal(t, []string{startOfOct2.ID}, ids("2026-10-02", ""), "มีแต่ from")
	assert.Equal(t, []string{endOfOct1.ID}, ids("", "2026-10-01"), "มีแต่ to")
	assert.Empty(t, ids("2026-10-03", ""))
}

// created_at เท่ากันต้องเรียงตาม id คงที่ ไม่งั้นข้ามหน้าแล้วแถวซ้ำ/ตกหล่น
func TestStockLog_Logs_PaginationStableOnTies(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-l-ties"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")
	same := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)

	want := map[string]bool{}
	for i := 0; i < 5; i++ {
		want[slLogAt(t, tx, s.ID, branch.ID, emp.ID, "receive", 1, same).ID] = true
	}

	var collected []string
	var lastMeta _package.PageMeta
	for page := 1; page <= 3; page++ {
		items, meta := slLogs(t, svc, stocklog.GetAllStockLogsQuery{StockID: s.ID, Page: page, PageSize: 2})
		collected = append(collected, slLogIDs(items)...)
		lastMeta = meta
	}

	require.Len(t, collected, 5)
	seen := map[string]bool{}
	for _, id := range collected {
		assert.False(t, seen[id], "แถวซ้ำข้ามหน้า: %s", id)
		assert.True(t, want[id], "แถวแปลกปลอม: %s", id)
		seen[id] = true
	}
	assert.Equal(t, int64(5), lastMeta.Total)
	assert.Equal(t, 3, lastMeta.TotalPages)
}

func TestStockLog_Logs_BranchUserSeesOnlyOwnBranch(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-l-own"), 0)
	a, b := mkBranch(t, tx), mkBranch(t, tx)
	logA := slLogAt(t, tx, s.ID, a.ID, mkEmployee(t, tx, a.ID, "manager").ID, "receive", 1, time.Now())
	slLogAt(t, tx, s.ID, b.ID, mkEmployee(t, tx, b.ID, "manager").ID, "receive", 1, time.Now())

	// บัญชีสาขา A ขอดูสาขา B ต้องได้เฉพาะ log ของสาขา A
	items, meta, err := svc.GetStockLogsByBranchId(ctx, "branch", a.ID, stocklog.GetAllStockLogsQuery{StockID: s.ID, BranchID: b.ID})
	require.NoError(t, err)
	assert.Equal(t, []string{logA.ID}, slLogIDs(items))
	assert.Equal(t, int64(1), meta.Total)
}

func TestStockLog_Logs_EmptyResult(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-l-empty"), 0)

	items, meta := slLogs(t, svc, stocklog.GetAllStockLogsQuery{StockID: s.ID})

	assert.NotNil(t, items)
	assert.Empty(t, items)
	assert.Equal(t, int64(0), meta.Total)
	assert.Equal(t, 0, meta.TotalPages)
}

// ---------- ข้อจำกัดของตาราง stock_logs ----------

func TestStockLog_DB_RejectsNullEmployee(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-l-null"), 0)
	branch := mkBranch(t, tx)

	err := expectDBError(t, tx, func() error {
		return tx.Exec(
			`INSERT INTO stock_logs (stock_id, branch_id, qty_change, log_type, employee_id) VALUES (?, ?, 1, 'receive', NULL)`,
			s.ID, branch.ID).Error
	})

	require.Error(t, err, "employee_id ต้อง NOT NULL")
	assert.Contains(t, err.Error(), "employee_id")
}

func TestStockLog_DB_RejectsUnknownEmployee(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-l-fk"), 0)
	branch := mkBranch(t, tx)

	err := expectDBError(t, tx, func() error {
		return tx.Create(&models.StockLog{
			StockID: s.ID, BranchID: branch.ID, QtyChange: 1, LogType: "receive",
			EmployeeID: "00000000-0000-4000-8000-000000000000",
		}).Error
	})

	assert.Error(t, err, "employee_id ต้องอ้างพนักงานที่มีอยู่จริง (FK)")
}

func TestStockLog_DB_RejectsUnknownLogType(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-l-check"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")

	err := expectDBError(t, tx, func() error {
		return tx.Create(&models.StockLog{
			StockID: s.ID, BranchID: branch.ID, EmployeeID: emp.ID, QtyChange: 1, LogType: "steal",
		}).Error
	})

	assert.Error(t, err, "log_type ต้องผ่าน CHECK constraint")
}

// ---------- รูปแบบ JSON จริงของทั้งสองเส้น (handler + service + repository + DB) ----------

func slApp(tx *gorm.DB, role, branchID string) *fiber.App {
	h := stocklog.NewHandler(slSvc(tx))
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("role", role)
		c.Locals("branchId", branchID)
		return c.Next()
	})
	app.Get("/stock-logs/quantities", h.GetAllStockQuantities)
	app.Get("/stock-logs", h.GetStockLogsByBranchId)
	app.Post("/stock-logs", h.AddStockLogMovement)
	return app
}

func slGet(t *testing.T, app *fiber.App, path string) (int, map[string]any) {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var body map[string]any
	require.NoError(t, json.Unmarshal(raw, &body), "response ต้องเป็น JSON: %s", raw)
	return resp.StatusCode, body
}

func slKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestStockLog_HTTP_LogsJSONShape(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-h-logs"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")
	slLogAt(t, tx, s.ID, branch.ID, emp.ID, "receive", 5, time.Now())

	status, body := slGet(t, slApp(tx, "admin", ""), "/stock-logs?stockId="+s.ID)

	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, true, body["success"])

	rows := body["data"].([]any)
	require.Len(t, rows, 1)
	row := rows[0].(map[string]any)
	assert.ElementsMatch(t,
		[]string{"id", "log_type", "qty_change", "note", "order_item_id", "usage_id", "created_at", "stock", "branch", "employee"},
		slKeys(row), "ต้องไม่มี stock_id/branch_id/employee_id ซ้ำที่ระดับบน")
	assert.ElementsMatch(t, []string{"id", "name"}, slKeys(row["stock"].(map[string]any)))
	assert.ElementsMatch(t, []string{"id", "code", "name"}, slKeys(row["branch"].(map[string]any)))
	assert.ElementsMatch(t, []string{"id", "name", "role"}, slKeys(row["employee"].(map[string]any)))
	assert.Equal(t, emp.Name, row["employee"].(map[string]any)["name"])
	assert.Equal(t, "manager", row["employee"].(map[string]any)["role"])

	meta := body["meta"].(map[string]any)
	assert.EqualValues(t, 1, meta["total"])
	assert.EqualValues(t, 1, meta["page"])
}

func TestStockLog_HTTP_QuantitiesJSONShape(t *testing.T) {
	tx := newTx(t)
	p := uniqueName("itest-h-qty")
	s := slStock(t, tx, p, 5)
	branch := mkBranch(t, tx)
	slLogAt(t, tx, s.ID, branch.ID, mkEmployee(t, tx, branch.ID, "manager").ID, "receive", 12, time.Now())

	status, body := slGet(t, slApp(tx, "admin", ""), "/stock-logs/quantities?branchId="+branch.ID+"&search="+p)

	require.Equal(t, http.StatusOK, status)
	rows := body["data"].([]any)
	require.Len(t, rows, 1)
	row := rows[0].(map[string]any)
	assert.ElementsMatch(t,
		[]string{"stock_id", "stock_name", "category", "uses_per_unit", "min_qty", "is_active", "branch_id", "branch_name", "quantity", "is_low_stock"},
		slKeys(row))
	assert.EqualValues(t, 12, row["quantity"])
	assert.Equal(t, false, row["is_low_stock"])
	assert.Equal(t, p, row["stock_name"])
}

func TestStockLog_HTTP_InvalidInputReturns400(t *testing.T) {
	tx := newTx(t)
	app := slApp(tx, "admin", "")

	for _, path := range []string{
		"/stock-logs?stockId=not-a-uuid",
		"/stock-logs?logType=steal",
		"/stock-logs?from=2026-10-02&to=2026-10-01",
		"/stock-logs?pageSize=101",
		"/stock-logs/quantities?category=luxury",
		"/stock-logs/quantities?active=maybe",
	} {
		t.Run(path, func(t *testing.T) {
			status, body := slGet(t, app, path)
			assert.Equal(t, http.StatusBadRequest, status)
			assert.Equal(t, false, body["success"])
		})
	}
}

func TestStockLog_HTTP_BranchUserWithoutBranchReturns403(t *testing.T) {
	tx := newTx(t)
	status, _ := slGet(t, slApp(tx, "branch", ""), "/stock-logs")
	assert.Equal(t, http.StatusForbidden, status)
}

// ---------- POST /stock-logs (receive / expired / damaged / internal) ----------

func slPost(t *testing.T, app *fiber.App, path string, payload map[string]any) (int, map[string]any) {
	t.Helper()
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var body map[string]any
	require.NoError(t, json.Unmarshal(out, &body), "response ต้องเป็น JSON: %s", out)
	return resp.StatusCode, body
}

func slMovement(stockID, branchID, employeeID, logType string, qty int) map[string]any {
	return map[string]any{"type": logType, "stockId": stockID, "branchId": branchID, "employeeId": employeeID, "qty": qty}
}

func slBalance(t *testing.T, tx *gorm.DB, stockID, branchID string) int {
	t.Helper()
	var sum int
	require.NoError(t, tx.Raw("SELECT COALESCE(SUM(qty_change), 0) FROM stock_logs WHERE stock_id = ? AND branch_id = ?", stockID, branchID).Scan(&sum).Error)
	return sum
}

func slCountLogs(t *testing.T, tx *gorm.DB, stockID string) int64 {
	t.Helper()
	var n int64
	require.NoError(t, tx.Model(&models.StockLog{}).Where("stock_id = ?", stockID).Count(&n).Error)
	return n
}

func TestStockLog_Add_SignAndPersistedRow(t *testing.T) {
	tests := []struct {
		logType string
		want    int
	}{
		{"receive", 7},
		{"expired", -7},
		{"damaged", -7},
		{"internal", -7},
	}
	for _, tt := range tests {
		t.Run(tt.logType, func(t *testing.T) {
			tx := newTx(t)
			svc := slSvc(tx)
			s := slStock(t, tx, uniqueName("itest-add-sign"), 0)
			branch := mkBranch(t, tx)
			emp := mkEmployee(t, tx, branch.ID, "manager")

			got, err := svc.AddStockLogMovement(ctx, "admin", "", stocklog.AddStockLogMovementRequest{
				Type: tt.logType, BranchID: branch.ID, StockID: s.ID, EmployeeID: emp.ID, Qty: 7,
			})
			require.NoError(t, err)
			assert.NotEmpty(t, got.ID, "GORM ต้องเติม id กลับมา")
			assert.False(t, got.CreatedAt.IsZero(), "GORM ต้องเติม created_at กลับมา")

			// อ่านจาก DB จริง ไม่ใช่เชื่อค่าที่คืนมา
			var saved models.StockLog
			require.NoError(t, tx.First(&saved, "id = ?", got.ID).Error)
			assert.Equal(t, tt.logType, saved.LogType)
			assert.Equal(t, tt.want, saved.QtyChange)
			assert.Equal(t, s.ID, saved.StockID)
			assert.Equal(t, branch.ID, saved.BranchID)
			assert.Equal(t, emp.ID, saved.EmployeeID)
			assert.Nil(t, saved.Note, "ไม่ส่ง note = NULL")
			assert.Nil(t, saved.OrderItemID)
			assert.Nil(t, saved.UsageID)
		})
	}
}

func TestStockLog_Add_NoteStoredTrimmedOrNull(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-add-note"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")

	withNote, err := svc.AddStockLogMovement(ctx, "admin", "", stocklog.AddStockLogMovementRequest{
		Type: "damaged", BranchID: branch.ID, StockID: s.ID, EmployeeID: emp.ID, Qty: 1, Note: "  ตกแตก  ",
	})
	require.NoError(t, err)
	blankNote, err := svc.AddStockLogMovement(ctx, "admin", "", stocklog.AddStockLogMovementRequest{
		Type: "damaged", BranchID: branch.ID, StockID: s.ID, EmployeeID: emp.ID, Qty: 1, Note: "   ",
	})
	require.NoError(t, err)

	var a, b models.StockLog
	require.NoError(t, tx.First(&a, "id = ?", withNote.ID).Error)
	require.NoError(t, tx.First(&b, "id = ?", blankNote.ID).Error)
	require.NotNil(t, a.Note)
	assert.Equal(t, "ตกแตก", *a.Note)
	assert.Nil(t, b.Note, "note ว่างต้องเป็น NULL ไม่ใช่ string ว่าง")
}

func TestStockLog_Add_AffectsQuantities(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-add-qty"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")

	for _, m := range []struct {
		logType string
		qty     int
	}{{"receive", 20}, {"expired", 3}, {"damaged", 2}, {"internal", 1}} {
		_, err := svc.AddStockLogMovement(ctx, "admin", "", stocklog.AddStockLogMovementRequest{
			Type: m.logType, BranchID: branch.ID, StockID: s.ID, EmployeeID: emp.ID, Qty: m.qty,
		})
		require.NoError(t, err)
	}

	items, _ := slQuantities(t, svc, stocklog.GetAllStockQuantitiesQuery{BranchID: branch.ID, Search: s.Name})
	require.Len(t, items, 1)
	assert.Equal(t, 20-3-2-1, items[0].Quantity, "ยอดคงเหลือ = SUM(qty_change)")

	logs, meta := slLogs(t, svc, stocklog.GetAllStockLogsQuery{StockID: s.ID, BranchID: branch.ID})
	assert.EqualValues(t, 4, meta.Total)
	assert.Len(t, logs, 4)
}

func TestStockLog_Add_BranchUserForcedToOwnBranch(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-add-branch"), 0)
	own := mkBranch(t, tx)
	other := mkBranch(t, tx)
	emp := mkEmployee(t, tx, own.ID, "manager")

	// บัญชีสาขาพยายามลงของเข้าสาขาอื่น ต้องถูกบังคับเป็นสาขาตัวเอง
	got, err := svc.AddStockLogMovement(ctx, "branch", own.ID, stocklog.AddStockLogMovementRequest{
		Type: "receive", BranchID: other.ID, StockID: s.ID, EmployeeID: emp.ID, Qty: 4,
	})
	require.NoError(t, err)
	assert.Equal(t, own.ID, got.BranchID)
	assert.Equal(t, 4, slBalance(t, tx, s.ID, own.ID))
	assert.Equal(t, 0, slBalance(t, tx, s.ID, other.ID), "สาขาอื่นต้องไม่ถูกแตะ")
}

func TestStockLog_Add_InvalidInputWritesNothing(t *testing.T) {
	tx := newTx(t)
	svc := slSvc(tx)
	s := slStock(t, tx, uniqueName("itest-add-invalid"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")

	base := func() stocklog.AddStockLogMovementRequest {
		return stocklog.AddStockLogMovementRequest{Type: "receive", BranchID: branch.ID, StockID: s.ID, EmployeeID: emp.ID, Qty: 5}
	}
	tests := []struct {
		name   string
		mutate func(r *stocklog.AddStockLogMovementRequest)
	}{
		{"type ไม่รู้จัก", func(r *stocklog.AddStockLogMovementRequest) { r.Type = "steal" }},
		{"type sale ห้ามในเส้นนี้", func(r *stocklog.AddStockLogMovementRequest) { r.Type = "sale" }},
		{"qty 0", func(r *stocklog.AddStockLogMovementRequest) { r.Qty = 0 }},
		{"qty ติดลบ", func(r *stocklog.AddStockLogMovementRequest) { r.Qty = -1 }},
		{"stockId ว่าง", func(r *stocklog.AddStockLogMovementRequest) { r.StockID = "" }},
		{"employeeId ไม่ใช่ UUID", func(r *stocklog.AddStockLogMovementRequest) { r.EmployeeID = "abc" }},
		{"branchId ว่าง (admin)", func(r *stocklog.AddStockLogMovementRequest) { r.BranchID = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := base()
			tt.mutate(&req)
			got, err := svc.AddStockLogMovement(ctx, "admin", "", req)

			assert.Nil(t, got)
			var appErr *_package.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, _package.CodeValidation, appErr.Code)
		})
	}
	assert.EqualValues(t, 0, slCountLogs(t, tx, s.ID), "validate ไม่ผ่านต้องไม่มีแถวถูกเขียน")
}

func TestStockLog_HTTP_AddReturns201WithModelShape(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-h-add"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")

	payload := slMovement(s.ID, branch.ID, emp.ID, "receive", 12)
	payload["note"] = "รับของจาก supplier"
	status, body := slPost(t, slApp(tx, "admin", ""), "/stock-logs", payload)

	require.Equal(t, http.StatusCreated, status)
	assert.Equal(t, true, body["success"])
	data := body["data"].(map[string]any)
	assert.NotEmpty(t, data["id"])
	assert.Equal(t, "receive", data["log_type"])
	assert.EqualValues(t, 12, data["qty_change"])
	assert.Equal(t, s.ID, data["stock_id"])
	assert.Equal(t, branch.ID, data["branch_id"])
	assert.Equal(t, emp.ID, data["employee_id"])
	assert.Equal(t, "รับของจาก supplier", data["note"])
	assert.NotEmpty(t, data["created_at"])

	// ยืนยันว่าเห็นในเส้น GET จริง
	status, listed := slGet(t, slApp(tx, "admin", ""), "/stock-logs?stockId="+s.ID)
	require.Equal(t, http.StatusOK, status)
	assert.Len(t, listed["data"].([]any), 1)
}

func TestStockLog_HTTP_AddInvalidInputReturns400(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-h-add-bad"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")
	app := slApp(tx, "admin", "")

	tests := []struct {
		name    string
		payload map[string]any
	}{
		{"type ผิด", slMovement(s.ID, branch.ID, emp.ID, "steal", 1)},
		{"qty 0", slMovement(s.ID, branch.ID, emp.ID, "receive", 0)},
		{"qty ติดลบ", slMovement(s.ID, branch.ID, emp.ID, "damaged", -3)},
		{"stockId ไม่ใช่ UUID", slMovement("abc", branch.ID, emp.ID, "receive", 1)},
		{"employeeId ว่าง", slMovement(s.ID, branch.ID, "", "receive", 1)},
		{"qty เป็น string (bind ไม่ได้)", map[string]any{"type": "receive", "stockId": s.ID, "branchId": branch.ID, "employeeId": emp.ID, "qty": "ten"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := slPost(t, app, "/stock-logs", tt.payload)
			assert.Equal(t, http.StatusBadRequest, status)
			assert.Equal(t, false, body["success"])
		})
	}
	assert.EqualValues(t, 0, slCountLogs(t, tx, s.ID), "request ที่ผิดต้องไม่มีแถวถูกเขียน")
}

func TestStockLog_HTTP_AddBranchUserWithoutBranchReturns403(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-h-add-403"), 0)
	status, _ := slPost(t, slApp(tx, "branch", ""), "/stock-logs", slMovement(s.ID, validUUID, validUUID, "receive", 1))
	assert.Equal(t, http.StatusForbidden, status)
}

func TestStockLog_HTTP_AddUnknownStockReturns400(t *testing.T) {
	tx := newTx(t)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")

	status, body := slPost(t, slApp(tx, "admin", ""), "/stock-logs", slMovement(validUUID, branch.ID, emp.ID, "receive", 1))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, false, body["success"])
}

func TestStockLog_HTTP_AddSoftDeletedStockReturns400(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-h-add-deleted"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")
	require.NoError(t, tx.Delete(&s).Error)

	status, _ := slPost(t, slApp(tx, "admin", ""), "/stock-logs", slMovement(s.ID, branch.ID, emp.ID, "receive", 1))
	assert.Equal(t, http.StatusBadRequest, status, "สินค้าที่ถูกลบไปแล้วห้ามลงสต็อก")
	assert.EqualValues(t, 0, slCountLogs(t, tx, s.ID))
}

func TestStockLog_HTTP_AddUnknownBranchReturns400(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-h-add-nobranch"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")

	status, _ := slPost(t, slApp(tx, "admin", ""), "/stock-logs", slMovement(s.ID, validUUID, emp.ID, "receive", 1))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.EqualValues(t, 0, slCountLogs(t, tx, s.ID))
}

func TestStockLog_HTTP_AddInactiveBranchReturns400(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-h-add-offbranch"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")
	slSetInactive(t, tx, &models.Branch{}, branch.ID)

	status, _ := slPost(t, slApp(tx, "admin", ""), "/stock-logs", slMovement(s.ID, branch.ID, emp.ID, "receive", 1))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.EqualValues(t, 0, slCountLogs(t, tx, s.ID))
}

func TestStockLog_HTTP_AddUnknownEmployeeReturns400(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-h-add-noemp"), 0)
	branch := mkBranch(t, tx)

	status, body := slPost(t, slApp(tx, "admin", ""), "/stock-logs", slMovement(s.ID, branch.ID, validUUID, "receive", 1))
	assert.Equal(t, http.StatusBadRequest, status, "ต้องไม่ใช่ 500 จาก FK")
	assert.Equal(t, false, body["success"])
	assert.EqualValues(t, 0, slCountLogs(t, tx, s.ID))
}

func TestStockLog_HTTP_AddInactiveEmployeeReturns400(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-h-add-offemp"), 0)
	branch := mkBranch(t, tx)
	emp := mkEmployee(t, tx, branch.ID, "manager")
	slSetInactive(t, tx, &models.Employee{}, emp.ID)

	status, _ := slPost(t, slApp(tx, "admin", ""), "/stock-logs", slMovement(s.ID, branch.ID, emp.ID, "receive", 1))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.EqualValues(t, 0, slCountLogs(t, tx, s.ID))
}

func TestStockLog_HTTP_AddEmployeeFromOtherBranchReturns400(t *testing.T) {
	tx := newTx(t)
	s := slStock(t, tx, uniqueName("itest-h-add-xbranch"), 0)
	mine, theirs := mkBranch(t, tx), mkBranch(t, tx)
	outsider := mkEmployee(t, tx, theirs.ID, "manager")

	// admin ลงสาขา mine แต่ใช้พนักงานของ theirs
	status, _ := slPost(t, slApp(tx, "admin", ""), "/stock-logs", slMovement(s.ID, mine.ID, outsider.ID, "receive", 1))
	assert.Equal(t, http.StatusBadRequest, status)

	// บัญชีสาขา mine พยายามอ้างพนักงานสาขาอื่น
	status, _ = slPost(t, slApp(tx, "branch", mine.ID), "/stock-logs", slMovement(s.ID, theirs.ID, outsider.ID, "receive", 1))
	assert.Equal(t, http.StatusBadRequest, status)

	assert.EqualValues(t, 0, slCountLogs(t, tx, s.ID))
}

// UUID รูปแบบถูกต้องแต่ไม่มีอยู่จริงใน DB
const validUUID = "8f64ee52-1111-4222-8333-944455556666"
