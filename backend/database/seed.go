package database

import (
	"backend/internal/models"

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
