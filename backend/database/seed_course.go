package database

import (
	"fmt"

	"backend/internal/models"

	"gorm.io/gorm"
)

// ชื่อคอร์สตัวอย่าง (ใช้ร่วมกันระหว่าง SeedCourses กับ SeedCourseItems)
const (
	courseClearFace     = "คอร์สหน้าใส 10 ครั้ง"
	courseFaceMassage   = "คอร์สนวดหน้า 5 ครั้ง"
	courseBodyMassage   = "คอร์สนวดตัว 10 ครั้ง"
	courseFaceTreatment = "ทรีตเมนต์หน้า 1 ครั้ง"
)

// seedCourseItem = ของที่คอร์สใช้ต่อการทำ 1 ครั้ง → course_items
type seedCourseItem struct {
	course string // ชื่อคอร์ส
	stock  string // ชื่อสินค้า
	useQty int    // ดูความหมายที่ models.CourseItem
}

// ผูกเฉพาะของใช้ในร้าน (uses_per_unit > 1) เพราะของขาย uses_per_unit = 1 จะถูกตัดทั้งชิ้นทุกครั้งที่ทำ
// คอร์สที่ใช้มาส์กจะตัดสต็อกที่ S1 ไม่ได้ (S1 ไม่มีมาส์ก ดู seedReceives) ใช้ทดสอบกรณีของไม่พอ
var seedCourseItems = []seedCourseItem{
	{courseClearFace, stockGel, 1}, {courseClearFace, stockMask, 1},
	{courseFaceMassage, stockGel, 2},
	{courseBodyMassage, stockOil, 1},
	{courseFaceTreatment, stockGel, 1}, {courseFaceTreatment, stockMask, 1},
}

// SeedCourseItems ผูกของที่ใช้ให้คอร์สตัวอย่าง ในทรานแซกชันเดียว รันซ้ำได้ (มีคู่คอร์ส-สินค้านี้อยู่แล้วจะข้าม
// และไม่แก้ use_qty เดิม) ต้องรันหลัง SeedCourses และ SeedStocks
//
// หาคอร์สและสินค้าจากชื่อ (ไม่นับแถวที่ถูกลบ) ถ้าไม่เจอหรือถูกปิดใช้งานจะ error แทนการข้าม
func SeedCourseItems(db *gorm.DB) error {
	courseByName, err := loadCoursesByName(db)
	if err != nil {
		return err
	}
	stockByName, err := loadStocksByName(db)
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, ci := range seedCourseItems {
			where := fmt.Sprintf("seed course item (%s / %s)", ci.course, ci.stock)

			course, ok := courseByName[ci.course]
			if !ok {
				return fmt.Errorf("%s: course not found (run SeedCourses first)", where)
			}
			stock, ok := stockByName[ci.stock]
			if !ok {
				return fmt.Errorf("%s: stock not found (run SeedStocks first)", where)
			}
			if !course.IsActive || !stock.IsActive {
				return fmt.Errorf("%s: course and stock must be active", where)
			}
			if ci.useQty <= 0 || ci.useQty > stock.UsesPerUnit {
				return fmt.Errorf("%s: useQty must be between 1 and %d", where, stock.UsesPerUnit)
			}

			item := models.CourseItem{CourseID: course.ID, StockID: stock.ID, UseQty: ci.useQty}
			err := tx.Where(models.CourseItem{CourseID: course.ID, StockID: stock.ID}).FirstOrCreate(&item).Error
			if err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
		}
		return nil
	})
}

func loadCoursesByName(db *gorm.DB) (map[string]models.Course, error) {
	var courses []models.Course
	if err := db.Find(&courses).Error; err != nil {
		return nil, err
	}
	m := make(map[string]models.Course, len(courses))
	for _, c := range courses {
		m[c.Name] = c
	}
	return m, nil
}
