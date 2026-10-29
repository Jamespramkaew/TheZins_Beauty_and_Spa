package course

import (
	"backend/internal/models"
	_package "backend/package"
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

type Service interface {
	GetAllCourse(ctx context.Context, q GetAllCourseQuery) ([]models.Course, _package.PageMeta, error)
	AddCourse(ctx context.Context, data AddCourseRequest) (models.Course, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) GetAllCourse(ctx context.Context, q GetAllCourseQuery) ([]models.Course, _package.PageMeta, error) {
	page, pageSize, err := _package.ParsePaging(q.Page, q.PageSize)
	if err != nil {
		return nil, _package.PageMeta{}, err
	}
	courses, total, err := s.repository.GetAllCourse(ctx, page, pageSize)
	if err != nil {
		return nil, _package.PageMeta{}, err
	}
	return courses, _package.NewPageMeta(page, pageSize, total), nil
}

func (s *service) AddCourse(ctx context.Context, data AddCourseRequest) (models.Course, error) {
	var emptyCourse models.Course

	// Validate ชื่อคอร์ส
	courseName := strings.TrimSpace(data.Name)
	if courseName == "" {
		return emptyCourse, invalid("course name is required")
	}
	if utf8.RuneCountInString(courseName) > 150 {
		return emptyCourse, invalid("course name must not exceed 150 characters")
	}

	// Validate Price
	if data.Price.IsNegative() {
		return emptyCourse, invalid("price must be greater than or equal to 0")
	}

	// Validate MaxSession
	if data.MaxSession < 1 {
		return emptyCourse, invalid("max session must be at least 1")
	}

	// Validate ValidDay (ถ้ามีค่า ต้องมากกว่า 0)
	if data.ValidDay != nil && *data.ValidDay < 1 {
		return emptyCourse, invalid("valid day must be at least 1")
	}

	// Validate ServiceComm
	if data.ServiceComm.IsNegative() {
		return emptyCourse, invalid("service commission must be greater than or equal to 0")
	}

	// Validate Category length
	category := strings.TrimSpace(data.Category)
	if utf8.RuneCountInString(category) > 50 {
		return emptyCourse, invalid("category must not exceed 50 characters")
	}

	// เช็คชื่อซ้ำในระบบ
	existingCourse, err := s.repository.FindCourseByName(ctx, courseName)
	if err != nil {
		return emptyCourse, err
	}

	if existingCourse != nil {
		// ถ้าไม่มี DeletedAt (ยังไม่ถูกลบ) → ห้ามเพิ่ม
		if existingCourse.DeletedAt.Time.IsZero() {
			return emptyCourse, duplicateName(courseName)
		}
	}

	// สร้าง Course ใหม่
	newCourse := models.Course{
		Name:        courseName,
		Price:       data.Price,
		MaxSession:  data.MaxSession,
		ValidDay:    data.ValidDay,
		ServiceComm: data.ServiceComm,
		Category:    category,
		IsActive:    true,
	}

	createdCourse, err := s.repository.CreateCourse(ctx, newCourse)
	if err != nil {
		// สองคำขอสร้างชื่อเดียวกันพร้อมกันผ่านการเช็คชื่อด้านบนได้ทั้งคู่ unique index เป็นตัวกันจริง
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return emptyCourse, duplicateName(courseName)
		}
		return emptyCourse, err
	}

	return createdCourse, nil
}

// Helper Functions

func invalid(format string, args ...any) error {
	return _package.NewAppError("Invalid request", _package.CodeValidation, fmt.Sprintf(format, args...))
}

func nameConflict(detail string) error {
	return _package.NewAppError("Course name already exists", _package.CodeConflict, detail)
}

func duplicateName(name string) error {
	return nameConflict(fmt.Sprintf("Course with name '%s' already exists in the system", name))
}
