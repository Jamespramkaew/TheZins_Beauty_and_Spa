package course

import (
	"context"
	"strings"

	"backend/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetAllCourse(ctx context.Context, search string, page, pageSize int) ([]models.Course, int64, error)
	FindCourseByName(ctx context.Context, name string) (*models.Course, error)
	CreateCourse(ctx context.Context, course models.Course, items []models.CourseItem) (models.Course, error)
	FindStocksByIds(ctx context.Context, ids []string) ([]models.Stock, error)
	FindCourseById(ctx context.Context, id string) (*models.Course, error)
	UpdateCourse(ctx context.Context, course models.Course, items *[]models.CourseItem) (models.Course, error)
	FindCourseItems(ctx context.Context, courseId string) ([]models.CourseItem, error)
	CheckCourseHasRelation(ctx context.Context, courseId string) (bool, error)
	SoftDeleteCourse(ctx context.Context, courseId string) error
	HardDeleteCourse(ctx context.Context, courseId string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// GetAllCourse คืนคอร์สทีละหน้า เรียงตามชื่อ พร้อมของที่คอร์สใช้ (items + stock) · search ว่าง = ทุกคอร์ส
// คืนจำนวนที่ตรงเงื่อนไขทั้งหมด (total) ไว้ทำ meta
func (r *repository) GetAllCourse(ctx context.Context, search string, page, pageSize int) ([]models.Course, int64, error) {
	all := r.db.WithContext(ctx).Model(&models.Course{})
	if search != "" {
		all = all.Where("name ILIKE ?", "%"+escapeLike(search)+"%")
	}
	// Session ทำให้ใช้ query เดียวกันซ้ำได้ทั้งตอน Count และ Find
	all = all.Session(&gorm.Session{})

	var total int64
	if err := all.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var courses []models.Course
	// id เป็นตัวตัดสินเมื่อชื่อเท่ากัน ไม่งั้นข้ามหน้าแล้วแถวอาจซ้ำหรือตกหล่น
	err := all.
		Scopes(preloadItems).
		Order("name ASC, id ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&courses).Error
	if err != nil {
		return nil, 0, err
	}
	return courses, total, nil
}

// preloadItems โหลดของที่คอร์สใช้ (items + stock) ใช้ร่วมกันทุกที่ที่คืน CourseResponse ลำดับจะได้ตรงกัน
// items เรียงตามชื่อสินค้าด้วย collation ของ DB (stock_id ตัดสินเมื่อชื่อเท่ากัน)
// stock เป็น Unscoped: สินค้าที่ผูกกับคอร์สลบได้แค่ soft delete ต้องยังเห็นชื่อ · response ใช้แค่ id กับชื่อ
func preloadItems(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("course_items.*").
				Joins("JOIN stocks ON stocks.id = course_items.stock_id").
				Order("stocks.name ASC, course_items.stock_id ASC")
		}).
		Preload("Items.Stock", func(tx *gorm.DB) *gorm.DB { return tx.Unscoped().Select("id", "name") })
}

// escapeLike กัน % และ _ ที่ user พิมพ์มาถูกตีความเป็น wildcard ของ ILIKE
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// FindCourseByName ค้นหาคอร์สตามชื่อ (รวมทั้งที่ถูกลบด้วย soft delete)
func (r *repository) FindCourseByName(ctx context.Context, name string) (*models.Course, error) {
	var course models.Course
	// เรียงตัวที่ยังไม่ถูกลบขึ้นก่อน
	err := r.db.WithContext(ctx).Unscoped().Where("name = ?", name).
		Order("(deleted_at IS NULL) DESC").First(&course).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &course, nil
}

// CreateCourse สร้างคอร์สใหม่พร้อมของที่ใช้ (course_id ใน items ใส่ให้เอง) ในทรานแซกชันเดียว
// ถ้าบันทึก item ไม่ผ่าน คอร์สจะไม่ถูกสร้าง · คืนคอร์สที่อ่านกลับจาก DB พร้อม items แบบเดียวกับ GetAllCourse
func (r *repository) CreateCourse(ctx context.Context, course models.Course, items []models.CourseItem) (models.Course, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Omit กัน GORM บันทึก association ใน course.Items เอง
		if err := tx.Omit("Items").Create(&course).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].CourseID = course.ID
		}
		return tx.Omit("Course", "Stock").Create(&items).Error
	})
	if err != nil {
		return models.Course{}, err
	}

	var created models.Course
	if err := r.db.WithContext(ctx).Scopes(preloadItems).First(&created, "id = ?", course.ID).Error; err != nil {
		return models.Course{}, err
	}
	return created, nil
}

// FindStocksByIds ค้นหาสินค้าตาม id (รวมที่ถูก soft delete ให้ service ตัดสินจาก deleted_at) ตัวที่ไม่เจอจะไม่อยู่ในผลลัพธ์
func (r *repository) FindStocksByIds(ctx context.Context, ids []string) ([]models.Stock, error) {
	var stocks []models.Stock
	if len(ids) == 0 {
		return stocks, nil
	}
	if err := r.db.WithContext(ctx).Unscoped().Where("id IN ?", ids).Find(&stocks).Error; err != nil {
		return nil, err
	}
	return stocks, nil
}

// FindCourseById ค้นหาคอร์สตาม id (ไม่รวมที่ถูกลบด้วย soft delete) ไม่เจอคืน nil
func (r *repository) FindCourseById(ctx context.Context, id string) (*models.Course, error) {
	var course models.Course
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&course).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &course, nil
}

// UpdateCourse บันทึกคอร์สทั้งแถว ในทรานแซกชันเดียวกับของที่ใช้
// items = nil ไม่แตะ course_items · ไม่ nil = แทนทั้งชุด (ลบของเดิมแล้วใส่ใหม่ course_id ใส่ให้เอง)
// คืนคอร์สที่อ่านกลับจาก DB พร้อม items แบบเดียวกับ GetAllCourse
func (r *repository) UpdateCourse(ctx context.Context, course models.Course, items *[]models.CourseItem) (models.Course, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Omit กัน GORM บันทึก association ใน course.Items เอง
		if err := tx.Omit("Items").Save(&course).Error; err != nil {
			return err
		}
		if items == nil {
			return nil
		}
		// course_items ไม่มีตารางไหนอ้างถึง ลบแล้วใส่ใหม่ได้เลย
		if err := tx.Where("course_id = ?", course.ID).Delete(&models.CourseItem{}).Error; err != nil {
			return err
		}
		if len(*items) == 0 {
			return nil
		}
		for i := range *items {
			(*items)[i].CourseID = course.ID
		}
		return tx.Omit("Course", "Stock").Create(items).Error
	})
	if err != nil {
		return models.Course{}, err
	}

	var updated models.Course
	if err := r.db.WithContext(ctx).Scopes(preloadItems).First(&updated, "id = ?", course.ID).Error; err != nil {
		return models.Course{}, err
	}
	return updated, nil
}

// FindCourseItems คืนของที่คอร์สนี้ใช้อยู่ตอนนี้
func (r *repository) FindCourseItems(ctx context.Context, courseId string) ([]models.CourseItem, error) {
	var items []models.CourseItem
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseId).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// CheckCourseHasRelation เช็คว่า Course มีประวัติที่อ้างถึงหรือไม่ (มี = ต้อง soft delete)
// ต้องครบทุกตารางประวัติที่มี FK course_id ชี้มาที่ courses ไม่งั้น hard delete จะติด FK
// Relations: OrderItem (เคยขายแล้ว), ServiceUsage (VIP เคยใช้บริการนี้)
// ไม่นับ CourseItem เพราะเป็นส่วนหนึ่งของคอร์ส ไม่ใช่ประวัติ (HardDeleteCourse ลบไปพร้อมกัน)
func (r *repository) CheckCourseHasRelation(ctx context.Context, courseId string) (bool, error) {
	relatedModels := []any{
		&models.OrderItem{},
		&models.ServiceUsage{},
	}

	for _, m := range relatedModels {
		var count int64
		err := r.db.WithContext(ctx).Model(m).Where("course_id = ?", courseId).Limit(1).Count(&count).Error
		if err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}

// SoftDeleteCourse เซ็ต deleted_at (แถวยังอยู่ ประวัติที่อ้างถึงยังเปิดดูได้)
func (r *repository) SoftDeleteCourse(ctx context.Context, courseId string) error {
	return r.db.WithContext(ctx).Delete(&models.Course{}, "id = ?", courseId).Error
}

// HardDeleteCourse ลบ Course ออกจากฐานข้อมูลจริงๆ พร้อมของที่คอร์สใช้ (course_items) ในทรานแซกชันเดียว
// ถ้าติด FK จากตารางประวัติ จะ rollback ทั้งหมด (course_items กลับมาครบ)
func (r *repository) HardDeleteCourse(ctx context.Context, courseId string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", courseId).Delete(&models.CourseItem{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&models.Course{}, "id = ?", courseId).Error
	})
}
