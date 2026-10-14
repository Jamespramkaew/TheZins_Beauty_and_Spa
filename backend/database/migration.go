package database

import (
	"log"

	"backend/internal/models"
	"gorm.io/gorm"
)

// allModels = ทุกตารางที่ต้องสร้าง/ลบ: 21 ตารางของร้าน + Example (ตัวอย่างที่ internal/example และ cmd/seed ยังใช้อยู่)
func allModels() []any {
	return append(models.All(), &Example{})
}

func RunMigration(db *gorm.DB) error {
	log.Println("Running database migration..")

	err := db.AutoMigrate(allModels()...)
	if err != nil {
		return err
	}

	// SQL เสริมที่ AutoMigrate ทำไม่ได้ (กันจองห้องซ้อน, CHECK ของ order_items, index สต็อก) รันซ้ำได้
	err = db.Exec(models.ExtraSQL).Error
	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func DropAllTables(db *gorm.DB) error {
	log.Println("Dropping all tables...")

	err := db.Migrator().DropTable(allModels()...)
	if err != nil {
		return err
	}

	log.Println("All tables dropped successfully")
	return nil
}

func ResetDatabase(db *gorm.DB) error {
	if err := DropAllTables(db); err != nil {
		return err
	}
	return RunMigration(db)
}
