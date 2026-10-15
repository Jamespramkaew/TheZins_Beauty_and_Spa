package database

import (
	"strings"

	"backend/internal/models"

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
