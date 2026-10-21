//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"backend/internal/models"
	"backend/internal/stock"
	_package "backend/package"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var ctx = context.Background()

// ---------- Repository: CRUD พื้นฐาน ----------

func TestRepo_CreateAndFind(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	name := uniqueName("itest-find")

	created, err := repo.CreateStock(ctx, models.Stock{
		Name: name, Price: decimal.NewFromInt(250), IsSellable: true, Category: "premium", UsesPerUnit: 2, IsActive: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID, "DB ต้องเจน uuid ให้")

	byID, err := repo.FindStockById(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, byID)
	assert.Equal(t, name, byID.Name)
	assert.True(t, byID.Price.Equal(decimal.NewFromInt(250)))

	byName, err := repo.FindStockByName(ctx, name)
	require.NoError(t, err)
	require.NotNil(t, byName)
	assert.Equal(t, created.ID, byName.ID)
}

func TestRepo_FindReturnsNilWhenMissing(t *testing.T) {
	repo := stock.NewRepository(newTx(t))

	byID, err := repo.FindStockById(ctx, "00000000-0000-4000-8000-000000000000")
	assert.NoError(t, err)
	assert.Nil(t, byID)

	byName, err := repo.FindStockByName(ctx, uniqueName("itest-missing"))
	assert.NoError(t, err)
	assert.Nil(t, byName)
}

func TestRepo_UpdateStock(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	s := mkStock(t, tx, uniqueName("itest-upd"))

	s.MinQty = 9
	s.Price = decimal.NewFromInt(777)
	_, err := repo.UpdateStock(ctx, s)
	require.NoError(t, err)

	got, err := repo.FindStockById(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, 9, got.MinQty)
	assert.True(t, got.Price.Equal(decimal.NewFromInt(777)))
}

func TestRepo_GetAllStock_SortedByNameAndExcludesDeleted(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	p := uniqueName("itest-all")
	b := mkStock(t, tx, p+"-b")
	a := mkStock(t, tx, p+"-a")
	gone := mkStock(t, tx, p+"-gone")
	require.NoError(t, repo.SoftDeleteStock(ctx, gone.ID))

	all, err := repo.GetAllStock(ctx)
	require.NoError(t, err)

	idx := map[string]int{}
	for i, s := range all {
		idx[s.ID] = i
	}
	require.Contains(t, idx, a.ID)
	require.Contains(t, idx, b.ID)
	assert.Less(t, idx[a.ID], idx[b.ID], "ต้องเรียงชื่อ ASC")
	assert.NotContains(t, idx, gone.ID, "ที่ soft delete แล้วต้องไม่โผล่")
}

// ---------- unique index ชื่อสินค้า (idx_stocks_name_active) ----------

func TestRepo_DuplicateActiveName_ViolatesUniqueIndex(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	name := uniqueName("itest-dup")
	mkStock(t, tx, name)

	err := expectDBError(t, tx, func() error {
		_, err := repo.CreateStock(ctx, models.Stock{Name: name, Category: "normal", UsesPerUnit: 1, IsActive: true})
		return err
	})

	// service อาศัยตัวนี้แปลงเป็น 409 ตอนสองคำขอชนกัน
	assert.ErrorIs(t, err, gorm.ErrDuplicatedKey)
}

func TestRepo_SoftDeletedName_CanBeReused(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	name := uniqueName("itest-reuse")
	old := mkStock(t, tx, name)
	require.NoError(t, repo.SoftDeleteStock(ctx, old.ID))

	again, err := repo.CreateStock(ctx, models.Stock{Name: name, Category: "normal", UsesPerUnit: 1, IsActive: true})

	require.NoError(t, err)
	assert.NotEqual(t, old.ID, again.ID)

	// FindStockByName ต้องได้ตัวที่ยังใช้งานก่อนตัวที่ถูกลบ
	found, err := repo.FindStockByName(ctx, name)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, again.ID, found.ID)
	assert.False(t, found.DeletedAt.Valid)
}

// ---------- CheckStockHasRelation ----------

func TestRepo_CheckStockHasRelation(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, tx *gorm.DB, stockID string)
		want  bool
	}{
		{"ไม่มีความสัมพันธ์", func(t *testing.T, tx *gorm.DB, id string) {}, false},
		{"ถูกขายใน order_items", func(t *testing.T, tx *gorm.DB, id string) { mkOrderItem(t, tx, id) }, true},
		{"อยู่ใน course_items", func(t *testing.T, tx *gorm.DB, id string) { mkCourseItem(t, tx, id) }, true},
		{"มี stock_logs", func(t *testing.T, tx *gorm.DB, id string) {
			mkStockLog(t, tx, id, mkBranch(t, tx).ID)
		}, true},
		{"มี stock_containers", func(t *testing.T, tx *gorm.DB, id string) {
			mkContainer(t, tx, id, mkBranch(t, tx).ID, time.Now(), false)
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := newTx(t)
			repo := stock.NewRepository(tx)
			s := mkStock(t, tx, uniqueName("itest-rel"))
			tt.setup(t, tx, s.ID)

			got, err := repo.CheckStockHasRelation(ctx, s.ID)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------- ลบ / กู้คืน ----------

func TestRepo_SoftDeleteAndFind(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	s := mkStock(t, tx, uniqueName("itest-soft"))

	require.NoError(t, repo.SoftDeleteStock(ctx, s.ID))

	normal, err := repo.FindStockById(ctx, s.ID)
	require.NoError(t, err)
	assert.Nil(t, normal, "FindStockById ต้องไม่เห็นของที่ลบแล้ว")

	incl, err := repo.FindStockByIdIncludingDeleted(ctx, s.ID)
	require.NoError(t, err)
	require.NotNil(t, incl)
	assert.True(t, incl.DeletedAt.Valid)
}

func TestRepo_HardDelete_RemovesRow(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	s := mkStock(t, tx, uniqueName("itest-hard"))

	require.NoError(t, repo.HardDeleteStock(ctx, s.ID))

	incl, err := repo.FindStockByIdIncludingDeleted(ctx, s.ID)
	require.NoError(t, err)
	assert.Nil(t, incl)
}

// เหตุผลที่ต้องมี CheckStockHasRelation: DB ไม่ยอมให้ลบจริงของที่ถูกอ้างอิง
func TestRepo_HardDelete_SoldStock_ViolatesForeignKey(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	s := mkStock(t, tx, uniqueName("itest-fk"))
	mkOrderItem(t, tx, s.ID)

	err := expectDBError(t, tx, func() error { return repo.HardDeleteStock(ctx, s.ID) })

	assert.ErrorIs(t, err, gorm.ErrForeignKeyViolated)
}

func TestRepo_RestoreStock(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	s := mkStock(t, tx, uniqueName("itest-restore"))
	require.NoError(t, repo.SoftDeleteStock(ctx, s.ID))

	restored, err := repo.RestoreStock(ctx, s.ID)

	require.NoError(t, err)
	assert.Equal(t, s.ID, restored.ID)
	assert.False(t, restored.DeletedAt.Valid)
	found, err := repo.FindStockById(ctx, s.ID)
	require.NoError(t, err)
	assert.NotNil(t, found, "กู้คืนแล้วต้องหาเจอตามปกติ")
}

func TestRepo_RestoreStock_NameTakenByActive_ViolatesUniqueIndex(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	name := uniqueName("itest-restore-dup")
	old := mkStock(t, tx, name)
	require.NoError(t, repo.SoftDeleteStock(ctx, old.ID))
	mkStock(t, tx, name) // มีตัวใหม่ใช้ชื่อนี้แล้ว

	err := expectDBError(t, tx, func() error {
		_, err := repo.RestoreStock(ctx, old.ID)
		return err
	})

	assert.ErrorIs(t, err, gorm.ErrDuplicatedKey)
}

// ---------- GetStockContainersByStockID ----------

func TestRepo_GetStockContainers(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	s := mkStock(t, tx, uniqueName("itest-cont"))
	other := mkStock(t, tx, uniqueName("itest-cont-other"))
	branchA, branchB := mkBranch(t, tx), mkBranch(t, tx)

	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	at := func(min int) time.Time { return base.Add(time.Duration(min) * time.Hour) }

	a0 := mkContainer(t, tx, s.ID, branchA.ID, at(0), false)
	a1 := mkContainer(t, tx, s.ID, branchA.ID, at(1), false)
	a2 := mkContainer(t, tx, s.ID, branchA.ID, at(2), true)
	b3 := mkContainer(t, tx, s.ID, branchB.ID, at(3), false)
	b4 := mkContainer(t, tx, s.ID, branchB.ID, at(4), true)
	mkContainer(t, tx, other.ID, branchA.ID, at(5), false) // ของสินค้าอื่น ห้ามปน

	ids := func(cs []models.StockContainer) []string {
		out := make([]string, len(cs))
		for i, c := range cs {
			out[i] = c.ID
		}
		return out
	}

	tests := []struct {
		name      string
		branch    string
		status    string
		page      int
		size      int
		want      []string
		wantTotal int64
	}{
		{"ทั้งหมด เรียงเปิดล่าสุดก่อน", "", stock.StatusAll, 1, 20, []string{b4.ID, b3.ID, a2.ID, a1.ID, a0.ID}, 5},
		{"เฉพาะที่ยังเปิด", "", stock.StatusOpen, 1, 20, []string{b3.ID, a1.ID, a0.ID}, 3},
		{"เฉพาะที่ปิดแล้ว", "", stock.StatusClosed, 1, 20, []string{b4.ID, a2.ID}, 2},
		{"กรองสาขา A", branchA.ID, stock.StatusAll, 1, 20, []string{a2.ID, a1.ID, a0.ID}, 3},
		{"สาขา A เฉพาะที่เปิด", branchA.ID, stock.StatusOpen, 1, 20, []string{a1.ID, a0.ID}, 2},
		{"หน้า 1 ขนาด 2", "", stock.StatusAll, 1, 2, []string{b4.ID, b3.ID}, 5},
		{"หน้า 2 ขนาด 2", "", stock.StatusAll, 2, 2, []string{a2.ID, a1.ID}, 5},
		{"หน้า 3 ขนาด 2 (เศษ)", "", stock.StatusAll, 3, 2, []string{a0.ID}, 5},
		{"เกินหน้าสุดท้าย ว่าง แต่ total ยังถูก", "", stock.StatusAll, 4, 2, []string{}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, total, err := repo.GetStockContainersByStockID(ctx, s.ID, tt.branch, tt.status, tt.page, tt.size)

			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)
			assert.Equal(t, tt.want, ids(got))
		})
	}

	t.Run("preload ข้อมูลสาขา", func(t *testing.T) {
		got, _, err := repo.GetStockContainersByStockID(ctx, s.ID, "", stock.StatusAll, 1, 1)
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.NotNil(t, got[0].Branch)
		assert.Equal(t, branchB.ID, got[0].Branch.ID)
	})

	t.Run("สินค้าไม่มีกระปุกเลย", func(t *testing.T) {
		got, total, err := repo.GetStockContainersByStockID(ctx, mkStock(t, tx, uniqueName("itest-empty")).ID, "", stock.StatusAll, 1, 20)
		require.NoError(t, err)
		assert.Empty(t, got)
		assert.Equal(t, int64(0), total)
	})
}

// opened_at เท่ากัน ต้องเรียงตาม id คงที่ ไม่งั้นข้ามหน้าแล้วแถวซ้ำ/ตกหล่น
func TestRepo_GetStockContainers_PaginationStableOnTies(t *testing.T) {
	tx := newTx(t)
	repo := stock.NewRepository(tx)
	s := mkStock(t, tx, uniqueName("itest-tie"))
	b := mkBranch(t, tx)
	same := time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC)
	want := map[string]bool{}
	for i := 0; i < 5; i++ {
		want[mkContainer(t, tx, s.ID, b.ID, same, false).ID] = true
	}

	seen := map[string]bool{}
	for page := 1; page <= 3; page++ {
		got, _, err := repo.GetStockContainersByStockID(ctx, s.ID, "", stock.StatusAll, page, 2)
		require.NoError(t, err)
		for _, c := range got {
			assert.False(t, seen[c.ID], "แถวซ้ำข้ามหน้า: %s", c.ID)
			seen[c.ID] = true
		}
	}
	assert.Equal(t, want, seen, "ต้องครบทุกแถว ไม่ตกหล่น")
}

// ---------- Service + Repository จริง ----------

func newService(tx *gorm.DB) stock.Service { return stock.NewService(stock.NewRepository(tx)) }

func addReq(name string) stock.AddStockRequest {
	return stock.AddStockRequest{
		StockName: name, Price: decimal.NewFromInt(100), IsSellable: true,
		Category: "normal", UsesPerUnit: 1, MinQty: 0,
	}
}

// บั๊กเดิม: ลบสินค้าที่เคยขายแล้วได้ 500 (FK) ตอนนี้ต้อง soft delete
func TestService_DeleteSoldStock_SoftDeletes(t *testing.T) {
	tx := newTx(t)
	svc := newService(tx)
	repo := stock.NewRepository(tx)
	s := mkStock(t, tx, uniqueName("itest-sold"))
	mkOrderItem(t, tx, s.ID)

	require.NoError(t, svc.DeleteStockById(ctx, s.ID))

	incl, err := repo.FindStockByIdIncludingDeleted(ctx, s.ID)
	require.NoError(t, err)
	require.NotNil(t, incl, "แถวต้องยังอยู่ (soft delete)")
	assert.True(t, incl.DeletedAt.Valid)
}

func TestService_DeleteUnreferencedStock_HardDeletes(t *testing.T) {
	tx := newTx(t)
	svc := newService(tx)
	repo := stock.NewRepository(tx)
	s := mkStock(t, tx, uniqueName("itest-unused"))

	require.NoError(t, svc.DeleteStockById(ctx, s.ID))

	incl, err := repo.FindStockByIdIncludingDeleted(ctx, s.ID)
	require.NoError(t, err)
	assert.Nil(t, incl, "ไม่มีใครอ้างอิง ต้องลบจริง")
}

// ลบทุกกรณีของความสัมพันธ์ ต้องไม่ error (ไม่หลุดเป็น 500)
func TestService_DeleteStockWithAnyRelation_DoesNotFail(t *testing.T) {
	setups := map[string]func(t *testing.T, tx *gorm.DB, id string){
		"order_items":      func(t *testing.T, tx *gorm.DB, id string) { mkOrderItem(t, tx, id) },
		"course_items":     func(t *testing.T, tx *gorm.DB, id string) { mkCourseItem(t, tx, id) },
		"stock_logs":       func(t *testing.T, tx *gorm.DB, id string) { mkStockLog(t, tx, id, mkBranch(t, tx).ID) },
		"stock_containers": func(t *testing.T, tx *gorm.DB, id string) { mkContainer(t, tx, id, mkBranch(t, tx).ID, time.Now(), false) },
	}

	for name, setup := range setups {
		t.Run(name, func(t *testing.T) {
			tx := newTx(t)
			s := mkStock(t, tx, uniqueName("itest-del-"+name))
			setup(t, tx, s.ID)

			require.NoError(t, newService(tx).DeleteStockById(ctx, s.ID))

			got, err := stock.NewRepository(tx).FindStockByIdIncludingDeleted(ctx, s.ID)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.True(t, got.DeletedAt.Valid)
		})
	}
}

func TestService_AddStock_DuplicateName_Returns409(t *testing.T) {
	tx := newTx(t)
	svc := newService(tx)
	name := uniqueName("itest-svc-dup")

	_, err := svc.AddStock(ctx, addReq(name))
	require.NoError(t, err)

	_, err = svc.AddStock(ctx, addReq(name))

	var appErr *_package.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, _package.CodeConflict, appErr.Code)
}

func TestService_AddStock_ReuseNameAfterSoftDelete(t *testing.T) {
	tx := newTx(t)
	svc := newService(tx)
	name := uniqueName("itest-svc-reuse")
	first := mkStock(t, tx, name)
	mkOrderItem(t, tx, first.ID) // ทำให้ลบแล้วเป็น soft delete
	require.NoError(t, svc.DeleteStockById(ctx, first.ID))

	again, err := svc.AddStock(ctx, addReq(name))

	require.NoError(t, err)
	assert.NotEqual(t, first.ID, again.ID)
}

func TestService_AddStock_ThaiNameOf150Chars_Persists(t *testing.T) {
	tx := newTx(t)
	svc := newService(tx)
	// "ก" 150 ตัวไม่ซ้ำกันไม่ได้ เลยใช้ตัวเลขท้ายชื่อแทน แล้วเติม "ก" ให้ครบ 150 ตัวอักษร
	suffix := tag()
	name := ""
	for i := 0; i < 150-len(suffix); i++ {
		name += "ก"
	}
	name += suffix

	got, err := svc.AddStock(ctx, addReq(name))

	require.NoError(t, err, "varchar(150) ต้องรับ 150 ตัวอักษรไทย (450 byte) ได้")
	found, err := stock.NewRepository(tx).FindStockById(ctx, got.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, name, found.Name)
}

func TestService_UpdateStock_PersistsAndKeepsUnsentFields(t *testing.T) {
	tx := newTx(t)
	svc := newService(tx)
	s := mkStock(t, tx, uniqueName("itest-svc-upd"))
	minQty := 12

	_, err := svc.UpdateStockById(ctx, s.ID, stock.UpdateStockRequest{MinQty: &minQty})
	require.NoError(t, err)

	got, err := stock.NewRepository(tx).FindStockById(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, 12, got.MinQty)
	assert.Equal(t, s.Name, got.Name)
	assert.True(t, got.Price.Equal(decimal.NewFromInt(100)), "ราคาเดิมต้องไม่ถูกเขียนทับ")
}

func TestService_UpdateStock_RenameToExistingName_Returns409(t *testing.T) {
	tx := newTx(t)
	svc := newService(tx)
	taken := uniqueName("itest-svc-taken")
	mkStock(t, tx, taken)
	s := mkStock(t, tx, uniqueName("itest-svc-other"))

	_, err := svc.UpdateStockById(ctx, s.ID, stock.UpdateStockRequest{StockName: &taken})

	var appErr *_package.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, _package.CodeConflict, appErr.Code)
}

func TestService_RestoreStock(t *testing.T) {
	t.Run("กู้คืนสำเร็จ", func(t *testing.T) {
		tx := newTx(t)
		svc := newService(tx)
		s := mkStock(t, tx, uniqueName("itest-svc-rest"))
		mkOrderItem(t, tx, s.ID)
		require.NoError(t, svc.DeleteStockById(ctx, s.ID))

		got, err := svc.RestoreStockById(ctx, s.ID)

		require.NoError(t, err)
		assert.False(t, got.DeletedAt.Valid)
	})
	t.Run("ชื่อถูกใช้ไปแล้ว 409", func(t *testing.T) {
		tx := newTx(t)
		svc := newService(tx)
		name := uniqueName("itest-svc-rest-dup")
		old := mkStock(t, tx, name)
		mkOrderItem(t, tx, old.ID)
		require.NoError(t, svc.DeleteStockById(ctx, old.ID))
		mkStock(t, tx, name)

		_, err := svc.RestoreStockById(ctx, old.ID)

		var appErr *_package.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, _package.CodeConflict, appErr.Code)
	})
}

func TestService_GetStockContainers_BranchUserSeesOnlyOwnBranch(t *testing.T) {
	tx := newTx(t)
	svc := newService(tx)
	s := mkStock(t, tx, uniqueName("itest-svc-scope"))
	mine, theirs := mkBranch(t, tx), mkBranch(t, tx)
	now := time.Now()
	own := mkContainer(t, tx, s.ID, mine.ID, now, false)
	mkContainer(t, tx, s.ID, theirs.ID, now.Add(time.Hour), false)

	// บัญชีสาขาพยายามขอดูสาขาอื่น ต้องถูกบังคับเป็นสาขาตัวเอง
	got, meta, err := svc.GetStockContainersByStockID(ctx, "branch", mine.ID,
		stock.ContainerQuery{StockID: s.ID, BranchID: theirs.ID, Status: stock.StatusAll})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, own.ID, got[0].ID)
	assert.Equal(t, int64(1), meta.Total)

	// admin ไม่ระบุสาขา เห็นทุกสาขา
	all, meta, err := svc.GetStockContainersByStockID(ctx, "admin", "",
		stock.ContainerQuery{StockID: s.ID, Status: stock.StatusAll})
	require.NoError(t, err)
	assert.Len(t, all, 2)
	assert.Equal(t, int64(2), meta.Total)
}
