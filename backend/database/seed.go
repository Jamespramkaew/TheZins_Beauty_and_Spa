package database

import (
	"fmt"
	"strings"

	"backend/internal/models"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedBranches สร้างสาขาตั้งต้น รันซ้ำได้ (มีรหัสสาขานี้อยู่แล้วจะข้าม)
func SeedBranches(db *gorm.DB) error {
	branches := []models.Branch{
		{Code: "S1", Name: "สาขา 1"},
		{Code: "S2", Name: "สาขา 2"},
		{Code: "S3", Name: "สาขา 3"},
	}

	for _, b := range branches {
		err := db.Where(models.Branch{Code: b.Code}).FirstOrCreate(&b).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// seedPassword = รหัสผ่านของบัญชีตั้งต้นทุกบัญชี ใช้ทดสอบในเครื่องเท่านั้น ห้ามใช้บนเซิร์ฟเวอร์จริง
const seedPassword = "zins1234"

// SeedUsers สร้าง admin 1 บัญชี + บัญชีสาขาละ 1 (s1, s2, s3) รันซ้ำได้ (มี username นี้อยู่แล้วจะข้าม)
func SeedUsers(db *gorm.DB) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// admin ไม่ผูกสาขา (branch_id = NULL)
	admin := models.User{Username: "admin", Password: string(hash), Role: "admin"}
	err = db.Where(models.User{Username: admin.Username}).FirstOrCreate(&admin).Error
	if err != nil {
		return err
	}

	// บัญชีสาขา: ชื่อผู้ใช้ = รหัสสาขาตัวเล็ก เช่น S1 → s1
	var branches []models.Branch
	if err := db.Find(&branches).Error; err != nil {
		return err
	}
	for _, b := range branches {
		user := models.User{
			Username: strings.ToLower(b.Code),
			Password: string(hash),
			Role:     "branch",
			BranchID: &b.ID,
		}
		err := db.Where(models.User{Username: user.Username}).FirstOrCreate(&user).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// seedEmployee = พนักงานตัวอย่าง 1 คน (ชื่อต้องไม่ซ้ำกันภายในสาขาเดียวกัน)
type seedEmployee struct {
	branch string // รหัสสาขา เช่น "S1"
	name   string
	role   string // consult / manager / TP
}

// สาขาละ 3 คน: manager (รับของเข้า) / TP (เปิดกระปุก) / consult
var seedEmployees = []seedEmployee{
	{"S1", "สมศรี ใจดี", "manager"}, {"S1", "วิภา สุขสันต์", "TP"}, {"S1", "นภา รักงาน", "consult"},
	{"S2", "สมชาย มั่นคง", "manager"}, {"S2", "จินตนา แสนสุข", "TP"}, {"S2", "ปริญญา ดีเลิศ", "consult"},
	{"S3", "สุดา พึ่งพา", "manager"}, {"S3", "มาลี งามตา", "TP"}, {"S3", "อรทัย ใจงาม", "consult"},
}

// SeedEmployees สร้างพนักงานตัวอย่าง รันซ้ำได้ (มีชื่อนี้ในสาขานี้อยู่แล้วจะข้าม) ต้องรันหลัง SeedBranches
func SeedEmployees(db *gorm.DB) error {
	branchByCode, err := loadBranchesByCode(db)
	if err != nil {
		return err
	}
	for _, e := range seedEmployees {
		branch, ok := branchByCode[e.branch]
		if !ok {
			return fmt.Errorf("seed employee %q: branch %q not found", e.name, e.branch)
		}
		emp := models.Employee{Name: e.name, Role: e.role, BranchID: branch.ID, IsActive: true}
		err := db.Where(models.Employee{Name: e.name, BranchID: branch.ID}).FirstOrCreate(&emp).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// SeedCourses สร้างคอร์สตัวอย่าง รันซ้ำได้ (มีชื่อคอร์สนี้อยู่แล้วจะข้าม)
func SeedCourses(db *gorm.DB) error {
	courses := []models.Course{
		{Name: "คอร์สหน้าใส 10 ครั้ง", Price: decimal.NewFromInt(3500), MaxSession: 10, ValidDay: new(180), Category: "facial"},
		{Name: "คอร์สนวดหน้า 5 ครั้ง", Price: decimal.NewFromInt(1500), MaxSession: 5, ValidDay: new(90), Category: "facial"},
		{Name: "คอร์สนวดตัว 10 ครั้ง", Price: decimal.NewFromInt(4900), MaxSession: 10, ValidDay: new(365), Category: "body"},
		// ValidDay ไม่ใส่ = nil = ไม่หมดอายุ
		{Name: "ทรีตเมนต์หน้า 1 ครั้ง", Price: decimal.NewFromInt(590), MaxSession: 1, Category: "facial"},
	}

	for _, c := range courses {
		err := db.Where(models.Course{Name: c.Name}).FirstOrCreate(&c).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// SeedStocks สร้างสินค้าตัวอย่าง รันซ้ำได้ (มีชื่อสินค้านี้อยู่แล้วจะข้าม)
func SeedStocks(db *gorm.DB) error {
	stocks := []models.Stock{
		{Name: stockCream, Price: decimal.NewFromInt(890), IsSellable: true, Category: "premium", UsesPerUnit: 1, MinQty: 5},
		{Name: stockSerum, Price: decimal.NewFromInt(1290), IsSellable: true, Category: "premium", UsesPerUnit: 1, MinQty: 3},
		{Name: stockSunscreen, Price: decimal.NewFromInt(450), IsSellable: true, Category: "normal", UsesPerUnit: 1, MinQty: 5},
		// ของใช้ในร้าน ห้ามขาย 1 กระปุกใช้ได้ 20 ครั้ง
		{Name: stockGel, IsSellable: false, Category: "normal", UsesPerUnit: 20, MinQty: 2},
		// ของใช้ในร้าน ห้ามขาย 1 กระปุกใช้ได้ 10 ครั้ง
		{Name: stockMask, IsSellable: false, Category: "normal", UsesPerUnit: 10, MinQty: 2},
	}

	for _, s := range stocks {
		err := db.Where(models.Stock{Name: s.Name}).FirstOrCreate(&s).Error
		if err != nil {
			return err
		}
	}
	return nil
}
