//go:build integration

package integration

import (
	"errors"
	"testing"
	"time"

	"backend/internal/course"
	"backend/internal/models"
	_package "backend/package"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// repository ของ course ทดสอบบน DB จริง (query, partial unique index, FK)
// ทุก test ทำงานใน transaction แล้ว rollback ข้อมูล seed ที่มีอยู่จึงไม่ถูกแตะ

// ---------- helpers ----------

func mkCourse(t *testing.T, tx *gorm.DB, name string) models.Course {
	t.Helper()
	c := models.Course{Name: name, Price: decimal.NewFromInt(1000), MaxSession: 5, IsActive: true}
	require.NoError(t, tx.Create(&c).Error)
	return c
}

func softDeleteCourse(t *testing.T, tx *gorm.DB, id string) {
	t.Helper()
	require.NoError(t, tx.Delete(&models.Course{}, "id = ?", id).Error)
}

// mkCourseOrderItem สร้างบิล 1 ใบที่ขายคอร์สนี้ 1 บรรทัด
func mkCourseOrderItem(t *testing.T, tx *gorm.DB, courseID string) models.OrderItem {
	t.Helper()
	branch := mkBranch(t, tx)
	customer := models.Customer{Fname: "itest", Tier: "general", HnNo: "T" + tag()}
	require.NoError(t, tx.Create(&customer).Error)
	order := models.Order{OrderNo: "T" + tag(), CustomerID: customer.ID, BranchID: branch.ID, OrderDate: time.Now()}
	require.NoError(t, tx.Create(&order).Error)

	item := models.OrderItem{
		OrderID:   order.ID,
		LineNo:    1,
		ItemType:  "course",
		CourseID:  &courseID,
		Qty:       1,
		UnitPrice: decimal.NewFromInt(1000),
	}
	require.NoError(t, tx.Create(&item).Error)
	return item
}

// mkVIPUsage สร้างการใช้บริการ VIP ที่อ้างคอร์สนี้ (ตัดจากรายการเครดิตของบิลอื่น ไม่ได้ขายคอร์สนี้)
func mkVIPUsage(t *testing.T, tx *gorm.DB, courseID string) models.ServiceUsage {
	t.Helper()
	item := mkOrderItem(t, tx, mkStock(t, tx, uniqueName("itest-vip-stock")).ID)
	var order models.Order
	require.NoError(t, tx.First(&order, "id = ?", item.OrderID).Error)

	u := models.ServiceUsage{
		OrderItemID:  item.ID,
		CourseID:     &courseID,
		PriceCharged: decimal.NewNullDecimal(decimal.NewFromInt(500)),
		BranchID:     order.BranchID,
	}
	require.NoError(t, tx.Create(&u).Error)
	return u
}

func countCourse(t *testing.T, tx *gorm.DB, id string, unscoped bool) int64 {
	t.Helper()
	q := tx.Model(&models.Course{})
	if unscoped {
		q = q.Unscoped()
	}
	var n int64
	require.NoError(t, q.Where("id = ?", id).Count(&n).Error)
	return n
}

func courseErrCode(err error) string {
	var appErr *_package.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}

// ---------- Repository: สร้าง / ค้นหา ----------

func TestCourseRepo_CreateAndFind(t *testing.T) {
	tx := newTx(t)
	repo := course.NewRepository(tx)
	name := uniqueName("itest-course")
	validDay := 90

	created, err := repo.CreateCourse(ctx, models.Course{
		Name: name, Price: decimal.NewFromInt(3500), MaxSession: 10, ValidDay: &validDay,
		ServiceComm: decimal.NewFromInt(150), Category: "facial", IsActive: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID, "DB ต้องเจน uuid ให้")

	byID, err := repo.FindCourseById(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, byID)
	assert.Equal(t, name, byID.Name)
	assert.True(t, byID.Price.Equal(decimal.NewFromInt(3500)))
	require.NotNil(t, byID.ValidDay)
	assert.Equal(t, 90, *byID.ValidDay)

	byName, err := repo.FindCourseByName(ctx, name)
	require.NoError(t, err)
	require.NotNil(t, byName)
	assert.Equal(t, created.ID, byName.ID)
}

func TestCourseRepo_FindReturnsNilWhenMissing(t *testing.T) {
	repo := course.NewRepository(newTx(t))

	byID, err := repo.FindCourseById(ctx, "00000000-0000-4000-8000-000000000000")
	assert.NoError(t, err)
	assert.Nil(t, byID)

	byName, err := repo.FindCourseByName(ctx, uniqueName("itest-missing"))
	assert.NoError(t, err)
	assert.Nil(t, byName)
}

// FindCourseById ใช้กับ update / delete: คอร์สที่ถูก soft delete ต้องหาไม่เจอ (→ 404)
func TestCourseRepo_FindByIdExcludesSoftDeleted(t *testing.T) {
	tx := newTx(t)
	repo := course.NewRepository(tx)
	c := mkCourse(t, tx, uniqueName("itest-del"))
	softDeleteCourse(t, tx, c.ID)

	got, err := repo.FindCourseById(ctx, c.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

// FindCourseByName รวมตัวที่ถูกลบ แต่ถ้ามีทั้งตัวที่ลบแล้วและตัวที่ยังอยู่ ต้องคืนตัวที่ยังอยู่ก่อน
func TestCourseRepo_FindByName_PrefersActiveOverSoftDeleted(t *testing.T) {
	tx := newTx(t)
	repo := course.NewRepository(tx)
	name := uniqueName("itest-same")

	old := mkCourse(t, tx, name)
	softDeleteCourse(t, tx, old.ID)

	onlyDeleted, err := repo.FindCourseByName(ctx, name)
	require.NoError(t, err)
	require.NotNil(t, onlyDeleted, "มีแต่ตัวที่ลบแล้วก็ต้องเจอ")
	assert.Equal(t, old.ID, onlyDeleted.ID)
	assert.True(t, onlyDeleted.DeletedAt.Valid)

	active := mkCourse(t, tx, name)

	got, err := repo.FindCourseByName(ctx, name)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, active.ID, got.ID)
	assert.False(t, got.DeletedAt.Valid)
}

// ---------- Repository: unique index (เฉพาะที่ยังไม่ถูกลบ) ----------

func TestCourseRepo_NameUniqueOnlyAmongActive(t *testing.T) {
	tx := newTx(t)
	repo := course.NewRepository(tx)
	name := uniqueName("itest-uniq")
	first := mkCourse(t, tx, name)

	err := expectDBError(t, tx, func() error {
		_, err := repo.CreateCourse(ctx, models.Course{Name: name, Price: decimal.NewFromInt(1), MaxSession: 1})
		return err
	})
	assert.ErrorIs(t, err, gorm.ErrDuplicatedKey, "ชื่อซ้ำกับคอร์สที่ยังอยู่ DB ต้องกัน")

	softDeleteCourse(t, tx, first.ID)

	_, err = repo.CreateCourse(ctx, models.Course{Name: name, Price: decimal.NewFromInt(1), MaxSession: 1})
	assert.NoError(t, err, "ลบไปแล้ว ต้องสร้างชื่อเดิมได้")
}

func TestCourseRepo_UpdateToActiveNameIsRejectedByDB(t *testing.T) {
	tx := newTx(t)
	repo := course.NewRepository(tx)
	a := mkCourse(t, tx, uniqueName("itest-a"))
	b := mkCourse(t, tx, uniqueName("itest-b"))

	b.Name = a.Name
	err := expectDBError(t, tx, func() error {
		_, err := repo.UpdateCourse(ctx, b)
		return err
	})
	assert.ErrorIs(t, err, gorm.ErrDuplicatedKey)
}

// ---------- Repository: GetAllCourse ----------

// เดินทีละหน้า (หน้าละ 2) ต้องได้ครบทุกคอร์สที่ยังไม่ถูกลบ ไม่ซ้ำ ไม่หาย
// ชื่อซ้ำกันไม่ได้ จึงใช้คอร์สที่ชื่อติดกันแทนการทดสอบ id เป็นตัวตัดสิน
func TestCourseRepo_GetAllCourse_PagesWithoutDuplicatesAndExcludesDeleted(t *testing.T) {
	tx := newTx(t)
	repo := course.NewRepository(tx)
	prefix := uniqueName("itest-page")
	mine := map[string]bool{}
	for _, suffix := range []string{"-c", "-a", "-b"} {
		mine[mkCourse(t, tx, prefix+suffix).ID] = true
	}
	deleted := mkCourse(t, tx, prefix+"-deleted")
	softDeleteCourse(t, tx, deleted.ID)

	var wantTotal int64
	require.NoError(t, tx.Model(&models.Course{}).Count(&wantTotal).Error)

	seen := map[string]bool{}
	var names []string
	pages := int(wantTotal+1) / 2
	for page := 1; page <= pages; page++ {
		rows, total, err := repo.GetAllCourse(ctx, "", page, 2)
		require.NoError(t, err)
		assert.Equal(t, wantTotal, total, "total ต้องนับทุกคอร์สที่ยังไม่ถูกลบ ไม่ใช่แค่หน้านี้")
		for _, r := range rows {
			assert.False(t, seen[r.ID], "คอร์ส %s โผล่ซ้ำข้ามหน้า", r.ID)
			seen[r.ID] = true
			if mine[r.ID] {
				names = append(names, r.Name)
			}
		}
	}

	assert.Len(t, seen, int(wantTotal), "ต้องเห็นครบทุกคอร์ส")
	assert.False(t, seen[deleted.ID], "คอร์สที่ถูก soft delete ต้องไม่โผล่")
	assert.Equal(t, []string{prefix + "-a", prefix + "-b", prefix + "-c"}, names, "ต้องเรียงตามชื่อ")

	rows, _, err := repo.GetAllCourse(ctx, "", pages+1, 2)
	require.NoError(t, err)
	assert.Empty(t, rows, "เกินหน้าสุดท้าย = ว่าง ไม่ error")
}

// ---------- Repository: search ----------

func courseNames(rows []models.Course) []string {
	names := make([]string, 0, len(rows))
	for _, r := range rows {
		names = append(names, r.Name)
	}
	return names
}

func TestCourseRepo_GetAllCourse_Search(t *testing.T) {
	tx := newTx(t)
	repo := course.NewRepository(tx)
	p := uniqueName("itest-srch")
	// ชื่อที่จะเทียบลำดับใช้ตัวพิมพ์เดียวกัน เพราะลำดับ ตัวใหญ่/ตัวเล็ก ขึ้นกับ collation ของ DB
	mkCourse(t, tx, p+"-facial gold")
	mkCourse(t, tx, p+"-facial basic")
	mkCourse(t, tx, p+"-body")
	gone := mkCourse(t, tx, p+"-facial old")
	softDeleteCourse(t, tx, gone.ID)

	rows, total, err := repo.GetAllCourse(ctx, p+"-FACIAL", 1, 20)
	require.NoError(t, err)
	assert.EqualValues(t, 2, total, "total ต้องนับเฉพาะที่ตรงคำค้น และไม่นับตัวที่ถูกลบ")
	assert.Equal(t, []string{p + "-facial basic", p + "-facial gold"}, courseNames(rows),
		"ไม่สนตัวพิมพ์เล็กใหญ่ และยังเรียงตามชื่อ")

	rows, total, err = repo.GetAllCourse(ctx, p+"-facial", 2, 1)
	require.NoError(t, err)
	assert.EqualValues(t, 2, total, "แบ่งหน้าแล้ว total ยังเป็นจำนวนที่ตรงคำค้นทั้งหมด")
	assert.Equal(t, []string{p + "-facial gold"}, courseNames(rows))

	rows, total, err = repo.GetAllCourse(ctx, uniqueName("itest-nomatch"), 1, 20)
	require.NoError(t, err)
	assert.EqualValues(t, 0, total)
	assert.Empty(t, rows)
}

// % และ _ ที่พิมพ์มาต้องค้นเป็นตัวอักษรจริง ไม่ใช่ wildcard
func TestCourseRepo_GetAllCourse_SearchEscapesWildcards(t *testing.T) {
	tx := newTx(t)
	repo := course.NewRepository(tx)
	p := uniqueName("itest-wild")
	mkCourse(t, tx, p+"-ลด 50% ทั้งคอร์ส")
	mkCourse(t, tx, p+"-ลด 500 บาท")
	mkCourse(t, tx, p+"-a_b")
	mkCourse(t, tx, p+"-axb")

	rows, _, err := repo.GetAllCourse(ctx, p+"-ลด 50%", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, []string{p + "-ลด 50% ทั้งคอร์ส"}, courseNames(rows))

	rows, _, err = repo.GetAllCourse(ctx, p+"-a_b", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, []string{p + "-a_b"}, courseNames(rows))
}

// ---------- Repository: UpdateCourse ----------

func TestCourseRepo_UpdateCourse(t *testing.T) {
	tx := newTx(t)
	repo := course.NewRepository(tx)
	c := mkCourse(t, tx, uniqueName("itest-upd"))
	validDay := 30
	c.ValidDay = &validDay

	c.Price = decimal.NewFromInt(777)
	c.MaxSession = 9
	c.IsActive = false
	_, err := repo.UpdateCourse(ctx, c)
	require.NoError(t, err)

	got, err := repo.FindCourseById(ctx, c.ID)
	require.NoError(t, err)
	assert.True(t, got.Price.Equal(decimal.NewFromInt(777)))
	assert.Equal(t, 9, got.MaxSession)
	assert.False(t, got.IsActive)
	require.NotNil(t, got.ValidDay)

	// ล้าง valid_day เป็น NULL (ไม่หมดอายุ) ต้องบันทึกได้จริง
	got.ValidDay = nil
	_, err = repo.UpdateCourse(ctx, *got)
	require.NoError(t, err)
	again, err := repo.FindCourseById(ctx, c.ID)
	require.NoError(t, err)
	assert.Nil(t, again.ValidDay)
}

// ---------- Repository: Relation / Delete ----------

func TestCourseRepo_CheckCourseHasRelation(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, tx *gorm.DB, courseID string)
		want  bool
	}{
		{"ไม่มีใครอ้างถึง", func(t *testing.T, tx *gorm.DB, courseID string) {}, false},
		{"course_items", func(t *testing.T, tx *gorm.DB, courseID string) {
			stock := mkStock(t, tx, uniqueName("itest-ci-stock"))
			require.NoError(t, tx.Create(&models.CourseItem{CourseID: courseID, StockID: stock.ID, UseQty: 1}).Error)
		}, true},
		{"order_items", func(t *testing.T, tx *gorm.DB, courseID string) { mkCourseOrderItem(t, tx, courseID) }, true},
		{"service_usages", func(t *testing.T, tx *gorm.DB, courseID string) { mkVIPUsage(t, tx, courseID) }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := newTx(t)
			repo := course.NewRepository(tx)
			c := mkCourse(t, tx, uniqueName("itest-rel"))
			other := mkCourse(t, tx, uniqueName("itest-other"))
			tt.setup(t, tx, c.ID)

			got, err := repo.CheckCourseHasRelation(ctx, c.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			gotOther, err := repo.CheckCourseHasRelation(ctx, other.ID)
			require.NoError(t, err)
			assert.False(t, gotOther, "relation ของคอร์สอื่นต้องไม่นับรวม")
		})
	}
}

// ทุกตารางที่ CheckCourseHasRelation เช็คต้องมี FK จริง (hard delete โดนกัน)
// service พึ่งพา ErrForeignKeyViolated เพื่อถอยไป soft delete ตอนมีคนขายพอดี
func TestCourseRepo_HardDeleteBlockedByFK(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, tx *gorm.DB, courseID string)
	}{
		{"course_items", func(t *testing.T, tx *gorm.DB, courseID string) {
			stock := mkStock(t, tx, uniqueName("itest-fk-stock"))
			require.NoError(t, tx.Create(&models.CourseItem{CourseID: courseID, StockID: stock.ID, UseQty: 1}).Error)
		}},
		{"order_items", func(t *testing.T, tx *gorm.DB, courseID string) { mkCourseOrderItem(t, tx, courseID) }},
		{"service_usages", func(t *testing.T, tx *gorm.DB, courseID string) { mkVIPUsage(t, tx, courseID) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := newTx(t)
			repo := course.NewRepository(tx)
			c := mkCourse(t, tx, uniqueName("itest-fk"))
			tt.setup(t, tx, c.ID)

			err := expectDBError(t, tx, func() error { return repo.HardDeleteCourse(ctx, c.ID) })
			assert.ErrorIs(t, err, gorm.ErrForeignKeyViolated)
			assert.EqualValues(t, 1, countCourse(t, tx, c.ID, false), "ต้องยังอยู่")
		})
	}
}

func TestCourseRepo_SoftAndHardDelete(t *testing.T) {
	tx := newTx(t)
	repo := course.NewRepository(tx)
	soft := mkCourse(t, tx, uniqueName("itest-soft"))
	hard := mkCourse(t, tx, uniqueName("itest-hard"))

	require.NoError(t, repo.SoftDeleteCourse(ctx, soft.ID))
	assert.EqualValues(t, 0, countCourse(t, tx, soft.ID, false), "soft delete ต้องหายจาก query ปกติ")
	assert.EqualValues(t, 1, countCourse(t, tx, soft.ID, true), "แต่แถวยังอยู่ในตาราง")

	require.NoError(t, repo.HardDeleteCourse(ctx, hard.ID))
	assert.EqualValues(t, 0, countCourse(t, tx, hard.ID, true), "hard delete ต้องหายจริง")
}

// ---------- Service + Repository จริง ----------

func TestCourseService_Delete_HardWithoutRelation_SoftWithRelation(t *testing.T) {
	tx := newTx(t)
	svc := course.NewService(course.NewRepository(tx))

	free := mkCourse(t, tx, uniqueName("itest-free"))
	sold := mkCourse(t, tx, uniqueName("itest-sold"))
	item := mkCourseOrderItem(t, tx, sold.ID)

	require.NoError(t, svc.DeleteCourseById(ctx, free.ID))
	assert.EqualValues(t, 0, countCourse(t, tx, free.ID, true), "ไม่มี relation = ลบจริง")

	require.NoError(t, svc.DeleteCourseById(ctx, sold.ID))
	assert.EqualValues(t, 0, countCourse(t, tx, sold.ID, false))
	assert.EqualValues(t, 1, countCourse(t, tx, sold.ID, true), "มี relation = soft delete")

	var stillThere models.OrderItem
	require.NoError(t, tx.First(&stillThere, "id = ?", item.ID).Error)
	assert.Equal(t, sold.ID, *stillThere.CourseID, "ประวัติบิลต้องยังอ้างคอร์สเดิมได้")

	err := svc.DeleteCourseById(ctx, sold.ID)
	assert.Equal(t, _package.CodeNotFound, courseErrCode(err), "ลบซ้ำ = 404")
}

func TestCourseService_Update_NameConflictRespectsSoftDelete(t *testing.T) {
	tx := newTx(t)
	svc := course.NewService(course.NewRepository(tx))
	target := mkCourse(t, tx, uniqueName("itest-target"))
	active := mkCourse(t, tx, uniqueName("itest-active"))
	gone := mkCourse(t, tx, uniqueName("itest-gone"))
	softDeleteCourse(t, tx, gone.ID)

	_, err := svc.UpdateCourseById(ctx, target.ID, course.UpdateCourseRequest{Name: &active.Name})
	assert.Equal(t, _package.CodeConflict, courseErrCode(err), "ชื่อซ้ำกับคอร์สที่ยังอยู่ = 409")

	got, err := svc.UpdateCourseById(ctx, target.ID, course.UpdateCourseRequest{Name: &gone.Name})
	require.NoError(t, err, "ชื่อซ้ำกับคอร์สที่ถูกลบแล้ว ต้องใช้ได้")
	assert.Equal(t, gone.Name, got.Name)

	_, err = svc.UpdateCourseById(ctx, gone.ID, course.UpdateCourseRequest{Name: &gone.Name})
	assert.Equal(t, _package.CodeNotFound, courseErrCode(err), "แก้คอร์สที่ถูกลบแล้ว = 404")
}

func TestCourseService_Add_AfterSoftDeleteSameNameAllowed(t *testing.T) {
	tx := newTx(t)
	svc := course.NewService(course.NewRepository(tx))
	old := mkCourse(t, tx, uniqueName("itest-readd"))

	_, err := svc.AddCourse(ctx, course.AddCourseRequest{Name: old.Name, Price: decimal.NewFromInt(1), MaxSession: 1})
	assert.Equal(t, _package.CodeConflict, courseErrCode(err))

	softDeleteCourse(t, tx, old.ID)

	created, err := svc.AddCourse(ctx, course.AddCourseRequest{Name: old.Name, Price: decimal.NewFromInt(1), MaxSession: 1})
	require.NoError(t, err)
	assert.NotEqual(t, old.ID, created.ID)
}
