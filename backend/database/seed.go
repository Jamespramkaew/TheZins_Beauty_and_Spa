package database

import (
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
		{Name: "ครีมบำรุงหน้า", Price: decimal.NewFromInt(890), IsSellable: true, Category: "premium", UsesPerUnit: 1, MinQty: 5},
		{Name: "เซรั่มวิตามินซี", Price: decimal.NewFromInt(1290), IsSellable: true, Category: "premium", UsesPerUnit: 1, MinQty: 3},
		{Name: "ครีมกันแดด", Price: decimal.NewFromInt(450), IsSellable: true, Category: "normal", UsesPerUnit: 1, MinQty: 5},
		// ของใช้ในร้าน ห้ามขาย 1 กระปุกใช้ได้ 20 ครั้ง
		{Name: "เจลนวดหน้า (ใช้ในร้าน)", IsSellable: false, Category: "normal", UsesPerUnit: 20, MinQty: 2},
	}

	for _, s := range stocks {
		err := db.Where(models.Stock{Name: s.Name}).FirstOrCreate(&s).Error
		if err != nil {
			return err
		}
	}
	return nil
}
