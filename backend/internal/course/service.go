package course

import (
	"backend/internal/models"
	_package "backend/package"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

type Service interface {
	GetAllCourse(ctx context.Context, q GetAllCourseQuery) ([]CourseResponse, _package.PageMeta, error)
	AddCourse(ctx context.Context, data AddCourseRequest) (CourseResponse, error)
	UpdateCourseById(ctx context.Context, courseId string, data UpdateCourseRequest) (CourseResponse, error)
	DeleteCourseById(ctx context.Context, courseId string) error
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) GetAllCourse(ctx context.Context, q GetAllCourseQuery) ([]CourseResponse, _package.PageMeta, error) {
	page, pageSize, err := _package.ParsePaging(q.Page, q.PageSize)
	if err != nil {
		return nil, _package.PageMeta{}, err
	}
	courses, total, err := s.repository.GetAllCourse(ctx, strings.TrimSpace(q.Search), page, pageSize)
	if err != nil {
		return nil, _package.PageMeta{}, err
	}
	res := make([]CourseResponse, 0, len(courses))
	for _, c := range courses {
		res = append(res, toCourseResponse(c))
	}
	return res, _package.NewPageMeta(page, pageSize, total), nil
}

func (s *service) AddCourse(ctx context.Context, data AddCourseRequest) (CourseResponse, error) {
	var emptyCourse CourseResponse

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

	// Validate ของที่ใช้ (เช็คกับสินค้าจริงใน DB)
	items, err := s.buildCourseItems(ctx, data.Items, nil)
	if err != nil {
		return emptyCourse, err
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

	createdCourse, err := s.repository.CreateCourse(ctx, newCourse, items)
	if err != nil {
		// สองคำขอสร้างชื่อเดียวกันพร้อมกันผ่านการเช็คชื่อด้านบนได้ทั้งคู่ unique index เป็นตัวกันจริง
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return emptyCourse, duplicateName(courseName)
		}
		// สินค้าถูกลบจริง (hard delete) ระหว่างเช็คกับบันทึก
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return emptyCourse, invalid("some stock in items no longer exists")
		}
		return emptyCourse, err
	}

	return toCourseResponse(createdCourse), nil
}

// buildCourseItems ตรวจ items ที่ส่งมาแล้วแปลงเป็น CourseItem (ยังไม่มี course_id)
// stock ต้องมีอยู่จริง ไม่ถูกลบ และยังเปิดใช้งาน · ห้าม stock_id ซ้ำ · use_qty ตั้งแต่ 1
// ของที่เปิดเป็นกระปุก (uses_per_unit > 1) ใช้ต่อการทำ 1 ครั้งได้ไม่เกิน 1 กระปุก
// linked = stock_id (ตัวเล็ก) ที่ผูกกับคอร์สอยู่แล้ว ข้ามการเช็คถูกลบ/ปิดใช้งาน แก้คอร์สได้โดยไม่ต้องเอาของเดิมออก
func (s *service) buildCourseItems(ctx context.Context, reqs []CourseItemRequest, linked map[string]bool) ([]models.CourseItem, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	ids := make([]string, 0, len(reqs))
	seen := make(map[string]bool, len(reqs))
	for i, it := range reqs {
		// UUID ตัวใหญ่ตัวเล็กเป็นตัวเดียวกัน
		stockId := strings.ToLower(strings.TrimSpace(it.StockID))
		if !uuidPattern.MatchString(stockId) {
			return nil, invalid("items[%d].stock_id must be a valid UUID", i)
		}
		if seen[stockId] {
			return nil, invalid("items[%d].stock_id '%s' is duplicated", i, stockId)
		}
		if it.UseQty < 1 {
			return nil, invalid("items[%d].use_qty must be at least 1", i)
		}
		seen[stockId] = true
		ids = append(ids, stockId)
	}

	stocks, err := s.repository.FindStocksByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	stockById := make(map[string]models.Stock, len(stocks))
	for _, st := range stocks {
		stockById[strings.ToLower(st.ID)] = st
	}

	items := make([]models.CourseItem, 0, len(reqs))
	for i, it := range reqs {
		st, ok := stockById[ids[i]]
		if !ok || (st.DeletedAt.Valid && !linked[ids[i]]) {
			return nil, invalid("items[%d].stock_id '%s' not found", i, ids[i])
		}
		if !st.IsActive && !linked[ids[i]] {
			return nil, invalid("items[%d].stock_id '%s' is inactive", i, ids[i])
		}
		if st.UsesPerUnit > 1 && it.UseQty > st.UsesPerUnit {
			return nil, invalid("items[%d].use_qty must not exceed %d (uses per unit of '%s')", i, st.UsesPerUnit, st.Name)
		}
		items = append(items, models.CourseItem{StockID: st.ID, UseQty: it.UseQty})
	}
	return items, nil
}

// UpdateCourseById แก้เฉพาะ field ที่ส่งมา คอร์สที่ถูก soft delete ไปแล้วตอบ 404
func (s *service) UpdateCourseById(ctx context.Context, courseId string, data UpdateCourseRequest) (CourseResponse, error) {
	var emptyCourse CourseResponse

	// id ที่ไม่ใช่ UUID ไม่มีในตารางแน่นอน และถ้าส่งไปถึง PostgreSQL จะ error กลายเป็น 500
	if !uuidPattern.MatchString(courseId) {
		return emptyCourse, notFound(courseId)
	}

	existingCourse, err := s.repository.FindCourseById(ctx, courseId)
	if err != nil {
		return emptyCourse, err
	}
	if existingCourse == nil {
		return emptyCourse, notFound(courseId)
	}

	// Name
	if data.Name != nil {
		courseName := strings.TrimSpace(*data.Name)
		if courseName == "" {
			return emptyCourse, invalid("course name cannot be empty")
		}
		if utf8.RuneCountInString(courseName) > 150 {
			return emptyCourse, invalid("course name must not exceed 150 characters")
		}

		// เช็คชื่อซ้ำกับคอร์สอื่นที่ยังไม่ถูกลบ (ชื่อเดิมของตัวเอง หรือชื่อของคอร์สที่ถูก soft delete แล้ว ใช้ได้)
		if courseName != existingCourse.Name {
			duplicateCourse, err := s.repository.FindCourseByName(ctx, courseName)
			if err != nil {
				return emptyCourse, err
			}
			if duplicateCourse != nil && duplicateCourse.ID != courseId && duplicateCourse.DeletedAt.Time.IsZero() {
				return emptyCourse, duplicateName(courseName)
			}
		}

		existingCourse.Name = courseName
	}

	// Price
	if data.Price != nil {
		if data.Price.IsNegative() {
			return emptyCourse, invalid("price must be greater than or equal to 0")
		}
		existingCourse.Price = *data.Price
	}

	// MaxSession
	if data.MaxSession != nil {
		if *data.MaxSession < 1 {
			return emptyCourse, invalid("max session must be at least 1")
		}
		existingCourse.MaxSession = *data.MaxSession
	}

	// ValidDay (null = ไม่หมดอายุ)
	if data.ValidDay != nil {
		validDay, err := parseValidDay(data.ValidDay)
		if err != nil {
			return emptyCourse, err
		}
		existingCourse.ValidDay = validDay
	}

	// ServiceComm
	if data.ServiceComm != nil {
		if data.ServiceComm.IsNegative() {
			return emptyCourse, invalid("service commission must be greater than or equal to 0")
		}
		existingCourse.ServiceComm = *data.ServiceComm
	}

	// Category
	if data.Category != nil {
		category := strings.TrimSpace(*data.Category)
		if utf8.RuneCountInString(category) > 50 {
			return emptyCourse, invalid("category must not exceed 50 characters")
		}
		existingCourse.Category = category
	}

	// IsActive
	if data.IsActive != nil {
		existingCourse.IsActive = *data.IsActive
	}

	// Items (แทนทั้งชุด) ของที่ผูกกับคอร์สนี้อยู่แล้วคงไว้ได้แม้สินค้าจะถูกปิดใช้งาน/ลบไปแล้ว
	var items *[]models.CourseItem
	if data.Items != nil {
		current, err := s.repository.FindCourseItems(ctx, courseId)
		if err != nil {
			return emptyCourse, err
		}
		linked := make(map[string]bool, len(current))
		for _, it := range current {
			linked[strings.ToLower(it.StockID)] = true
		}

		built, err := s.buildCourseItems(ctx, *data.Items, linked)
		if err != nil {
			return emptyCourse, err
		}
		if built == nil {
			built = []models.CourseItem{}
		}
		items = &built
	}

	updatedCourse, err := s.repository.UpdateCourse(ctx, *existingCourse, items)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return emptyCourse, duplicateName(existingCourse.Name)
		}
		// สินค้าถูกลบจริง (hard delete) ระหว่างเช็คกับบันทึก
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return emptyCourse, invalid("some stock in items no longer exists")
		}
		return emptyCourse, err
	}

	return toCourseResponse(updatedCourse), nil
}

// DeleteCourseById ลบคอร์ส:
// - ถ้าไม่มีประวัติ (ยังไม่เคยขาย / ใช้บริการ) → Hard Delete (ลบจริง พร้อมของที่คอร์สใช้)
// - ถ้ามีประวัติ → Soft Delete (เซ็ต DeletedAt) ประวัติบิล / การใช้บริการ และของที่คอร์สใช้ยังอยู่ครบ
func (s *service) DeleteCourseById(ctx context.Context, courseId string) error {
	if !uuidPattern.MatchString(courseId) {
		return notFound(courseId)
	}

	existingCourse, err := s.repository.FindCourseById(ctx, courseId)
	if err != nil {
		return err
	}
	if existingCourse == nil {
		return notFound(courseId)
	}

	hasRelation, err := s.repository.CheckCourseHasRelation(ctx, courseId)
	if err != nil {
		return err
	}
	if hasRelation {
		return s.repository.SoftDeleteCourse(ctx, courseId)
	}

	err = s.repository.HardDeleteCourse(ctx, courseId)
	// มีคนขายคอร์สนี้ระหว่างที่เช็คกับลบพอดี FK จะกันไว้ → ถอยไปใช้ Soft Delete แทนการตอบ 500
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return s.repository.SoftDeleteCourse(ctx, courseId)
	}
	return err
}

// Helper Functions

// parseValidDay แปลง valid_day ที่รับดิบมา null = ไม่หมดอายุ (คืน nil) ไม่งั้นต้องเป็นจำนวนเต็มตั้งแต่ 1
// (ไม่ส่งข้อความดิบของ encoding/json ออกไป)
func parseValidDay(raw json.RawMessage) (*int, error) {
	if string(raw) == "null" {
		return nil, nil
	}
	var validDay int
	if err := json.Unmarshal(raw, &validDay); err != nil {
		return nil, invalid("valid_day must be a whole number or null")
	}
	if validDay < 1 {
		return nil, invalid("valid day must be at least 1")
	}
	return &validDay, nil
}

func notFound(id string) error {
	return _package.NewAppError("Course not found", _package.CodeNotFound,
		fmt.Sprintf("Course with id '%s' not found", id))
}

func invalid(format string, args ...any) error {
	return _package.NewAppError("Invalid request", _package.CodeValidation, fmt.Sprintf(format, args...))
}

func nameConflict(detail string) error {
	return _package.NewAppError("Course name already exists", _package.CodeConflict, detail)
}

func duplicateName(name string) error {
	return nameConflict(fmt.Sprintf("Course with name '%s' already exists in the system", name))
}
