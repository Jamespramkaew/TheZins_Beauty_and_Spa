package order

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	CustomerExists(ctx context.Context, id string) (bool, error)
	FindCourses(ctx context.Context, ids []string) (map[string]models.Course, error)
	FindStocks(ctx context.Context, ids []string) (map[string]models.Stock, error)
	Create(ctx context.Context, yymm string, order *models.Order, items []models.OrderItem, payments []models.Payment) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CustomerExists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Customer{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// FindCourses ดึงคอร์สตาม id คืนเป็น map[id]คอร์ส (คอร์สที่ถูกลบแบบ soft delete จะไม่มีใน map)
func (r *repository) FindCourses(ctx context.Context, ids []string) (map[string]models.Course, error) {
	result := map[string]models.Course{}
	if len(ids) == 0 {
		return result, nil
	}
	var courses []models.Course
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&courses).Error; err != nil {
		return nil, err
	}
	for _, c := range courses {
		result[c.ID] = c
	}
	return result, nil
}

// FindStocks ดึงสินค้าตาม id คืนเป็น map[id]สินค้า
func (r *repository) FindStocks(ctx context.Context, ids []string) (map[string]models.Stock, error) {
	result := map[string]models.Stock{}
	if len(ids) == 0 {
		return result, nil
	}
	var stocks []models.Stock
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&stocks).Error; err != nil {
		return nil, err
	}
	for _, s := range stocks {
		result[s.ID] = s
	}
	return result, nil
}

// Create บันทึกบิล + รายการ + การจ่าย ใน transaction เดียว และออกเลขบิลให้
// เลขบิล = รหัสสาขา-YYMM-เลขรัน 4 หลักของสาขานั้นในเดือนนั้น เช่น S1-2610-0001
// อะไรผิดกลางทาง ทุกอย่างถูกยกเลิก (rollback) ไม่มีบิลครึ่งๆ กลางๆ ค้างในฐานข้อมูล
func (r *repository) Create(ctx context.Context, yymm string, order *models.Order, items []models.OrderItem, payments []models.Payment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) ล็อกแถวสาขา (SELECT ... FOR UPDATE)
		// คนอื่นที่ออกบิลสาขาเดียวกันพร้อมกันต้องรอจนเรา commit เลขจึงไม่ซ้ำ
		var branch models.Branch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&branch, "id = ?", order.BranchID).Error; err != nil {
			return err
		}

		// 2) หาเลขล่าสุดของสาขานี้ในเดือนนี้ แล้ว +1
		prefix := fmt.Sprintf("%s-%s-", branch.Code, yymm)
		var last models.Order
		err := tx.Where("order_no LIKE ?", prefix+"%").Order("order_no DESC").First(&last).Error
		next := 1
		if err == nil {
			n, convErr := strconv.Atoi(strings.TrimPrefix(last.OrderNo, prefix))
			if convErr != nil {
				return convErr
			}
			next = n + 1
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		order.OrderNo = fmt.Sprintf("%s%04d", prefix, next)

		// 3) บันทึกบิล (ได้ order.ID กลับมา)
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 4) บันทึกรายการ: ผูกกับบิล และให้เลขบรรทัด 1..n ตามลำดับที่ส่งมา
		for i := range items {
			items[i].OrderID = order.ID
			items[i].LineNo = i + 1
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		// 5) บันทึกการจ่าย (ถ้ามี)
		if len(payments) > 0 {
			for i := range payments {
				payments[i].OrderID = order.ID
			}
			if err := tx.Create(&payments).Error; err != nil {
				return err
			}
		}

		return nil // return nil = commit
	})
}
