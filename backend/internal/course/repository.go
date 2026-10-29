package course

import (
	"context"

	"backend/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetAllCourse(ctx context.Context, page, pageSize int) ([]models.Course, int64, error)
	FindCourseByName(ctx context.Context, name string) (*models.Course, error)
	CreateCourse(ctx context.Context, course models.Course) (models.Course, error)
	FindCourseById(ctx context.Context, id string) (*models.Course, error)
	UpdateCourse(ctx context.Context, course models.Course) (models.Course, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// GetAllCourse คืนคอร์สทีละหน้า เรียงตามชื่อ · คืนจำนวนทั้งหมด (total) ไว้ทำ meta
func (r *repository) GetAllCourse(ctx context.Context, page, pageSize int) ([]models.Course, int64, error) {
	// Session ทำให้ใช้ query เดียวกันซ้ำได้ทั้งตอน Count และ Find
	all := r.db.WithContext(ctx).Model(&models.Course{}).Session(&gorm.Session{})

	var total int64
	if err := all.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var courses []models.Course
	// id เป็นตัวตัดสินเมื่อชื่อเท่ากัน ไม่งั้นข้ามหน้าแล้วแถวอาจซ้ำหรือตกหล่น
	err := all.
		Order("name ASC, id ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&courses).Error
	if err != nil {
		return nil, 0, err
	}
	return courses, total, nil
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

// CreateCourse สร้างคอร์สใหม่
func (r *repository) CreateCourse(ctx context.Context, course models.Course) (models.Course, error) {
	if err := r.db.WithContext(ctx).Create(&course).Error; err != nil {
		return models.Course{}, err
	}
	return course, nil
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

// UpdateCourse บันทึกคอร์สทั้งแถว
func (r *repository) UpdateCourse(ctx context.Context, course models.Course) (models.Course, error) {
	if err := r.db.WithContext(ctx).Save(&course).Error; err != nil {
		return models.Course{}, err
	}
	return course, nil
}
