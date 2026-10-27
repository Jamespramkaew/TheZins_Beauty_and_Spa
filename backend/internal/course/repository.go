package course

import (
	"context"

	"backend/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	GetAllCourse(ctx context.Context) ([]models.Course, error)
	FindCourseByName(ctx context.Context, name string) (*models.Course, error)
	CreateCourse(ctx context.Context, course models.Course) (models.Course, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAllCourse(ctx context.Context) ([]models.Course, error) {
	var courses []models.Course
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
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
