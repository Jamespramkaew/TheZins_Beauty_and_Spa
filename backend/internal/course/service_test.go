package course

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"backend/internal/models"
	_package "backend/package"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// fakeRepo เก็บคอร์สไว้ใน memory รวมตัวที่ถูก soft delete (DeletedAt ไม่ว่าง)
type fakeRepo struct {
	courses []models.Course
	saved   *models.Course

	related     map[string]bool // course id ที่มีตารางอื่นอ้างถึง
	hardErr     error           // error ที่ HardDeleteCourse จะคืน (จำลอง FK)
	softDeleted []string
	hardDeleted []string
}

func (r *fakeRepo) CheckCourseHasRelation(ctx context.Context, courseId string) (bool, error) {
	return r.related[courseId], nil
}

func (r *fakeRepo) SoftDeleteCourse(ctx context.Context, courseId string) error {
	r.softDeleted = append(r.softDeleted, courseId)
	return nil
}

func (r *fakeRepo) HardDeleteCourse(ctx context.Context, courseId string) error {
	if r.hardErr != nil {
		return r.hardErr
	}
	r.hardDeleted = append(r.hardDeleted, courseId)
	return nil
}

func (r *fakeRepo) GetAllCourse(ctx context.Context, page, pageSize int) ([]models.Course, int64, error) {
	return r.courses, int64(len(r.courses)), nil
}

// FindCourseByName เหมือนของจริง: รวมตัวที่ถูกลบ แต่ตัวที่ยังไม่ถูกลบขึ้นก่อน
func (r *fakeRepo) FindCourseByName(ctx context.Context, name string) (*models.Course, error) {
	var deleted *models.Course
	for i := range r.courses {
		if r.courses[i].Name != name {
			continue
		}
		if r.courses[i].DeletedAt.Time.IsZero() {
			return &r.courses[i], nil
		}
		deleted = &r.courses[i]
	}
	return deleted, nil
}

func (r *fakeRepo) CreateCourse(ctx context.Context, course models.Course) (models.Course, error) {
	return course, nil
}

func (r *fakeRepo) FindCourseById(ctx context.Context, id string) (*models.Course, error) {
	for i := range r.courses {
		if r.courses[i].ID == id && r.courses[i].DeletedAt.Time.IsZero() {
			c := r.courses[i]
			return &c, nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) UpdateCourse(ctx context.Context, course models.Course) (models.Course, error) {
	r.saved = &course
	return course, nil
}

const (
	idFacial  = "11111111-1111-1111-1111-111111111111"
	idMassage = "22222222-2222-2222-2222-222222222222"
	idOldSpa  = "33333333-3333-3333-3333-333333333333"
	idMissing = "99999999-9999-9999-9999-999999999999"
)

func newFakeRepo() *fakeRepo {
	validDay := 90
	return &fakeRepo{courses: []models.Course{
		{ID: idFacial, Name: "Facial", Price: decimal.NewFromInt(1000), MaxSession: 5, ValidDay: &validDay, IsActive: true},
		{ID: idMassage, Name: "Massage", Price: decimal.NewFromInt(800), MaxSession: 10, IsActive: true},
		{ID: idOldSpa, Name: "Old Spa", MaxSession: 1,
			DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true}},
	}}
}

func ptr[T any](v T) *T { return &v }

func errCode(err error) string {
	var appErr *_package.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}

func TestUpdateCourse_Name(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		newName  string
		wantCode string // "" = สำเร็จ
	}{
		{"ชื่อซ้ำกับคอร์สอื่นที่ยังไม่ถูกลบ → 409", idFacial, "Massage", _package.CodeConflict},
		{"ชื่อซ้ำหลัง trim ช่องว่าง → 409", idFacial, "  Massage  ", _package.CodeConflict},
		{"ชื่อตรงกับคอร์สที่ถูก soft delete แล้ว → ใช้ได้", idFacial, "Old Spa", ""},
		{"ส่งชื่อเดิมของตัวเอง → ใช้ได้", idFacial, "Facial", ""},
		{"ชื่อใหม่ไม่ซ้ำใคร → ใช้ได้", idFacial, "Facial Deluxe", ""},
		{"ชื่อว่าง → 400", idFacial, "   ", _package.CodeValidation},
		{"คอร์สไม่มีในระบบ → 404", idMissing, "Anything", _package.CodeNotFound},
		{"คอร์สถูก soft delete แล้ว → 404", idOldSpa, "Anything", _package.CodeNotFound},
		{"id ไม่ใช่ UUID → 404", "abc", "Anything", _package.CodeNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			svc := NewService(repo)

			_, err := svc.UpdateCourseById(context.Background(), tt.id, UpdateCourseRequest{Name: ptr(tt.newName)})
			if got := errCode(err); got != tt.wantCode {
				t.Fatalf("code = %q, want %q (err = %v)", got, tt.wantCode, err)
			}
			if tt.wantCode != "" && repo.saved != nil {
				t.Fatal("must not save when the request is rejected")
			}
		})
	}
}

func TestUpdateCourse_OnlySentFieldsChange(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)

	got, err := svc.UpdateCourseById(context.Background(), idFacial, UpdateCourseRequest{Price: ptr(decimal.NewFromInt(1500))})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Price.Equal(decimal.NewFromInt(1500)) {
		t.Errorf("price = %s, want 1500", got.Price)
	}
	if got.Name != "Facial" || got.MaxSession != 5 || got.ValidDay == nil || *got.ValidDay != 90 {
		t.Errorf("fields not sent must stay the same, got %+v", got)
	}
}

func TestUpdateCourse_ValidDay(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		want     *int
		wantCode string
	}{
		{"ไม่ส่ง valid_day → คงค่าเดิม", `{"price": 1}`, ptr(90), ""},
		{"ส่ง null → ไม่หมดอายุ", `{"valid_day": null}`, nil, ""},
		{"ส่งตัวเลข → เปลี่ยนค่า", `{"valid_day": 30}`, ptr(30), ""},
		{"ส่ง 0 → 400", `{"valid_day": 0}`, nil, _package.CodeValidation},
		{"ส่งเป็นข้อความ → 400", `{"valid_day": "30"}`, nil, _package.CodeValidation},
		{"ส่งทศนิยม → 400", `{"valid_day": 2.5}`, nil, _package.CodeValidation},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req UpdateCourseRequest
			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatal(err)
			}
			got, err := NewService(newFakeRepo()).UpdateCourseById(context.Background(), idFacial, req)
			if code := errCode(err); code != tt.wantCode {
				t.Fatalf("code = %q, want %q (err = %v)", code, tt.wantCode, err)
			}
			if tt.wantCode != "" {
				return
			}
			if (got.ValidDay == nil) != (tt.want == nil) || (got.ValidDay != nil && *got.ValidDay != *tt.want) {
				t.Errorf("valid_day = %v, want %v", got.ValidDay, tt.want)
			}
		})
	}
}

func TestDeleteCourse(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		related  bool
		hardErr  error
		wantCode string
		wantSoft bool
		wantHard bool
	}{
		{"ไม่มี relation → hard delete", idFacial, false, nil, "", false, true},
		{"มี relation → soft delete", idFacial, true, nil, "", true, false},
		{"ติด FK ตอน hard delete (มีคนขายพอดี) → ถอยไป soft delete", idFacial, false, gorm.ErrForeignKeyViolated, "", true, false},
		{"คอร์สไม่มีในระบบ → 404", idMissing, false, nil, _package.CodeNotFound, false, false},
		{"คอร์สถูก soft delete ไปแล้ว → 404", idOldSpa, false, nil, _package.CodeNotFound, false, false},
		{"id ไม่ใช่ UUID → 404", "abc", false, nil, _package.CodeNotFound, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			repo.related = map[string]bool{tt.id: tt.related}
			repo.hardErr = tt.hardErr

			err := NewService(repo).DeleteCourseById(context.Background(), tt.id)
			if code := errCode(err); code != tt.wantCode {
				t.Fatalf("code = %q, want %q (err = %v)", code, tt.wantCode, err)
			}
			if tt.wantCode == "" && err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got := len(repo.softDeleted) > 0; got != tt.wantSoft {
				t.Errorf("soft deleted = %v, want %v", got, tt.wantSoft)
			}
			if got := len(repo.hardDeleted) > 0; got != tt.wantHard {
				t.Errorf("hard deleted = %v, want %v", got, tt.wantHard)
			}
		})
	}
}
